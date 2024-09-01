/*
 * Copyright 2024 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
		strings.TrimSpace(currentPath)
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
