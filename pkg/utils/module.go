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
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func InitGoMod(moduleName, path string) error {
	cmd := exec.Command("go", "mod", "init", moduleName)

	cmd.Dir = path

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to initialize go.mod in path '%s': %w", path, err)
	}

	// Set Go version to 1.18
	cmd = exec.Command("go", "mod", "edit", "-go=1.18")
	cmd.Dir = path

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to set Go version 1.18 in path '%s': %w", path, err)
	}

	return nil
}

func InitGoWork(modules ...string) error {
	args := append([]string{"work", "init"}, modules...)
	cmd := exec.Command("go", args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return errors.New(fmt.Sprintf("Error initializing Go workspace: %v\n", err))
	}

	return nil
}

func AddModuleToGoWork(modules ...string) error {
	args := append([]string{"work", "use"}, modules...)
	cmd := exec.Command("go", args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Error adding module(s) to Go workspace: %v", err)
	}

	return RunGoWorkSync()
}

func RemoveModulesFromGoWork(workFilePath string, modulesToRemove []string) error {
	// Read the go.work file content
	file, err := os.Open(workFilePath)
	if err != nil {
		return fmt.Errorf("failed to open go.work: %v", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)

	// Store each line that doesn't contain the module path to remove
	for scanner.Scan() {
		line := scanner.Text()
		shouldRemove := false

		for _, module := range modulesToRemove {
			if strings.Contains(line, module) {
				shouldRemove = true
				break
			}
		}

		if !shouldRemove {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading go.work: %v", err)
	}

	// Write the updated content back to the go.work file
	file, err = os.Create(workFilePath)
	if err != nil {
		return fmt.Errorf("failed to create go.work: %v", err)
	}
	defer file.Close()

	for _, line := range lines {
		_, err = file.WriteString(line + "\n")
		if err != nil {
			return fmt.Errorf("error writing to go.work: %v", err)
		}
	}

	return RunGoWorkSync()
}

func RunGoWorkSync() error {
	cmd := exec.Command("go", "work", "sync")

	// Run the command and capture output or error
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute 'go work sync': %v, output: %s", err, string(output))
	}

	return nil
}

func RunGoModTidyInDir(dir string) error {
	cmd := exec.Command("go", "mod", "tidy")

	if dir == "" {
		dir = "."
	}

	cmd.Dir = dir

	// Run the command and capture output or error
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute 'go mod tidy' in directory %s: %v, output: %s", dir, err, string(output))
	}

	return nil
}
