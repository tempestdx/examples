package opentofu

import (
	"fmt"
)

func (tf *Runner) Apply(input map[string]any) (*State, error) {
	if err := tf.init(); err != nil {
		return nil, fmt.Errorf("opentofu init: %w", err)
	}

	variables := make([]string, 0, len(input))
	for k, v := range input {
		variables = append(variables, fmt.Sprintf("-var %s=%v", k, v))
	}

	if err := tf.apply(variables); err != nil {
		return nil, fmt.Errorf("opentofu apply: %w", err)
	}

	res, err := tf.show()
	if err != nil {
		return nil, fmt.Errorf("opentofu show: %w", err)
	}

	return res, nil
}

func (tf *Runner) Destroy(input map[string]any) error {
	if err := tf.init(); err != nil {
		return fmt.Errorf("opentofu init: %w", err)
	}

	variables := make([]string, 0, len(input))
	for k, v := range input {
		variables = append(variables, fmt.Sprintf("-var %s=%v", k, v))
	}

	return tf.destroy(variables)
}

func (tf *Runner) Show() (*State, error) {
	if err := tf.init(); err != nil {
		return nil, fmt.Errorf("opentofu init: %w", err)
	}

	state, err := tf.show()
	if err != nil {
		return nil, fmt.Errorf("opentofu show: %w", err)
	}

	return state, nil
}

func (tf *Runner) List(t string) ([]*Resource, error) {
	if err := tf.init(); err != nil {
		return nil, fmt.Errorf("opentofu init: %w", err)
	}

	state, err := tf.show()
	if err != nil {
		return nil, fmt.Errorf("opentofu show: %w", err)
	}

	var resources []*Resource
	for _, r := range state.Values.RootModule.Resources {
		if r.Type == t {
			resources = append(resources, &r)
		}
	}

	return resources, nil
}
