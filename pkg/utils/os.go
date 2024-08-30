package utils

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func GetDefaultUserPath() string {
	var homeDir string
	switch runtime.GOOS {
	case "windows":
		homeDir = os.Getenv("USERPROFILE")
	case "darwin":
		homeDir = os.Getenv("HOME")
	case "linux":
		homeDir = os.Getenv("HOME")
	default:
		log.Fatalf("Unsupported OS: %s", runtime.GOOS)
	}
	if homeDir == "" {
		log.Fatal("Cannot get user home directory")
	}
	return homeDir
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

func GetFileNameWithoutExt(filePath string) string {
	base := filepath.Base(filePath)
	nameWithoutExt := strings.TrimSuffix(base, filepath.Ext(base))
	return nameWithoutExt
}

func GetCurrentPathWithUnderline() (string, error) {
	currentPath, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Windows
	if strings.HasPrefix(currentPath, "\\") || strings.Contains(currentPath, ":\\") {
		currentPath = strings.ReplaceAll(currentPath, ":", "")
		currentPath = strings.ReplaceAll(currentPath, "\\", "_")
	} else {
		// Unix
		if strings.HasPrefix(currentPath, "/") {
			currentPath = currentPath[1:]
		}
		currentPath = strings.ReplaceAll(currentPath, "/", "_")
	}

	return currentPath, nil
}

func FindGoModDirectories(root string) ([]string, error) {
	var goModDirs []string

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Name() == "go.mod" {
			dir := filepath.Dir(path)
			goModDirs = append(goModDirs, dir)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return goModDirs, nil
}
