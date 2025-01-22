package sysinfo

import (
	"errors"
	"runtime"
	"strings"

	"github.com/av1ppp/chafa-welcome/internal/config"
)

func collectUptime(conf *config.Config) (string, error) {
	if runtime.GOOS == "darwin" {
		s, err := execute("system_profiler", "SPSoftwareDataType", "-detailLevel", "mini")
		if err != nil {
			return "", err
		}
		for _, line := range strings.Split(s, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "Time since boot: ") {
				return strings.TrimPrefix(line, "Time since boot: "), nil
			}
		}
		return "", errors.New("unknown system_profiler format")
	}
	return execute("uptime", "-p")
}
