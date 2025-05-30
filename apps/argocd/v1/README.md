# ArgoCD Tempest Private App

This directory contains a complete example of a **Tempest Private App** that
manages [ArgoCD Applications](https://argo-cd.readthedocs.io/en/stable/) in
Kubernetes clusters. This Private App demonstrates all the key concepts and
patterns for building Tempest Private Apps.

## 📖 What is a Tempest Private App?

A Tempest Private App is a custom integration that extends Tempest's
capabilities to manage any type of resource or service. Private Apps allow you
to:

- Define custom resource types for your software catalog
- Implement CRUD operations for managing those resources
- Provide self-service capabilities to developers
- Integrate with any API or service

For more information, see the
[Tempest Private Apps documentation](https://docs.tempestdx.com/developer/private-apps/overview).

## 🏗️ Architecture Overview

This ArgoCD Private App demonstrates the complete lifecycle of managing ArgoCD
Applications:

1. **Resource Definition**: Defines an "application" resource type in Tempest's
   software catalog
2. **Input Validation**: Uses JSON schemas to validate user input for
   create/update operations
3. **Template Processing**: Generates Kubernetes manifests using Go templates
4. **Kubernetes Integration**: Applies manifests using the Kubernetes dynamic
   client
5. **State Management**: Tracks resource state and provides read operations
6. **Health Monitoring**: Waits for ArgoCD applications to become healthy

## 📁 File Structure

```
apps/argocd/v1/
├── app.go                              # Main Private App implementation
├── README.md                           # This documentation file
├── schema/
│   ├── create.json                     # Input validation for create operations
│   ├── update.json                     # Input validation for update operations
│   └── properties.json                 # Resource properties schema
└── templates/
    ├── application.yaml.tmpl           # ArgoCD Application manifest template
    └── argocd_secret.yaml.tmpl         # Repository secret manifest template
```

## 🔧 Core Components

### 1. Resource Definition (`app.go`)

The `application` resource definition tells Tempest:

- **Type**: `"application"` - unique identifier for this resource type
- **Display Name**: `"Application"` - human-readable name in the UI
- **Lifecycle Stage**: `Deploy` - this is a deployment-stage resource
- **Properties Schema**: Defines what properties are exposed in the software
  catalog

### 2. JSON Schemas (`schema/`)

Tempest Private Apps use JSON Schema (Draft 7) for input validation and property
definitions:

#### `create.json` - Create Operation Schema

- Validates user input when creating new applications
- Defines required fields: `name`, `source_path`, `image`
- Provides default values and examples to guide users
- Maps to form fields in the Tempest UI

Key fields:

- `name`: Application name (required)
- `namespace`: Target Kubernetes namespace (default: "default")
- `repo_url`: Git repository URL (default: example repository)
- `source_path`: Path within the repository (required)
- `image`: Container image to deploy (required)
- `target_revision`: Git branch/tag/commit (default: "HEAD")

#### `update.json` - Update Operation Schema

- Similar to create schema but only allows updating certain fields
- Required fields: `source_path`, `image`
- Cannot change `name` or `namespace` after creation

#### `properties.json` - Resource Properties Schema

- Defines what properties are exposed in Tempest's software catalog
- Used for displaying resource information in the UI
- All fields are required for complete resource representation

### 3. Templates (`templates/`)

Templates use Go's `text/template` package to generate Kubernetes manifests:

#### `application.yaml.tmpl` - ArgoCD Application

Generates a complete ArgoCD Application manifest with:

- Metadata (name, namespace, finalizers)
- Source configuration (repository, path, target revision)
- Destination configuration (cluster, namespace)
- Sync policy (automated with pruning and self-healing)
- Kustomize image overrides

#### `argocd_secret.yaml.tmpl` - Repository Secret

Generates a Kubernetes Secret for ArgoCD repository authentication with:

- GitHub App credentials (App ID, Installation ID, Private Key)
- Repository metadata (URL, name, type)
- Proper ArgoCD secret annotations and labels

## 🔐 Environment Variables

The Private App expects these environment variables (provided by Tempest):

- `KUBECONFIG`: Path to Kubernetes configuration file
- `DEPLOY_KEY_FILE`: Path to SSH private key for Git access
- `GITHUB_APP_ID`: GitHub App ID for authentication
- `GITHUB_INSTALLATION_ID`: GitHub App Installation ID

## 🚀 How It Works

### Create Operation Flow

1. **Input Validation**: User input is validated against `create.json` schema
2. **Environment Setup**: Extract configuration from environment variables
3. **Template Processing**: Generate Kubernetes manifests from templates
4. **Secret Creation**: Apply repository secret for Git authentication
5. **Application Creation**: Apply ArgoCD Application manifest
6. **Health Check**: Wait for application to become "Synced" and "Healthy"
7. **Response**: Return resource metadata to Tempest

### Update Operation Flow

1. **Input Validation**: User input is validated against `update.json` schema
2. **Resource Identification**: Parse ExternalID to find existing resource
3. **Template Processing**: Generate updated manifest with new values
4. **Application Update**: Apply updated ArgoCD Application manifest
5. **Response**: Return updated resource metadata to Tempest

### Read Operation Flow

1. **Resource Identification**: Parse ExternalID to find resource
2. **Kubernetes Query**: Fetch current ArgoCD Application from cluster
3. **Data Extraction**: Extract relevant fields from the Application spec
4. **Response**: Return current resource state to Tempest

## 🧪 Testing and Development

To test this Private App locally:

1. **Prerequisites**:
   - Go 1.24+ installed
   - Access to Kubernetes cluster with ArgoCD installed
   - Valid kubeconfig file
   - GitHub App credentials

2. **Environment Setup**:
   ```bash
   export KUBECONFIG=/path/to/your/kubeconfig
   export DEPLOY_KEY_FILE=/path/to/ssh/private/key
   export GITHUB_APP_ID=your_github_app_id
   export GITHUB_INSTALLATION_ID=your_installation_id
   ```

3. **Build and Test**:
   ```bash
   cd apps/argocd/v1
   go mod tidy
   go build -o argocd-app .
   ./argocd-app
   ```

## 📚 Learning Resources

- [Tempest Private Apps Overview](https://docs.tempestdx.com/developer/private-apps/overview)
- [Tempest SDK for Go](https://docs.tempestdx.com/developer/sdk/overview)
- [ArgoCD Documentation](https://argo-cd.readthedocs.io/en/stable/)
- [JSON Schema Specification](https://json-schema.org/draft-07/schema)
- [Go Templates](https://pkg.go.dev/text/template)
- [Kubernetes Dynamic Client](https://pkg.go.dev/k8s.io/client-go/dynamic)

## 🤝 Contributing

This example serves as a reference implementation for building Tempest Private
Apps. Key concepts demonstrated here can be adapted for managing any type of
resource or service integration.

When building your own Private Apps, consider:

- Clear resource type definitions
- Comprehensive input validation
- Robust error handling
- Proper state management
- Health monitoring
- Documentation and examples
