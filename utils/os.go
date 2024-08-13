package utils

import (
	"log"
	"os/user"
	"path/filepath"
	"runtime"
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
