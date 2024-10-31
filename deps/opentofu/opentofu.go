package opentofu

import (
	"log/slog"
	"os"
)

type Runner struct {
	openTofuBinary string
	workDir        string
	logger         *slog.Logger
}

type Options struct {
	WorkDir      string
	OpenTofuPath string
}

func New(tfPath, workdir string) (*Runner, error) {
	return &Runner{
		openTofuBinary: tfPath,
		workDir:        workdir,
		logger:         slog.New(slog.NewJSONHandler(os.Stdout, nil)),
	}, nil
}
