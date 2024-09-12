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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

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
func FileExistsInPath(dir, filename string) (bool, error) {
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

func GetProjectHashPathWithUnderline() (string, error) {
	currentPath, err := os.Getwd()
	if err != nil {
		return "", err
	}

	currentPath = strings.TrimSpace(currentPath)

	projectName := filepath.Base(currentPath)

	hasher := sha256.New()
	hasher.Write([]byte(currentPath))
	hash := hex.EncodeToString(hasher.Sum(nil))

	hashedPath := fmt.Sprintf("%s_%s", projectName, hash)

	return hashedPath, nil
}

func GetDefaultUserPath() string {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		log.Fatal("Cannot get user home directory")
	}
	return homeDir
}
