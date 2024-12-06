package opentofu

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
)

func (tf *Runner) initCmd() error {
	cmd := exec.Command(tf.openTofuBinary, "init", "-no-color")
	cmd.Dir = tf.workDir
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Env = tf.environment

	return cmd.Run()
}

func (tf *Runner) applyCmd(variables []string) error {
	args := []string{"apply", "-no-color", "-auto-approve", "-input=false"}
	args = append(args, variables...)

	cmd := exec.Command(tf.openTofuBinary, args...)
	cmd.Dir = tf.workDir
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Env = tf.environment

	return cmd.Run()
}

func (tf *Runner) importCmd(variables []string, resourceID, externalID string) error {
	args := []string{"import", "-no-color", "-input=false", resourceID, externalID}
	args = append(args, variables...)

	cmd := exec.Command(tf.openTofuBinary, args...)
	cmd.Dir = tf.workDir
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Env = tf.environment

	return cmd.Run()
}

func (tf *Runner) destroyCmd() error {
	args := []string{"destroy", "-no-color", "-auto-approve", "-input=false", "-refresh=false"}

	cmd := exec.Command(tf.openTofuBinary, args...)
	cmd.Dir = tf.workDir
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Env = tf.environment

	return cmd.Run()
}

func (tf *Runner) showCmd() (*State, error) {
	out := new(bytes.Buffer)

	args := []string{"show", "-json", "-no-color"}
	cmd := exec.Command(tf.openTofuBinary, args...)
	cmd.Dir = tf.workDir
	cmd.Stdout = out
	cmd.Stderr = os.Stderr
	cmd.Env = tf.environment

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	var v State
	if err := json.Unmarshal(out.Bytes(), &v); err != nil {
		return nil, err
	}

	return &v, nil
}
