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
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/cloudwego-contrib/rgo/pkg/config"
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
	cmd := exec.Command("go", append([]string{"work", "init"}, modules...)...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error initializing Go workspace: %v", err)
	}

	return nil
}

func AddModuleToGoWork(modules ...string) error {
	cmd := exec.Command("go", append([]string{"work", "use"}, modules...)...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error adding module(s) to Go workspace: %v", err)
	}

	return nil
}

func ReplaceModulesInGoWork(oldModule, newModule string) error {
	removeCmd := exec.Command("go", "work", "edit", "-dropuse", oldModule)

	if err := removeCmd.Run(); err != nil {
		return fmt.Errorf("error removing old module from Go workspace: %v", err)
	}

	addCmd := exec.Command("go", "work", "use", newModule)

	addCmd.Stdout = os.Stdout
	addCmd.Stderr = os.Stderr

	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("error adding new module to Go workspace: %v", err)
	}

	return nil
}

func RemoveModulesFromGoWork(modulesToRemove []string) error {
	for _, mod := range modulesToRemove {
		cmd := exec.Command("go", "work", "edit", "-dropuse", mod)

		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to execute 'go work edit -dropuse': %v, output: %s", err, string(output))
		}
	}

	return nil
}

func GetGoWorkJson() (*config.GoWork, error) {
	cmd := exec.Command("go", "work", "edit", "-json")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to execute 'go work list -json': %v, output: %s", err, string(output))
	}

	var goWork *config.GoWork
	if err := json.Unmarshal(output, &goWork); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	return goWork, nil
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
