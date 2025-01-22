package sysinfo

import (
	"errors"
	"runtime"

	"github.com/av1ppp/chafa-welcome/internal/config"
)

func collectOS(conf *config.Config) (string, error) {
	if runtime.GOOS == "darwin" {
		return "Darwin", nil
	}

	if runtime.GOOS == "windows" {
		return "", errors.New("Windows is not supproted")
	}

	data, err := execute("sh", "-c", "cat /etc/*-release")
	if err != nil {
		return "", err
	}

	prettyName, found := findAndTrimLinePrefix(data, "PRETTY_NAME=")
	if found {
		return prettyName + " " + runtime.GOOS, nil
	}

	name, found := findAndTrimLinePrefix(data, "NAME=")
	if found {
		return name + " " + runtime.GOOS, nil
	}

	return runtime.GOOS, nil
}
