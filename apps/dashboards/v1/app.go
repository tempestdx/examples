package dashboards

import (
	"context"
	_ "embed"
	"strconv"

	"github.com/tempestdx/examples/deps/dashboards/server/client"
	"github.com/tempestdx/examples/deps/dashboards/server/models"
	"github.com/tempestdx/sdk-go/app"
)

var (
	//go:embed schema/properties.json
	propertiesSchema []byte

	resourceDefinition = app.ResourceDefinition{
		Type:             "dashboard",
		Description:      "A dashboard resource represents a Dashboard object in the Server.",
		DisplayName:      "Dashboard",
		LifecycleStage:   app.LifecycleStageMonitor,
		PropertiesSchema: app.MustParseJSONSchema(propertiesSchema),
	}

	baseURL = "http://localhost:8080"
)

func createFn(ctx context.Context, req *app.OperationRequest) (*app.OperationResponse, error) {
	client := client.NewClient(baseURL)

	dashboard := models.Dashboard{
		// Name is required in the JSON Schema, and will be validated by the SDK.
		// We can safely assume that the name is present in the input, and is a string.
		Name: req.Input["name"].(string),
		// Additionally, the Project field can be filled with the ProjectID from the Metadata.
		Project: req.Metadata.ProjectID,
	}

	// This is a double assertion (first for map existence, then for type), but Go handles this just fine.
	// Again, we can safely assume that the description, if present in the input, is a string.
	if description, ok := req.Input["description"].(string); ok {
		dashboard.Description = description
	}

	// Call the client to create the dashboard.
	res, err := client.CreateDashboard(ctx, dashboard)
	if err != nil {
		return nil, err
	}

	// resource is the SDK representation of the resource that was created,
	// and the properties will be stored in the Tempest server.
	// The properties returned will be validated against the ResourceDefinition's properties schema.
	resource := &app.Resource{
		ExternalID:  res.ID,
		DisplayName: res.Name,
		Properties: map[string]any{
			"id":          res.ID,
			"name":        res.Name,
			"description": res.Description,
			"project_id":  res.Project,
		},
	}

	return &app.OperationResponse{
		Resource: resource,
	}, nil
}

func updateFn(ctx context.Context, req *app.OperationRequest) (*app.OperationResponse, error) {
	client := client.NewClient(baseURL)

	dashboard := models.Dashboard{}

	// There are no required fields in the Update input schema, so we need to check if the fields are present.
	if description, ok := req.Input["description"].(string); ok {
		dashboard.Description = description
	}

	if name, ok := req.Input["name"].(string); ok {
		dashboard.Name = name
	}

	// Call the client to update the dashboard with the new model.
	res, err := client.UpdateDashboard(ctx, req.Resource.ExternalID, dashboard)
	if err != nil {
		return nil, err
	}

	// resource is the SDK representation of the resource that was updated,
	// and the properties will be stored in the Tempest server.
	// The properties returned will be validated against the ResourceDefinition's properties schema.
	resource := &app.Resource{
		ExternalID:  res.ID,
		DisplayName: res.Name,
		Properties: map[string]any{
			"id":          res.ID,
			"name":        res.Name,
			"description": res.Description,
			"project_id":  res.Project,
		},
	}

	return &app.OperationResponse{
		Resource: resource,
	}, nil
}

func deleteFn(ctx context.Context, req *app.OperationRequest) (*app.OperationResponse, error) {
	client := client.NewClient(baseURL)

	err := client.DeleteDashboard(ctx, req.Resource.ExternalID)
	if err != nil {
		return nil, err
	}

	return &app.OperationResponse{
		Resource: req.Resource,
	}, nil
}

func readFn(ctx context.Context, req *app.OperationRequest) (*app.OperationResponse, error) {
	client := client.NewClient(baseURL)

	res, err := client.GetDashboard(ctx, req.Resource.ExternalID)
	if err != nil {
		return nil, err
	}

	// resource is the SDK representation of the resource that was read,
	// and the properties will be stored in the Tempest server.
	// The properties returned will be validated against the ResourceDefinition's properties schema.
	resource := &app.Resource{
		ExternalID:  res.ID,
		DisplayName: res.Name,
		Properties: map[string]any{
			"id":          res.ID,
			"name":        res.Name,
			"description": res.Description,
			"project_id":  res.Project,
		},
	}

	return &app.OperationResponse{
		Resource: resource,
	}, nil
}

func listFn(ctx context.Context, req *app.ListRequest) (*app.ListResponse, error) {
	client := client.NewClient(baseURL)

	// The req object for a ListRequest contains a `Next` field, which can be used to paginate the results.
	// If the Next field is not empty, it should be used to fetch the next page of results.
	// And the Next field in the ListResponse object should be set to the value returned by the server, if any.
	// In this case, the server takes an int, so it must be converted.
	res, err := client.ListDashboards(ctx, req.Next)
	if err != nil {
		return nil, err
	}

	// resources is a slice of SDK representations of the resources that were listed,
	// and the properties will be stored in the Tempest server.
	// The properties returned will be validated against the ResourceDefinition's properties schema.
	resources := make([]*app.Resource, 0, len(res.Dashboards))
	for _, r := range res.Dashboards {
		resources = append(resources, &app.Resource{
			ExternalID:  r.ID,
			DisplayName: r.Name,
			Properties: map[string]any{
				"id":          r.ID,
				"name":        r.Name,
				"description": r.Description,
				"project_id":  r.Project,
			},
		})
	}

	var next string
	if res.Next != 0 {
		next = strconv.Itoa(res.Next)
	}

	return &app.ListResponse{
		Resources: resources,
		Next:      next, // Include the returned Next value in the response.
	}, nil
}

func healthFn(ctx context.Context) (*app.HealthCheckResponse, error) {
	client := client.NewClient(baseURL)

	err := client.Healthz(ctx)
	if err != nil {
		return nil, err
	}

	return &app.HealthCheckResponse{
		Status:  app.HealthCheckStatusHealthy,
		Message: "The dashboard server is healthy.",
	}, nil
}

var (
	//go:embed schema/create.json
	createSchema []byte
	//go:embed schema/update.json
	updateSchema []byte
)

// Each App must have a function called App() that returns an *app.App.
func App() *app.App {
	// Add various operations to the resource definition.
	resourceDefinition.CreateFn(
		createFn,
		app.MustParseJSONSchema(createSchema),
	)

	resourceDefinition.UpdateFn(
		updateFn,
		app.MustParseJSONSchema(updateSchema),
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
