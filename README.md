# Tempest Private App Examples

## OpenTofu

The OpenTofu example shows how to execute a simple OpenTofu configuration, and interact with the resources in Tempest.

### Usage

1. Navigate to the directory where the OpenTofu module is found. In the example code this would be `apps/opentofu/v1/module`.
2. Run `OPENTOFU_WORKDIR=$(pwd) tempest app <command> opentofu:v1`

### Example

```shell
$ cd apps/opentofu/v1/module
$ OPENTOFU_WORKDIR=$(pwd) tempest app describe opentofu:v1

Tempest App Description
-----------------------

Describing app: opentofu:v1
Location: tempestdx/examples/apps/opentofu/v1/module/apps/opentofu/v1

Resource Type: S3 Bucket

Operations Supported:
- ✅ Read
- ✅ List
- ✅ Create
- ✅ Update
- ✅ Delete

Health Check Supported: ✅

$ OPENTOFU_WORKDIR=$(pwd) tempest app test opentofu:v1 --operation create --type s3_bucket --input '{"name":"my-test-bucket"}'

Resource created with ID:  arn:aws:s3:::my-test-bucket
Properties:
{
  "arn": "arn:aws:s3:::my-test-bucket",
  "bucket": "my-test-bucket",
  "region": "us-east-1",
  "versioning": "Suspended"
}
```

## Dashboards

The Dashboards example implements a Private App that interacts with the fictional "Dashboards" server found in `deps/`. This example showcases how Tempest can orchestrate your custom in-house servers via Private Apps.

### Usage

1. In one terminal, run the dashboards server. `go run deps/dashboards/server/main.go`
2. In another terminal, navigate to the `tempestdx/examples` repository and use the app.

### Example

```shell
$ go run deps/dashboards/server/main.go
Server started at :8080

$ tempest app describe dashboards:v1
Tempest App Description
-----------------------

Describing app: dashboards:v1
Location: /Users/mike/git/tempestdx/examples/apps/dashboards/v1

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
