package helloworld

import (
	"context"
	_ "embed"
	"fmt"
	"strconv"
	"time"

	"github.com/tempestdx/sdk-go/app"
)

// The create function should perform the operations necessary to create the resource in the external system.
// This could be the result of multiple API calls, execution of a script, or any other operation.
// In this example, we are simply returning a new resource with a generated ExternalID.
//
// Check the "dashboards" and "opentofu" examples for more complex examples.
// https://github.com/tempestdx/examples
func createFn(ctx context.Context, req *app.OperationRequest) (*app.OperationResponse, error) {
	// The incoming request will contain the input values that the user has provided,
	// the environment variables and secrets that are configured for the Project, as well as Metadata about the Project.

	// ... perform some operation to create the resource with the specified color in the external system ...
	// id, err := client.Create(ctx, req.Input["color"])

	color := req.Input["color"]
	id := strconv.Itoa(int(time.Now().Unix()))

	// Construct the OperationResponse which contains the Resource that was created.
	// Tempest will store these values in the Resource and relate it to the Project that created it.
	return &app.OperationResponse{
		Resource: &app.Resource{
			// The ExternalID is the unique identifier for the resource in the external system.
			// In AWS, this may be the ARN. A GitHub repo would be identified by the owner and repo name (org/repo).
			ExternalID: id,
			// DisplayName is the name of the resource that will be displayed in the Tempest UI.
			DisplayName: fmt.Sprintf("Example Resource - %s", color),
			// These properties should match what is defined in the `properties.json` schema.
			Properties: map[string]any{
				"color": color,
			},
			// Links is a list of links that will be displayed in the UI.
			// There are different types of links that can be displayed:
			// - External: The link to view the resource in the external system.
			// - Support: The link to the support page for the individual resource.
			// - Endpoint: The link to the endpoint of the resource. For example, the connection string for a database.
			// - Documentation: The link to the documentation for the resource.
			// - Administration: The link to the administration page for the resource.
			Links: []*app.Link{
				{
					URL:   fmt.Sprintf("https://example.com/resource/%s", id),
					Title: "Example.com Resource Console",
					Type:  app.LinkTypeExternal,
				},
			},
		},
	}, nil
}

// The update function should perform the operations necessary to update the resource in the external system.
// In this example, we are simply returning the same resource with a new DisplayName.
func updateFn(ctx context.Context, req *app.OperationRequest) (*app.OperationResponse, error) {
	// The incoming request will contain the input values that the user has provided,
	// the environment variables and secrets that are configured for the Project, as well as Metadata about the Project.
	//
	// For an Update operation, the req.Resource object will also be populated with the current state of the resource in Tempest.
	// This includes the ExternalID, DisplayName, and Properties.

	// ... perform some operation to update the resource's color in the external system ...
	// newColor, err := client.Update(ctx, req.Resource.ExternalID, req.Input["color"])
	newColor := req.Input["color"]

	// Construct the updated Resource and wrap it in an OperationResponse.
	// Tempest will update the values of the Resource according to the values returned here.
	return &app.OperationResponse{
		Resource: &app.Resource{
			// Fields that aren't changing (such as the ExternalID) can be omitted.
			DisplayName: fmt.Sprintf("Example Resource - %s", newColor),
			Properties: map[string]any{
				"color": newColor,
			},
		},
	}, nil
}

// The read function should perform the operations necessary to gather the current state of the resource in the external system.
func readFn(ctx context.Context, req *app.OperationRequest) (*app.OperationResponse, error) {
	// The incoming request will contain the ExternalID of the resource that needs to be retrieved in the req.Resource object,
	// the environment variables and secrets that are configured for the Project, as well as Metadata about the Project.

	// ... perform some operation to read the resource and its color from the external system ...
	// color, err := client.Get(ctx, req.Resource.ExternalID)
	color := "blue"

	// Construct the OperationResponse which contains the Resource that was read.
	// Tempest will store these values in the Resource and relate it to the Project or Import Policy.
	return &app.OperationResponse{
		Resource: &app.Resource{
			DisplayName: fmt.Sprintf("Example Resource - %s", color),
			Properties: map[string]any{
				"color": color,
			},
			Links: []*app.Link{
				{
					URL:   fmt.Sprintf("https://example.com/resource/%s", req.Resource.ExternalID),
					Title: "Example.com Resource Console",
					Type:  app.LinkTypeExternal,
				},
			},
		},
	}, nil
}

// The delete function should perform the operations necessary to delete the resource in the external system.
func deleteFn(ctx context.Context, req *app.OperationRequest) (*app.OperationResponse, error) {
	// The incoming request will contain the ExternalID of the resource that needs to be deleted in the req.Resource object,
	// the environment variables and secrets that are configured for the Project, as well as Metadata about the Project.
	//
	// For a Delete operation, the req.Resource object will also be populated with the current state of the resource in Tempest.
	// This includes the ExternalID, DisplayName, and Properties.

	// ... perform some operation to delete the resource from the external system ...
	// err := client.Delete(ctx, req.Resource.ExternalID)

	// Construct the OperationResponse which contains the Resource that was deleted.
	return &app.OperationResponse{
		Resource: &app.Resource{
			ExternalID: req.Resource.ExternalID,
		},
	}, nil
}

// The list function should perform the operations necessary to list all resources of this type in the external system.
func listFn(ctx context.Context, req *app.ListRequest) (*app.ListResponse, error) {
	// The ListRequest includes Metadata about the Project.
	// It also includes a special "Next" field that can be used to paginate the results.
	// Tempest will set the "Next" field to "" when it is the first call.

	// ... perform some operation to list all resources of this type in the external system ...
	// resources, err := client.List(ctx)
	resource1 := &app.Resource{
		ExternalID:  "1234567890",
		DisplayName: "Example Resource - Blue",
		Properties: map[string]any{
			"color": "blue",
		},
	}
	resource2 := &app.Resource{
		ExternalID:  "0987654321",
		DisplayName: "Example Resource - Red",
		Properties: map[string]any{
			"color": "red",
		},
	}

	// Construct the ListResponse which contains a list of Resources in the external system.
	return &app.ListResponse{
		Resources: []*app.Resource{
			resource1,
			resource2,
		},
		// Optionally, set the "Next" field to a value that can be used to paginate the results.
		Next: "2",
	}, nil
}

var (
	//go:embed instructions.md
	instructions string

	//go:embed schema/properties.json
	properties []byte

	//go:embed schema/input.json
	input []byte
)

// App is the main function that is called by the Tempest SDK and returns the app object.
// The name of the function must be 'App' and it must return an *app.App object.
func App() *app.App {
	// The ResourceDefinition is the main object that defines the resource as part of the app.
	// An App can have multiple ResourceDefinitions configured, but the 'Type' must be unique.
	exampleDef := app.ResourceDefinition{
		// Type is the internal identifier for the resource type, and must be unique within the app.
		Type: "example",
		// Description is a short description of the resource type that will be displayed in the Tempest UI.
		Description: "And example resource for the Hello World app.",
		// DisplayName is the name of the resource type that will be displayed in the Tempest UI.
		DisplayName: "Example Resource",
		// The LifecycleStage is the stage of the resource in the Developer Lifecycle.
		// The options, in order, are:
		// LifecycleStageCode
		// LifecycleStageBuild
		// LifecycleStageTest
		// LifecycleStageRelease
		// LifecycleStageDeploy
		// LifecycleStageOperate
		// LifecycleStageMonitor
		LifecycleStage: app.LifecycleStageCode,
		// The PropertiesSchema is the JSON schema that defines the properties of the individual resource.
		// After every Create, Update, Read, or List operation, the properties returned by the function
		// will be validated against this schema.
		// The properties in this schema can be used as input to other recipe steps in the UI.
		//
		// This example schema defines one property, 'name', which is a string.
		PropertiesSchema: app.MustParseJSONSchema(app.GenericEmptySchema),
		// Links are a list of links that will be displayed in the UI for the resource type.
		// These links should be relevant to the resource type and provide additional information or actions.
		// Individual resources can also have links that are specific to that resource.
		//
		// There are different types of links that can be displayed:
		// - Support: The link to the support page for the resource type.
		// - Documentation: The link to the documentation for the resource type.
		// - Administration: The link to the administration page for the resource type.
		Links: []app.Link{
			{
				URL:   "http://example.com",
				Title: "Example Link",
				Type:  app.LinkTypeDocumentation,
			},
		},
		// InstructionsMarkdown is the markdown content that will be displayed in the UI for the resource type.
		// This can include instructions on how to use the resource, best practices, or other helpful information.
		InstructionsMarkdown: instructions,
	}

	// Assign a function for each of the CRUD operations.

	// Create is called when Tempest is instructed to create a new resource in the external system.
	//
	// Create includes a schema that defines the shape of the input.
	// The Tempest UI will display these inputs to the user when creating or updating a resource as part of a Recipe.
	// That input will be validated against the schema before the function is called.
	exampleDef.CreateFn(
		// The function handler that will be run when the Create operation is called.
		createFn,
		// The schema that defines the input for the Create operation.
		app.MustParseJSONSchema(input),
	)

	// Update is called when Tempest is instructed to update an existing resource in the external system.
	// This process can happen as part of a Deployment in a Project.
	//
	// Update includes a schema that defines the shape of the input.
	// The Tempest UI will display these inputs to the user when creating or updating a resource as part of a Recipe.
	// That input will be validated against the schema before the function is called.
	exampleDef.UpdateFn(
		// The function handler that will be run when the Update operation is called.
		updateFn,
		// The schema that defines the input for the Update operation.
		app.MustParseJSONSchema(input),
	)

	// Read and Delete functions are also assigned to the resource definition.
	// There is no input schema for these functions, as they do not take input from the user.
	// Environment Variables and Secrets can still be accessed from these functions, however.

	// Read is called when Tempest needs to sync the state of the resource with the external system.
	exampleDef.ReadFn(readFn)
	// Delete is called when Tempest is instructed to delete the resource from the external system.
	exampleDef.DeleteFn(deleteFn)

	// List is called when Tempest needs to list all resources of this type in the external system.
	// This process happens as part of a Resource Import Policy.
	exampleDef.ListFn(listFn)

	// Add a healthcheck function to the resource definition.
	// This will allow Tempest to monitor and display the health of the resource provider.
	// The function in this case is fairly simple, as it just returns a healthy status.
	// More complex health checks can be implemented as needed, such as checking the status of the external system.
	exampleDef.HealthCheckFn(func(ctx context.Context) (*app.HealthCheckResponse, error) {
		return &app.HealthCheckResponse{
			Status: app.HealthCheckStatusHealthy,
		}, nil
	})

	// Finally, add the resource definition(s) to the app and return the configured App object.
	return app.New(
		app.WithResourceDefinition(exampleDef),
	)
}
