package opentofu

import (
	"fmt"
)

// Apply runs "opentofu apply" with the given input variables.
// The returned state is a parsed version of the JSON output from "opentofu show".
// This output contains the properties and values of the resource(s) created by the module.
func (tf *Runner) Apply(input map[string]any) (*State, error) {
	if err := tf.initCmd(); err != nil {
		return nil, fmt.Errorf("opentofu init: %w", err)
	}

	variables := make([]string, 0, len(input))
	for k, v := range input {
		arg := []string{"-var", fmt.Sprintf("%s=%v", k, v)}
		variables = append(variables, arg...)
	}

	if err := tf.applyCmd(variables); err != nil {
		return nil, fmt.Errorf("opentofu apply: %w", err)
	}

	state, err := tf.showCmd()
	if err != nil {
		return nil, fmt.Errorf("opentofu show: %w", err)
	}

	return state, nil
}

// Import will run "opentofu import" for each resource ID in the given map.
// The map associates resource IDs to their external IDs.
// This is only needed if running OpenTofu in a "stateless" mode.
// If you're using remote state, you can ignore this method, and just use Apply.
func (tf *Runner) Import(input map[string]any, resourceIDsToExternalIDs map[string]string) error {
	if err := tf.initCmd(); err != nil {
		return fmt.Errorf("opentofu init: %w", err)
	}

	variables := make([]string, 0, len(input))
	for k, v := range input {
		arg := []string{"-var", fmt.Sprintf("%s=%v", k, v)}
		variables = append(variables, arg...)
	}

	for id, externalID := range resourceIDsToExternalIDs {
		err := tf.importCmd(variables, id, externalID)
		if err != nil {
			return fmt.Errorf("opentofu import: %w", err)
		}
	}

	return nil
}

// Destroy runs "opentofu destroy" to remove all resources created by the module.
func (tf *Runner) Destroy() error {
	if err := tf.initCmd(); err != nil {
		return fmt.Errorf("opentofu init: %w", err)
	}

	return tf.destroyCmd()
}

// Show runs "opentofu show" and returns the parsed state.
// This output contains the properties and values of the resource(s) created by the module.
func (tf *Runner) Show() (*State, error) {
	if err := tf.initCmd(); err != nil {
		return nil, fmt.Errorf("opentofu init: %w", err)
	}

	state, err := tf.showCmd()
	if err != nil {
		return nil, fmt.Errorf("opentofu show: %w", err)
	}

	return state, nil
}
