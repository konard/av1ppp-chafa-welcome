package global

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var (
	homeDir     = ""
	homeDirOnce = &sync.Once{}
)

const (
	ModeFile = 0664
	ModeDir  = os.ModePerm
)

func HomeDir() string {
	homeDirOnce.Do(func() {
		userHomeDir, err := os.UserHomeDir()
		if err != nil {
			msg := fmt.Sprintf("Failed to get user home dir: %s", err.Error())
			panic(msg)
		}
		homeDir = filepath.Join(userHomeDir, ".chafa-welcome")

		err = os.MkdirAll(homeDir, os.ModePerm)
		if err != nil {
			msg := fmt.Sprintf("Failed to make dir ~/.chafa-welcome: %s", err.Error())
			panic(msg)
		}
	})

	return homeDir
}
