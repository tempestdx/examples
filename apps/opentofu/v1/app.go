package opentofu

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os/exec"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/tempestdx/examples/deps/opentofu"
	"github.com/tempestdx/sdk-go/app"
)

const consoleURLTemplate = "https://%s.console.aws.amazon.com/s3/buckets/%s"

//go:embed module/*.tf
var moduleFS embed.FS

// setupTofu is a helper function that creates a new OpenTofu runner.
// It first pulls the 'SECRET_KEY' and 'ACCESS_KEY' from the Tempest Environment Variables.
// It then creates a new OpenTofu runner with the path to the tofu binary, the module filesystem, and the environment variables.
func setupTofu(env map[string]app.EnvironmentVariable) (*opentofu.Runner, error) {
	secretKey, ok := env["SECRET_KEY"]
	if !ok {
		return nil, fmt.Errorf("secret_key not found in environment")
	}

	accessKey, ok := env["ACCESS_KEY"]
	if !ok {
		return nil, fmt.Errorf("access_key not found in environment")
	}

	e := map[string]string{
		"AWS_ACCESS_KEY_ID":     accessKey.Value,
		"AWS_SECRET_ACCESS_KEY": secretKey.Value,
	}

	tfPath, err := exec.LookPath("tofu")
	if err != nil {
		log.Fatalf("Failed to find tofu binary: %s", err)
	}

	module, err := fs.Sub(moduleFS, "module")
	if err != nil {
		return nil, fmt.Errorf("failed to get module sub filesystem: %w", err)
	}

	tofu, err := opentofu.New(tfPath, module, e)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenTofu runner: %w", err)
	}

	return tofu, nil
}

func createFn(ctx context.Context, req *app.OperationRequest) (*app.OperationResponse, error) {
	// Create a new OpenTofu runner for this operation.
	tofu, err := setupTofu(req.Environment)
	if err != nil {
		return nil, err
	}

	// Run the `opentofu apply` command with the input map used as variables.
	state, err := tofu.Apply(req.Input)
	if err != nil {
		return nil, fmt.Errorf("failed to apply OpenTofu module: %w", err)
	}

	// Create a new properties map with the properties from the state.
	// These properties should match with the `properties.json` schema.
	// This object will be returned to the Tempest API and the properties will be displayed in the UI.
	properties := make(map[string]any)
	for _, r := range state.Values.RootModule.Resources {
		switch {
		case r.Type == "aws_s3_bucket" && r.Values["bucket"] == req.Input["name"]:
			properties["arn"] = r.Values["arn"]
			properties["bucket"] = r.Values["bucket"]
			properties["region"] = r.Values["region"]
		case r.Type == "aws_s3_bucket_versioning" && r.Values["bucket"] == req.Input["name"]:
			vconfs := r.Values["versioning_configuration"].([]any)
			if len(vconfs) == 1 {
				properties["versioning"] = vconfs[0].(map[string]any)["status"]
			}
		}
	}

	// Construct the OperationResponse which includes the Resource that was operated on.
	return &app.OperationResponse{
		Resource: &app.Resource{
			// ExternalID is the unique identifier for the resource in the external system.
			// This is used to identify the resource in subsequent operations.
			ExternalID: properties["arn"].(string),
			// DisplayName is the name of the resource that will be displayed in the UI.
			DisplayName: properties["bucket"].(string),
			// Properties is a map of key-value pairs that represent the properties of the resource.
			Properties: properties,
			// Links is a list of links that will be displayed in the UI.
			// There are different types of links that can be displayed:
			// - External: The link to view the resource in the external system.
			// - Support: The link to the support page for the individual resource.
			// - Endpoint: The link to the endpoint of the resource. For example, the connection string for a database.
			// - Documentation: The link to the documentation for the resource.
			// - Administration: The link to the administration page for the resource.
			Links: []*app.Link{
				{
					URL:   fmt.Sprintf(consoleURLTemplate, properties["region"], properties["bucket"]),
					Title: "AWS Console",
					Type:  app.LinkTypeExternal,
				},
			},
		},
	}, nil
}

func updateFn(ctx context.Context, req *app.OperationRequest) (*app.OperationResponse, error) {
	// The ExternalID is the unique identifier for the resource in the external system.
	// In this case, it's an ARN, but we only need the Resource part.
	a, err := arn.Parse(req.Resource.ExternalID)
	if err != nil {
		return nil, err
	}

	// Create a new OpenTofu runner for this operation.
	tofu, err := setupTofu(req.Environment)
	if err != nil {
		return nil, err
	}

	// First, import the existing resource into the fresh state.
	err = tofu.Import(req.Input, map[string]string{
		"aws_s3_bucket.bucket":                a.Resource,
		"aws_s3_bucket_versioning.versioning": a.Resource,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to import OpenTofu module: %w", err)
	}

	// Run the `opentofu apply` command with the new input.
	state, err := tofu.Apply(req.Input)
	if err != nil {
		return nil, fmt.Errorf("failed to apply OpenTofu module: %w", err)
	}

	// Create a new properties map with the properties from the state.
	// These properties should match with the `properties.json` schema.
	// This object will be returned to the Tempest API and the properties will be displayed in the UI.
	properties := make(map[string]any)
	for _, r := range state.Values.RootModule.Resources {
		switch {
		case r.Type == "aws_s3_bucket" && r.Values["bucket"] == req.Input["name"]:
			properties["arn"] = r.Values["arn"]
			properties["bucket"] = r.Values["bucket"]
			properties["region"] = r.Values["region"]
		case r.Type == "aws_s3_bucket_versioning" && r.Values["bucket"] == req.Input["name"]:
			vconfs := r.Values["versioning_configuration"].([]any)
			if len(vconfs) == 1 {
				properties["versioning"] = vconfs[0].(map[string]any)["status"]
			}
		}
	}

	return &app.OperationResponse{
		Resource: &app.Resource{
			// In this case the ExternalID did not change so it is omitted.
			// The DisplayName also did not change, so it is omitted.
			// The Properties are updated with the new values.
			Properties: properties,
		},
	}, nil
}

// readFn is called when Tempest needs to sync the state of the resource with the external system.
func readFn(ctx context.Context, req *app.OperationRequest) (*app.OperationResponse, error) {
	a, err := arn.Parse(req.Resource.ExternalID)
	if err != nil {
		return nil, err
	}

	// Create a new OpenTofu runner for this operation.
	tofu, err := setupTofu(req.Environment)
	if err != nil {
		return nil, err
	}

	// First, import the existing resource into the fresh state.
	err = tofu.Import(req.Input, map[string]string{
		"aws_s3_bucket.bucket":                a.Resource,
		"aws_s3_bucket_versioning.versioning": a.Resource,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to import OpenTofu module: %w", err)
	}

	state, err := tofu.Show()
	if err != nil {
		return nil, fmt.Errorf("failed to show OpenTofu state: %w", err)
	}

	properties := make(map[string]any)
	for _, r := range state.Values.RootModule.Resources {
		switch {
		case r.Type == "aws_s3_bucket" && r.Values["bucket"] == req.Resource.ExternalID:
			properties["arn"] = r.Values["arn"]
			properties["bucket"] = r.Values["bucket"]
			properties["region"] = r.Values["region"]
		case r.Type == "aws_s3_bucket_versioning" && r.Values["bucket"] == req.Resource.ExternalID:
			vconfs := r.Values["versioning_configuration"].([]any)
			if len(vconfs) == 1 {
				properties["versioning"] = vconfs[0].(map[string]any)["status"]
			}
		}
	}

	if len(properties) == 0 {
		return nil, fmt.Errorf("resource not found")
	}

	return &app.OperationResponse{
		Resource: &app.Resource{
			DisplayName: properties["bucket"].(string),
			Properties:  properties,
			Links: []*app.Link{
				{
					URL:   fmt.Sprintf(consoleURLTemplate, properties["region"], properties["bucket"]),
					Title: "AWS Console",
					Type:  app.LinkTypeExternal,
				},
			},
		},
	}, nil
}

func deleteFn(ctx context.Context, req *app.OperationRequest) (*app.OperationResponse, error) {
	a, err := arn.Parse(req.Resource.ExternalID)
	if err != nil {
		return nil, err
	}

	tofu, err := setupTofu(req.Environment)
	if err != nil {
		return nil, err
	}

	err = tofu.Import(nil, map[string]string{
		"aws_s3_bucket.bucket":                a.Resource,
		"aws_s3_bucket_versioning.versioning": a.Resource,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to import OpenTofu module: %w", err)
	}

	err = tofu.Destroy()
	if err != nil {
		return nil, err
	}

	// Construct the OperationResponse which includes the Resource that was deleted.
	// The Delete Operation should return the ExternalID of the resource that was deleted.
	return &app.OperationResponse{
		Resource: &app.Resource{
			ExternalID: req.Resource.ExternalID,
		},
	}, nil
}

var (
	//go:embed schema/properties.json
	propertiesSchema []byte

	//go:embed schema/apply.json
	applySchema []byte

	//go:embed instructions.md
	instructions string
)

// App is the main function that is called by the Tempest SDK and returns the app object.
// The name of the function must be 'App' and it must return an *app.App object.
func App() *app.App {
	// The ResourceDefinition is the main object that defines the resource as part of the app.
	// An App can have multiple ResourceDefinitions configured, but the 'Type' must be unique.
	bucketDef := app.ResourceDefinition{
		// Type is the internal identifier for the resource type, and must be unique within the app.
		Type: "s3_bucket",
		// Description is a short description of the resource type that will be displayed in the Tempest UI.
		Description: "An example resource that creates an AWS S3 Bucket by executing an OpenTofu module.",
		// DisplayName is the name of the resource type that will be displayed in the Tempest UI.
		DisplayName: "S3 Bucket",
		// The LifecycleStage is the stage of the resource in the Developer Lifecycle.
		// The options, in order, are:
		// LifecycleStageCode
		// LifecycleStageBuild
		// LifecycleStageTest
		// LifecycleStageRelease
		// LifecycleStageDeploy
		// LifecycleStageOperate
		// LifecycleStageMonitor
		LifecycleStage: app.LifecycleStageOperate,
		// The PropertiesSchema is the JSON schema that defines the properties of the individual resource.
		// After every Create, Update, Read, or List operation, the properties returned by the function
		// will be validated against this schema.
		// The properties in this schema can be used as input to other recipe steps in the UI.
		PropertiesSchema: app.MustParseJSONSchema(propertiesSchema),
		// Links are a list of links that will be displayed in the UI for the resource type.
		// These links should be relevant to the resource type and provide additional information or actions.
		// Individual resources can also have links that are specific to that resource.
		Links: []app.Link{
			{
				URL:   "https://docs.aws.amazon.com/AmazonS3/latest/userguide/Welcome.html",
				Title: "AWS S3 Documentation",
				Type:  app.LinkTypeDocumentation,
			},
		},
		// InstructionsMarkdown is the markdown content that will be displayed in the UI for the resource type.
		// This can include instructions on how to use the resource, best practices, or other helpful information.
		InstructionsMarkdown: instructions,
	}

	// Assign a function for each of the CRUD operations, starting with Create and Update.
	// Create and Update also include a schema that defines the shape of the input.
	// The Tempest UI will display these inputs to the user when creating or updating a resource as part of a Recipe.
	// That input will be validated against the schema before the function is called.

	// Create is called when Tempest is instructed to create a new resource in the external system.
	bucketDef.CreateFn(
		createFn,
		app.MustParseJSONSchema(applySchema),
	)

	// Update is called when Tempest is instructed to update the resource.
	// This can happen as part of a new Deployment in the Tempest Project.
	bucketDef.UpdateFn(
		updateFn,
		app.MustParseJSONSchema(applySchema),
	)

	// Read and Delete functions are also assigned to the resource definition.
	// There is no input schema for these functions, as they do not take input from the user.
	// Environment Variables and Secrets can still be accessed from these functions, however.
	// Read is called when Tempest needs to sync the state of the resource with the external system.
	bucketDef.ReadFn(readFn)
	// Delete is called when Tempest is instructed to delete the resource from the external system.
	bucketDef.DeleteFn(deleteFn)

	// Add a healthcheck function to the resource definition.
	// This will allow Tempest to monitor and display the health of the resource provider.
	// The function in this case is fairly simple, as it just returns a healthy status.
	// More complex health checks can be implemented as needed, such as checking the status of the external system.
	bucketDef.HealthCheckFn(func(ctx context.Context) (*app.HealthCheckResponse, error) {
		return &app.HealthCheckResponse{
			Status: app.HealthCheckStatusHealthy,
		}, nil
	})

	// Finally, add the resource definition to the app and return the configured App object.
	return app.New(
		app.WithResourceDefinition(bucketDef),
	)
}
