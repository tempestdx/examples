package opentofu

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
)

func (tf *Runner) init() error {
	cmd := exec.Command(tf.openTofuBinary, "init")
	cmd.Dir = tf.workDir
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	return cmd.Run()
}

func (tf *Runner) apply(variables []string) error {
	args := []string{"apply", "-no-color", "-auto-approve", "-input=false"}
	args = append(args, variables...)

	cmd := exec.Command(tf.openTofuBinary, args...)
	cmd.Dir = tf.workDir
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	return cmd.Run()
}

func (tf *Runner) destroy(variables []string) error {
	args := []string{"destroy", "-no-color", "-auto-approve", "-input=false", "-refresh=false"}
	args = append(args, variables...)

	cmd := exec.Command(tf.openTofuBinary, args...)
	cmd.Dir = tf.workDir
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	return cmd.Run()
}

func (tf *Runner) show() (*State, error) {
	out := new(bytes.Buffer)

	args := []string{"show", "-json"}
	cmd := exec.Command(tf.openTofuBinary, args...)
	cmd.Dir = tf.workDir
	cmd.Stdout = out
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	var v State
	if err := json.Unmarshal(out.Bytes(), &v); err != nil {
		return nil, err
	}

	return &v, nil
}
