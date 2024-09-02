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
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAddModuleToGoWork(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "testgowork")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	modules := []string{"module1", "module2"}
	for _, module := range modules {
		modulePath := filepath.Join(tempDir, module)
		err := os.Mkdir(modulePath, 0755)
		if err != nil {
			t.Fatalf("Failed to create module directory %s: %v", module, err)
		}
		err = InitGoMod(module, modulePath)
		if err != nil {
			t.Fatalf("Failed to create go.mod for module %s: %v", module, err)
		}
	}

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change directory to %s: %v", tempDir, err)
	}
	err = InitGoWork(modules...)
	if err != nil {
		t.Fatalf("InitGoWork returned an error: %v", err)
	}
	// Call AddModuleToGoWork function
	err = AddModuleToGoWork(modules...)
	if err != nil {
		t.Fatalf("AddModuleToGoWork returned an error: %v", err)
	}

	// Check if go.work file is created
	goWorkPath := filepath.Join(tempDir, "go.work")
	if _, err := os.Stat(goWorkPath); os.IsNotExist(err) {
		t.Fatalf("go.work file was not created")
	}

	// Check the content of go.work file
	content, err := os.ReadFile(goWorkPath)
	if err != nil {
		t.Fatalf("Failed to read go.work file: %v", err)
	}

	for _, module := range modules {
		if !contains(content, module) {
			t.Errorf("go.work file does not contain module: %s", module)
		}
	}
}

func TestInitGoMod(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "testgomod")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	moduleName := "example.com/testmodule"

	err = InitGoMod(moduleName, tempDir)
	if err != nil {
		t.Fatal(err)
	}

	goModPath := filepath.Join(tempDir, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		t.Fatal("go.mod not found")
	}

	content, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatalf("cannot read go.mod: %v", err)
	}

	if !contains(content, "module "+moduleName) {
		t.Errorf("go.mod has not module: %s", moduleName)
	}
}

func TestInitGoWork(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "testgowork")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create dummy module directories
	modules := []string{"module1", "module2"}
	for _, module := range modules {
		modulePath := filepath.Join(tempDir, module)
		err := os.Mkdir(modulePath, 0755)
		if err != nil {
			t.Fatalf("Failed to create module directory %s: %v", module, err)
		}
		err = InitGoMod(module, modulePath)
		if err != nil {
			t.Fatalf("Failed to create go.mod for module %s: %v", module, err)
		}
	}

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change directory to %s: %v", tempDir, err)
	}
	// Call InitGoWork function
	err = InitGoWork(modules...)
	if err != nil {
		t.Fatalf("InitGoWork returned an error: %v", err)
	}

	// Check if go.work file is created
	goWorkPath := filepath.Join(tempDir, "go.work")
	if _, err := os.Stat(goWorkPath); os.IsNotExist(err) {
		t.Fatalf("go.work file was not created")
	}

	// Check the content of go.work file
	content, err := os.ReadFile(goWorkPath)
	if err != nil {
		t.Fatalf("Failed to read go.work file: %v", err)
	}

	for _, module := range modules {
		if !contains(content, module) {
			t.Errorf("go.work file does not contain module: %s", module)
		}
	}
}

func TestRemoveModulesFromGoWork(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "testgowork")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a temporary go.work file
	goWorkPath := filepath.Join(tempDir, "go.work")
	content := `use ./module1
use ./module2
use ./module3
`
	if err := os.WriteFile(goWorkPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write go.work file: %v", err)
	}

	// Modules to remove
	modulesToRemove := []string{"./module2", "./module3"}

	modules := []string{"module1", "module2"}
	for _, module := range modules {
		modulePath := filepath.Join(tempDir, module)
		err := os.Mkdir(modulePath, 0755)
		if err != nil {
			t.Fatalf("Failed to create module directory %s: %v", module, err)
		}
		err = InitGoMod(module, modulePath)
		if err != nil {
			t.Fatalf("Failed to create go.mod for module %s: %v", module, err)
		}
	}

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change directory to %s: %v", tempDir, err)
	}

	// Call RemoveModulesFromGoWork function
	err = RemoveModulesFromGoWork(goWorkPath, modulesToRemove)
	if err != nil {
		t.Fatalf("RemoveModulesFromGoWork returned an error: %v", err)
	}

	// Check the content of go.work file
	updatedContent, err := os.ReadFile(goWorkPath)
	if err != nil {
		t.Fatalf("Failed to read go.work file: %v", err)
	}

	expectedContent := `use ./module1
`
	if string(updatedContent) != expectedContent {
		t.Errorf("go.work file content mismatch. Expected:\n%s\nGot:\n%s", expectedContent, string(updatedContent))
	}
}

func TestRunGoModTidyInDir(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "testgomodtidy")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a dummy go.mod file in the temporary directory
	goModPath := filepath.Join(tempDir, "go.mod")
	content := `module example.com/testmodule`
	if err := os.WriteFile(goModPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write go.mod file: %v", err)
	}

	// Call RunGoModTidyInDir function
	err = RunGoModTidyInDir(tempDir)
	if err != nil {
		t.Fatalf("RunGoModTidyInDir returned an error: %v", err)
	}

	// Check if go.mod file is updated (you can add more checks based on expected changes)
	updatedContent, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatalf("Failed to read go.mod file: %v", err)
	}

	if string(updatedContent) == content {
		t.Errorf("go.mod file was not updated by 'go mod tidy'")
	}
}

func contains(content []byte, substr string) bool {
	return strings.Contains(string(content), substr)
}
