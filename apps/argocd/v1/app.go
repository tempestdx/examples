/*
Package appargocd implements a Tempest Private App for managing ArgoCD Applications.

This is an example of how to build a Tempest Private App that:
1. Defines a custom resource type ("application")
2. Implements CRUD operations for managing ArgoCD Applications in Kubernetes
3. Uses JSON schemas for input validation and property definitions
4. Leverages Go templates to generate Kubernetes manifests
5. Integrates with the Kubernetes API using dynamic clients
6. Handles environment-based configuration and secrets

For more information about Tempest Private Apps, see:
https://docs.tempestdx.com/developer/private-apps/overview
*/
package appargocd

import (
	"bytes"
	"context"
	"embed"
	"encoding/base64"
	"errors"
	"fmt"
	"hash/fnv"
	"io/fs"
	"os"
	"strings"
	"text/template"
	"time"

	argocd "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	"github.com/cenkalti/backoff/v4"
	"github.com/tempestdx/sdk-go/app"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	// Embed the JSON schema file for resource properties at compile time
	// This schema defines what properties are exposed for the "application" resource
	// in the Tempest catalog. See: https://docs.tempestdx.com/software-catalog/resources
	//go:embed schema/properties.json
	propertiesSchema []byte

	// ResourceDefinition tells Tempest:
	// - What this resource is called ("application")
	// - How to display it in the UI ("Application")
	// - What lifecycle stage it belongs to (Deploy)
	// - What properties it exposes (via JSON schema)
	application = app.ResourceDefinition{
		Type:             "application",                                            // Unique identifier for this resource type
		DisplayName:      "Application",                                            // Human-readable name shown in Tempest UI
		Description:      "Manages an ArgoCD Application in a Kubernetes cluster.", // Description for users
		LifecycleStage:   app.LifecycleStageDeploy,                                 // This is a deployment-stage resource
		PropertiesSchema: app.MustParseJSONSchema(propertiesSchema),                // Schema for resource properties
	}

	// Embed JSON schemas for create and update operations
	// These schemas validate the input when users create or update applications
	//go:embed schema/create.json
	createSchema []byte

	//go:embed schema/update.json
	updateSchema []byte

	// Embed the templates directory containing Kubernetes manifest templates
	// Templates are processed with Go's text/template package to generate actual manifests
	//go:embed templates
	templatesFS embed.FS
)

// ApplicationTemplateInput defines the data structure passed to Go templates
// when generating ArgoCD Application manifests. This struct maps the user input
// from Tempest to the template variables used in application.yaml.tmpl
type ApplicationTemplateInput struct {
	Name           string // Name of the ArgoCD Application
	Namespace      string // Target namespace for deployed resources
	SourcePath     string // Path within the Git repository
	RepoURL        string // Git repository URL
	Image          string // Container image to deploy
	TargetRevision string // Git branch/tag/commit to deploy
}

// secretTemplateInput defines the data structure for generating ArgoCD repository secrets
// ArgoCD needs authentication credentials to access private Git repositories
type secretTemplateInput struct {
	GitHubAppID          string // GitHub App ID for authentication
	GitHubInstallationID string // GitHub App Installation ID
	DeployKey            string // SSH private key for Git access
	Project              string // ArgoCD project name
	Name                 string // Application name
	RepoURL              string // Git repository URL
	SecretName           string // Kubernetes secret name
	Type                 string // Repository type (git)
}

// getConfigFromEnv creates a Kubernetes client configuration from environment variables
// Tempest Private Apps receive configuration through environment variables passed
// from the Tempest platform. This is how the app connects to the target Kubernetes cluster.
func getConfigFromEnv(env map[string]app.EnvironmentVariable) (*rest.Config, error) {
	kubeconfig, ok := env["KUBECONFIG"]
	if !ok || kubeconfig.Value == "" {
		return nil, errors.New("KUBECONFIG not found in environment")
	}

	// Build Kubernetes client config from the kubeconfig file path
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to build kubeconfig: %w", err)
	}

	return config, nil
}

// getDeployKeyFileFromEnv reads an SSH private key from a file specified in environment variables
// This key is used by ArgoCD to authenticate with private Git repositories
func getDeployKeyFileFromEnv(env map[string]app.EnvironmentVariable) (string, error) {
	deployKeyFile, ok := env["DEPLOY_KEY_FILE"]
	if !ok || deployKeyFile.Value == "" {
		return "", errors.New("DEPLOY_KEY_FILE not found in environment")
	}
	content, err := os.ReadFile(deployKeyFile.Value)
	if err != nil {
		return "", fmt.Errorf("failed to read SSH private key file: %w", err)
	}

	return strings.TrimSpace(string(content)), nil
}

// getGitHubAppIDFromEnv retrieves the GitHub App ID from environment variables
// Used for GitHub App-based authentication with ArgoCD
func getGitHubAppIDFromEnv(env map[string]app.EnvironmentVariable) (string, error) {
	githubAppID, ok := env["GITHUB_APP_ID"]
	if !ok || githubAppID.Value == "" {
		return "", errors.New("GITHUB_APP_ID not found in environment")
	}

	return githubAppID.Value, nil
}

// getGitHubInstallationIDFromEnv retrieves the GitHub App Installation ID from environment variables
// This ID specifies which GitHub organization/repository the app is installed for
func getGitHubInstallationIDFromEnv(env map[string]app.EnvironmentVariable) (string, error) {
	githubInstallationID, ok := env["GITHUB_INSTALLATION_ID"]
	if !ok || githubInstallationID.Value == "" {
		return "", errors.New("GITHUB_INSTALLATION_ID not found in environment")
	}

	return githubInstallationID.Value, nil
}

// createFn implements the CREATE operation for the application resource type
// This function is called when users create a new ArgoCD Application through Tempest
// It demonstrates the core pattern of Tempest Private Apps:
// 1. Extract configuration from environment variables
// 2. Validate and process user input
// 3. Generate Kubernetes manifests from templates
// 4. Apply manifests to the cluster using Kubernetes API
// 5. Return resource metadata to Tempest
func createFn(ctx context.Context, req *app.OperationRequest) (*app.OperationResponse, error) {
	// Step 1: Get Kubernetes configuration from environment
	// The OperationRequest contains environment variables provided by Tempest
	config, err := getConfigFromEnv(req.Environment)
	if err != nil {
		return nil, err
	}

	// Create a dynamic Kubernetes client for applying manifests
	// Dynamic clients can work with any Kubernetes resource type
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	// Step 2: Extract GitHub authentication credentials from environment
	githubAppID, err := getGitHubAppIDFromEnv(req.Environment)
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub app ID: %w", err)
	}

	githubInstallationID, err := getGitHubInstallationIDFromEnv(req.Environment)
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub installation ID: %w", err)
	}

	// Step 3: Load and parse Go templates from embedded filesystem
	// Templates are embedded at compile time for easy distribution
	templates, err := fs.Sub(templatesFS, "templates")
	if err != nil {
		return nil, err
	}

	tmpl, err := template.ParseFS(templates, "application.yaml.tmpl")
	if err != nil {
		return nil, err
	}

	// Step 4: Prepare template input from user-provided data
	// req.Input contains the validated user input matching create.json schema
	applicationInput := ApplicationTemplateInput{
		Name:           req.Input["name"].(string),
		Namespace:      req.Input["namespace"].(string),
		SourcePath:     req.Input["source_path"].(string),
		RepoURL:        req.Input["repo_url"].(string),
		Image:          req.Input["image"].(string),
		TargetRevision: req.Input["target_revision"].(string),
	}

	// Step 5: Execute template to generate ArgoCD Application manifest
	var applicationManifest bytes.Buffer
	err = tmpl.Execute(&applicationManifest, applicationInput)
	if err != nil {
		return nil, err
	}

	// Step 6: Prepare ArgoCD repository secret for Git authentication
	deployKey, err := getDeployKeyFileFromEnv(req.Environment)
	if err != nil {
		return nil, err
	}

	// Generate a deterministic secret name based on repository URL
	// This ensures one secret per repository to avoid ArgoCD conflicts
	h := fnv.New32a()
	_, _ = h.Write([]byte(req.Input["repo_url"].(string)))
	secretName := fmt.Sprintf("repo-%v", h.Sum32())

	tmpl, err = template.ParseFS(templates, "argocd_secret.yaml.tmpl")
	if err != nil {
		return nil, err
	}

	// ArgoCD secrets store authentication data as base64-encoded values
	secretInput := secretTemplateInput{
		DeployKey:            toBase64(deployKey),
		GitHubAppID:          toBase64(githubAppID),
		GitHubInstallationID: toBase64(githubInstallationID),
		Project:              toBase64("default"),
		Name:                 toBase64(req.Input["name"].(string)),
		RepoURL:              toBase64(req.Input["repo_url"].(string)),
		SecretName:           secretName,
		Type:                 toBase64("git"),
	}

	var secretManifest bytes.Buffer
	err = tmpl.Execute(&secretManifest, secretInput)
	if err != nil {
		return nil, err
	}

	// Step 7: Apply the secret first (ArgoCD needs repository access)
	uid, err := apply(ctx, dynamicClient, secretManifest.Bytes())
	if err != nil {
		return nil, err
	}

	if uid == "" {
		return nil, fmt.Errorf("failed to apply secret manifest")
	}

	// Step 8: Apply the ArgoCD Application manifest
	uid, err = apply(ctx, dynamicClient, applicationManifest.Bytes())
	if err != nil {
		return nil, err
	}

	if uid == "" {
		return nil, fmt.Errorf("failed to apply application manifest")
	}

	// Step 9: Return resource metadata to Tempest
	// The OperationResponse tells Tempest about the created resource:
	// - ExternalID: Unique identifier for this resource instance
	// - DisplayName: Human-readable name for the Tempest UI
	// - Properties: Key-value pairs exposed in the software catalog
	return &app.OperationResponse{
		Resource: &app.Resource{
			ExternalID:  strings.Join([]string{applicationInput.Namespace, applicationInput.Name, uid}, "/"),
			DisplayName: applicationInput.Name,
			Properties: map[string]any{
				"name":        applicationInput.Name,
				"namespace":   applicationInput.Namespace,
				"repo_url":    applicationInput.RepoURL,
				"source_path": applicationInput.SourcePath,
				"image":       applicationInput.Image,
				"cluster":     config.Host,
			},
		},
	}, nil
}

// updateFn implements the UPDATE operation for the application resource type
// This function is called when users modify an existing ArgoCD Application through Tempest
// It follows a similar pattern to createFn but updates an existing resource
func updateFn(ctx context.Context, req *app.OperationRequest) (*app.OperationResponse, error) {
	// Get Kubernetes configuration and client
	config, err := getConfigFromEnv(req.Environment)
	if err != nil {
		return nil, err
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	// Parse the ExternalID to extract namespace, name, and UID
	// ExternalID format: "namespace/name/uid"
	id := strings.Split(req.Resource.ExternalID, "/")
	if len(id) != 3 {
		return nil, errors.New("invalid external ID")
	}

	// Load and prepare templates
	templates, err := fs.Sub(templatesFS, "templates")
	if err != nil {
		return nil, err
	}

	tmpl, err := template.ParseFS(templates, "application.yaml.tmpl")
	if err != nil {
		return nil, err
	}

	// Prepare template input using existing resource metadata and new input
	// For updates, we preserve the namespace and name from the existing resource
	in := ApplicationTemplateInput{
		Namespace:      id[0], // Preserve original namespace
		Name:           id[1], // Preserve original name
		SourcePath:     req.Input["source_path"].(string),
		RepoURL:        req.Input["repo_url"].(string),
		Image:          req.Input["image"].(string),
		TargetRevision: req.Input["target_revision"].(string),
	}

	// Generate and apply the updated manifest
	var manifest bytes.Buffer
	err = tmpl.Execute(&manifest, in)
	if err != nil {
		return nil, err
	}

	uid, err := apply(ctx, dynamicClient, manifest.Bytes())
	if err != nil {
		return nil, err
	}

	if uid == "" {
		return nil, fmt.Errorf("failed to update application manifest")
	}

	// Return updated resource metadata
	return &app.OperationResponse{
		Resource: &app.Resource{
			ExternalID:  req.Resource.ExternalID, // Keep the same ExternalID
			DisplayName: in.Name,
			Properties: map[string]any{
				"name":        in.Name,
				"namespace":   in.Namespace,
				"repo_url":    in.RepoURL,
				"source_path": in.SourcePath,
				"image":       in.Image,
				"cluster":     config.Host,
			},
		},
	}, nil
}

// readFn implements the READ operation for the application resource type
// This function is called when Tempest needs to fetch current state of a resource
// It demonstrates how to query Kubernetes resources and extract relevant data
func readFn(ctx context.Context, req *app.OperationRequest) (*app.OperationResponse, error) {
	// Get Kubernetes configuration and client
	config, err := getConfigFromEnv(req.Environment)
	if err != nil {
		return nil, err
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	// Parse ExternalID to get resource identifiers
	id := strings.Split(req.Resource.ExternalID, "/")
	if len(id) != 3 {
		return nil, errors.New("invalid external ID")
	}

	// Define the GroupVersionResource for ArgoCD Applications
	// This tells Kubernetes API which resource type we want to query
	gvr := schema.GroupVersionResource{
		Group:    "argoproj.io",  // ArgoCD API group
		Version:  "v1alpha1",     // ArgoCD API version
		Resource: "applications", // Resource type plural name
	}

	// Fetch the ArgoCD Application from Kubernetes
	// ArgoCD Applications are typically deployed in the "argocd" namespace
	obj, err := dynamicClient.Resource(gvr).Namespace("argocd").Get(ctx, id[1], metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// Convert the unstructured object to a typed ArgoCD Application
	// This allows us to access ArgoCD-specific fields safely
	scheme := runtime.NewScheme()
	err = argocd.AddToScheme(scheme)
	if err != nil {
		return nil, err
	}

	data, err := obj.MarshalJSON()
	if err != nil {
		return nil, err
	}

	var application argocd.Application
	_, _, err = serializer.NewCodecFactory(scheme).UniversalDeserializer().Decode(data, nil, &application)
	if err != nil {
		return nil, err
	}

	// Extract the current image from the ArgoCD Application spec
	// ArgoCD stores Kustomize image overrides in the source configuration
	var image string
	if len(application.Spec.Source.Kustomize.Images) > 0 {
		image = string(application.Spec.Source.Kustomize.Images[0])
	}

	// Return current resource state to Tempest
	return &app.OperationResponse{
		Resource: &app.Resource{
			ExternalID:  req.Resource.ExternalID,
			DisplayName: application.GetName(),
			Properties: map[string]any{
				"name":        application.GetName(),
				"namespace":   application.GetNamespace(),
				"repo_url":    application.Spec.Source.RepoURL,
				"source_path": application.Spec.Source.Path,
				"image":       image,
				"cluster":     config.Host,
			},
		},
	}, nil
}

// apply is a helper function that applies Kubernetes manifests to the cluster
// It handles both ArgoCD Applications and Kubernetes Secrets with appropriate logic
// This function demonstrates the "apply" pattern used by kubectl and other tools
func apply(ctx context.Context, dc *dynamic.DynamicClient, manifest []byte) (string, error) {
	// Parse the YAML manifest into an unstructured Kubernetes object
	decoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	obj := &unstructured.Unstructured{}
	_, _, err := decoder.Decode(manifest, nil, obj)
	if err != nil {
		return "", fmt.Errorf("failed to decode manifest: %w", err)
	}

	var res *unstructured.Unstructured

	// Handle different resource types with specific logic
	switch obj.GetKind() {
	case "Application":
		// Define GroupVersionResource for ArgoCD Applications
		gvr := schema.GroupVersionResource{
			Group:    "argoproj.io",
			Version:  "v1alpha1",
			Resource: "applications",
		}

		// Apply the Application manifest using server-side apply
		// FieldManager identifies this app as the owner of applied fields
		res, err = dc.Resource(gvr).Namespace(obj.GetNamespace()).Apply(ctx, obj.GetName(), obj, metav1.ApplyOptions{
			FieldManager: "tempest", // Important: This identifies our app as the field manager
		})
		if err != nil {
			return "", fmt.Errorf("failed to apply application manifest: %w", err)
		}

		// Wait for the ArgoCD Application to become healthy
		// This demonstrates how to wait for resources to reach desired state
		if err := backoff.Retry(func() error {
			var err error

			// Fetch the resource to check its status
			app, err := dc.Resource(gvr).Namespace(obj.GetNamespace()).Get(ctx, obj.GetName(), metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("failed to fetch application: %w", err)
			}

			// Extract status fields from the ArgoCD Application
			// ArgoCD Applications have sync status and health status
			status, found, _ := unstructured.NestedString(app.Object, "status", "sync", "status")
			health, foundHealth, _ := unstructured.NestedString(app.Object, "status", "health", "status")

			if !found || !foundHealth {
				return fmt.Errorf("status fields not found in application")
			}

			// Check if the Application has reached the desired state
			// "Synced" means ArgoCD has applied all manifests
			// "Healthy" means all resources are running correctly
			if status == "Synced" && health == "Healthy" {
				fmt.Println("Application is Synced and Healthy!")
				return nil
			}

			return fmt.Errorf("application is not yet Synced and Healthy (sync: %s, health: %s)", status, health)
		}, backoff.NewExponentialBackOff(
			backoff.WithMaxElapsedTime(10*time.Minute), // Wait up to 10 minutes
		)); err != nil {
			return "", err
		}
	case "Secret":
		// Define GroupVersionResource for Kubernetes Secrets
		gvr := schema.GroupVersionResource{
			Group:    "",        // Core Kubernetes API group (empty string)
			Version:  "v1",      // Kubernetes API version
			Resource: "secrets", // Resource type plural name
		}

		// Check if a secret with the same name already exists
		// ArgoCD does not support multiple secrets of the same type for a single repository
		// Although Kubernetes allows multiple secret objects, ArgoCD will only recognize the first one
		// To prevent conflicts, we ensure that only one secret object is maintained for each repository
		//
		// TODO: More advanced logic can be added here to detect changes and update the secret accordingly
		res, err = dc.Resource(gvr).Namespace(obj.GetNamespace()).Get(ctx, obj.GetName(), metav1.GetOptions{})

		// Check if an error occurred and it's not a "not found" error
		if err != nil {
			if !k8serrors.IsNotFound(err) {
				// Return an error if it's not a "not found" error
				return "", fmt.Errorf("failed to check if secret exists: %w", err)
			}
			// If it's a "not found" error, continue to apply the secret
		} else {
			// If no error occurred, check if the response is valid
			if res != nil {
				// Secret already exists, return its UID
				return string(res.GetUID()), nil
			}
		}

		// Continue to apply the secret if it was not found
		res, err = dc.Resource(gvr).Namespace(obj.GetNamespace()).Apply(ctx, obj.GetName(), obj, metav1.ApplyOptions{
			FieldManager: "tempest", // Field manager for server-side apply
		})
		if err != nil {
			return "", fmt.Errorf("failed to apply secret manifest: %w", err)
		}
	}

	if res == nil {
		return "", fmt.Errorf("failed to apply resource")
	}

	// Return the UID of the applied resource
	// UIDs are unique identifiers assigned by Kubernetes
	return string(res.GetUID()), nil
}

// toBase64 is a helper function to encode strings as base64
// ArgoCD secrets require base64-encoded values for authentication data
func toBase64(input string) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}

// App returns the configured Tempest Private App instance
// This is the main entry point that connects all the pieces together:
// 1. Resource definition with schemas
// 2. CRUD operation handlers
// 3. Health check implementation
//
// The returned app.App instance is what gets registered with Tempest
func App() *app.App {
	// Configure the CREATE operation with input validation schema
	// When users create applications through Tempest, their input will be validated
	// against the create.json schema before calling createFn
	application.CreateFn(
		createFn,
		app.MustParseJSONSchema(createSchema),
	)

	// Configure the UPDATE operation with input validation schema
	// Similar to create, but uses update.json schema for validation
	application.UpdateFn(
		updateFn,
		app.MustParseJSONSchema(updateSchema),
	)

	// Configure the READ operation (no input schema needed)
	// This allows Tempest to fetch current resource state
	application.ReadFn(readFn)

	// Configure a health check for this Tempest Private App
	// Tempest calls this periodically to ensure the app is functioning
	// Health checks help with monitoring and troubleshooting
	application.HealthCheckFn(func(ctx context.Context) (*app.HealthCheckResponse, error) {
		return &app.HealthCheckResponse{
			Status:  app.HealthCheckStatusHealthy,
			Message: "ArgoCD Application is healthy",
		}, nil
	})

	// Create and return the Tempest Private App instance
	// This app can manage "application" resources with full CRUD operations
	return app.New(
		app.WithResourceDefinition(application),
	)
}
