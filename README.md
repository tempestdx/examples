# Tempest Private App Examples

The official [Tempest][tempest] Private App examples repository.

## OpenTofu

The OpenTofu example shows how to execute a simple OpenTofu configuration, and
interact with the resources in Tempest.

### Usage

1. Run the app with the CLI from with the `examples/` directory.

### Example

```shell
$ tempest app describe opentofu:v1

Tempest App Description
-----------------------

Describing app: opentofu:v1
Location: tempestdx/examples/apps/opentofu/v1/module/apps/opentofu/v1

Resource Type: S3 Bucket

Operations Supported:
- ✅ Read
- ❌ List
- ✅ Create
- ✅ Update
- ✅ Delete

Health Check Supported: ✅

$ tempest app test opentofu:v1 --operation create --type s3_bucket --input '{"name":"my-test-bucket"}' --env SECRET_KEY=abcd123 --env ACCESS_KEY=12345

Resource created with ID:  arn:aws:s3:::my-test-bucket
Properties:
{
  "arn": "arn:aws:s3:::my-test-bucket",
  "bucket": "my-test-bucket",
  "region": "us-east-1",
  "versioning": "Suspended"
}
```

## ArgoCD

The ArgoCD example implements a Private App that manages ArgoCD Applications in
Kubernetes clusters. This example demonstrates how to integrate with GitOps
workflows by creating and managing ArgoCD Applications through Tempest.

### Usage

1. Ensure you have a Kubernetes cluster with ArgoCD installed and configured.
2. Set the required environment variables for Kubernetes access and GitHub
   authentication.
3. Run the app with the CLI from the `examples/` directory.

### Example

```shell
$ tempest app describe argocd:v1

Tempest App Description
-----------------------

Describing app: argocd:v1
Location: tempestdx/examples/apps/argocd/v1

Resource Type: Application

Operations Supported:
- ✅ Read
- ❌ List
- ✅ Create
- ✅ Update
- ❌ Delete

Health Check Supported: ✅

$ tempest app test argocd:v1 --operation create --type application --input '{"name":"my-app","source_path":"applications/example-app/kustomize/overlays/sandbox","image":"us-west2-docker.pkg.dev/tempestdx/example-repository/app:1.0.1"}' --env KUBECONFIG=/path/to/kubeconfig --env GITHUB_APP_ID=123456 --env GITHUB_INSTALLATION_ID=78910

Resource created with ID:  my-app/default
Properties:
{
  "name": "my-app",
  "namespace": "default",
  "repo_url": "https://github.com/tempestdx/example-repository.git",
  "source_path": "applications/example-app/kustomize/overlays/sandbox",
  "image": "us-west2-docker.pkg.dev/tempestdx/example-repository/app:1.0.1",
  "target_revision": "HEAD"
}

$ tempest app test argocd:v1 --operation update --type application -e my-app/default --input '{"source_path":"applications/example-app/kustomize/overlays/production","image":"us-west2-docker.pkg.dev/tempestdx/example-repository/app:1.0.2"}'
Resource updated with ID:  my-app/default
Properties:
{
  "name": "my-app",
  "namespace": "default",
  "repo_url": "https://github.com/tempestdx/example-repository.git",
  "source_path": "applications/example-app/kustomize/overlays/production",
  "image": "us-west2-docker.pkg.dev/tempestdx/example-repository/app:1.0.2",
  "target_revision": "HEAD"
}
```

## Dashboards

The Dashboards example implements a Private App that interacts with the
fictional "Dashboards" server found in `deps/`. This example showcases how
Tempest can orchestrate your custom in-house servers via Private Apps.

### Usage

1. In one terminal, run the dashboards server.
   `go run deps/dashboards/server/main.go`
2. In another terminal, navigate to the `tempestdx/examples` repository and use
   the app.

### Example

```shell
$ go run deps/dashboards/server/main.go
Server started at :8080

$ tempest app describe dashboards:v1
Tempest App Description
-----------------------

Describing app: dashboards:v1
Location: tempestdx/examples/apps/dashboards/v1

Resource Type: Dashboard

Operations Supported:
- ✅ Read
- ✅ List
- ✅ Create
- ✅ Update
- ✅ Delete

Health Check Supported: ✅

$ tempest app test dashboards:v1 --operation create --type dashboard --input '{"name": "my-example-dashboard"}'
Resource created with ID:  6Oh8ZeHr
Properties:
{
  "description": "",
  "id": "6Oh8ZeHr",
  "name": "my-example-dashboard",
  "project_id": "TEMPESTCLIFoWJFMwo"
}

$ tempest app test dashboards:v1 --operation delete --type dashboard -e 6Oh8ZeHr
Resource deleted with ID:  6Oh8ZeHr
```

## Support

To share any requests, bugs or comments, please [open an issue][issues] or
[submit a pull request][pulls].

[issues]: https://github.com/tempestdx/examples/issues/new
[pulls]: https://github.com/tempestdx/examples/pulls
[tempest]: https://tempestdx.com/
