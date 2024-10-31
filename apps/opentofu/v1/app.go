package opentofu

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/tempestdx/examples/deps/opentofu"
	"github.com/tempestdx/sdk-go/app"
)

const consoleURLTemplate = "https://%s.console.aws.amazon.com/s3/buckets/%s"

var (
	tfPath  string
	workDir string
)

func init() {
	var err error
	tfPath, err = exec.LookPath("tofu")
	if err != nil {
		log.Fatalf("Failed to find tofu binary: %s", err)
	}

	workDir = os.Getenv("OPENTOFU_WORKDIR")
	if workDir == "" {
		log.Fatal("OPENTOFU_WORKDIR is not set")
	}
}

func applyFn(ctx context.Context, req *app.OperationRequest) (*app.OperationResponse, error) {
	tofu, err := opentofu.New(tfPath, workDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenTofu runner: %w", err)
	}

	state, err := tofu.Apply(req.Input)
	if err != nil {
		return nil, fmt.Errorf("failed to apply OpenTofu module: %w", err)
	}

	properties := make(map[string]any)
	for _, r := range state.Values.RootModule.Resources {
		if r.Type == "aws_s3_bucket" && r.Values["bucket"] == req.Input["name"] {
			properties = r.Values
			break
		}
	}

	return &app.OperationResponse{
		Resource: &app.Resource{
			ExternalID:  properties["arn"].(string),
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

	tofu, err := opentofu.New(tfPath, workDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenTofu runner: %w", err)
	}

	err = tofu.Destroy(map[string]any{
		"name": a.Resource,
	})
	if err != nil {
		return nil, err
	}

	return &app.OperationResponse{
		Resource: &app.Resource{
			ExternalID: req.Resource.ExternalID,
		},
	}, nil
}

func readFn(ctx context.Context, req *app.OperationRequest) (*app.OperationResponse, error) {
	tofu, err := opentofu.New(tfPath, workDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenTofu runner: %w", err)
	}

	state, err := tofu.Show()
	if err != nil {
		return nil, fmt.Errorf("failed to show OpenTofu state: %w", err)
	}

	properties := make(map[string]any)
	for _, r := range state.Values.RootModule.Resources {
		if r.Type == "aws_s3_bucket" && r.Values["arn"] == req.Resource.ExternalID {
			properties = r.Values
			break
		}
	}

	return &app.OperationResponse{
		Resource: &app.Resource{
			ExternalID:  properties["arn"].(string),
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

func listFn(ctx context.Context, req *app.ListRequest) (*app.ListResponse, error) {
	tofu, err := opentofu.New(tfPath, workDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenTofu runner: %w", err)
	}

	res, err := tofu.List("aws_s3_bucket")
	if err != nil {
		return nil, fmt.Errorf("failed to list OpenTofu resources: %w", err)
	}

	resources := make([]*app.Resource, 0, len(res))
	for _, r := range res {
		resources = append(resources, &app.Resource{
			ExternalID:  r.Values["arn"].(string),
			DisplayName: r.Values["bucket"].(string),
			Properties:  r.Values,
			Links: []*app.Link{
				{
					URL:   fmt.Sprintf(consoleURLTemplate, r.Values["region"], r.Values["bucket"]),
					Title: "AWS Console",
					Type:  app.LinkTypeExternal,
				},
			},
		})
	}

	return &app.ListResponse{
		Resources: resources,
	}, nil
}

func healthFn(ctx context.Context) (*app.HealthCheckResponse, error) {
	return &app.HealthCheckResponse{
		Status: app.HealthCheckStatusHealthy,
	}, nil
}

var (
	//go:embed schema/properties.json
	propertiesSchema []byte

	//go:embed schema/apply.json
	applySchema []byte
)

func App() *app.App {
	resourceDefinition := app.ResourceDefinition{
		Type:             "s3_bucket",
		Description:      "An example resource that creates an AWS S3 Bucket by executing an OpenTofu module.",
		DisplayName:      "S3 Bucket",
		LifecycleStage:   app.LifecycleStageOperate,
		PropertiesSchema: app.MustParseJSONSchema(propertiesSchema),
	}

	// Add various operations to the resource definition.
	resourceDefinition.CreateFn(
		applyFn,
		app.MustParseJSONSchema(applySchema),
	)

	resourceDefinition.UpdateFn(
		applyFn,
		app.MustParseJSONSchema(applySchema),
	)

	resourceDefinition.ReadFn(readFn)
	resourceDefinition.DeleteFn(deleteFn)
	resourceDefinition.ListFn(listFn)

	// Add a healthcheck function to the resource definition.
	// This will allow Tempest to monitor and display the health of the resource provider.
	resourceDefinition.HealthCheckFn(healthFn)

	// Add the resource definition to the app.
	return app.New(
		app.WithResourceDefinition(resourceDefinition),
	)
}
