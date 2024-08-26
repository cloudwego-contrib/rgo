package utils

import (
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
)

func GetDefaultUserPath() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	var userPath string
	switch runtime.GOOS {
	case "windows":
		userPath = filepath.Join("C:\\Users", usr.Username)
	case "darwin":
		userPath = filepath.Join("/Users", usr.Username)
	case "linux":
		userPath = filepath.Join("/home", usr.Username)
	default:
		log.Fatalf("Unsupported OS: %s", runtime.GOOS)
	}

	return userPath
}

func GetCurrentPathWithUnderline() (string, error) {
	currentPath, err := os.Getwd()
	if err != nil {
		return "", err
	}

	//todo:windows
	if strings.HasPrefix(currentPath, "/") {
		currentPath = currentPath[1:]
	}

	currentPath = strings.ReplaceAll(currentPath, "/", "_")

	return currentPath, nil
}
