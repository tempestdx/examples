package opentofu

import (
	"io/fs"
	"os"
)

type Runner struct {
	openTofuBinary string
	workDir        string
	environment    []string
}

func New(tfPath string, moduleFS fs.FS, environment map[string]string) (*Runner, error) {
	tmpDir, err := os.MkdirTemp("", "opentofu")
	if err != nil {
		return nil, err
	}

	err = os.CopyFS(tmpDir, moduleFS)
	if err != nil {
		return nil, err
	}

	var env []string
	for k, v := range environment {
		env = append(env, k+"="+v)
	}

	return &Runner{
		openTofuBinary: tfPath,
		workDir:        tmpDir,
		environment:    env,
	}, nil
}
