package opentofu

// State represents OpenTofu state file.
type State struct {
	Values Value `json:"values"`
}

// Value holds the root module of the state.
type Value struct {
	RootModule RootModule `json:"root_module"`
}

// RootModule details the resources within the module.
type RootModule struct {
	Resources []Resource `json:"resources"`
}

// Resource describes a single resource, capturing its values.
type Resource struct {
	Address string         `json:"address"`
	Type    string         `json:"type"`
	Name    string         `json:"name"`
	Values  map[string]any `json:"values"`
}
