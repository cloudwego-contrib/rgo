package utils

import (
	"log"
	"os"
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

// PathExist is used to judge whether the path exists in file system.
func PathExist(path string) (bool, error) {
	abPath, err := filepath.Abs(path)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(abPath)
	if err != nil {
		return os.IsExist(err), nil
	}
	return true, nil
}

// FileExistsInPath checks if a specific file exists at a given path.
func FileExistsInPath(dir string, filename string) (bool, error) {
	abDir, err := filepath.Abs(dir)
	if err != nil {
		return false, err
	}

	filePath := filepath.Join(abDir, filename)

	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return !info.IsDir(), nil
}
