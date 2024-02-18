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
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
)

// BuildDependence store the path of the b001 file obtained by compiling the test code to TEMPDIR/b001.txt
func BuildDependence(genGoFilePath, genExePath, genTxtPath, code string) error {
	err := os.WriteFile(genGoFilePath, []byte(code), 0o777)
	if err != nil {
		return err
	}

	buildCmd := strings.Split(fmt.Sprintf("build -a -work -o %s %s", genExePath, genGoFilePath), " ")
	output, err := exec.Command("go", buildCmd...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("output: %v, err: %v", string(output), err)
	}
	if !strings.HasPrefix(string(output), "WORK=") {
		return errors.New("bad output")
	}
	b001 := path.Join(strings.TrimPrefix(strings.TrimSpace(string(output)), "WORK="), "b001")

	f, err := os.OpenFile(genTxtPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o777)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = f.WriteString(b001 + "\n"); err != nil {
		return err
	}
	return nil
}

// MergeImportCfg merge tempCfg into originalCfg.
// When it is the compile step, the cfg parameter should be `importcfg`.
// When it is the link step, the cfg parameter should be `importcfg.link`.
func MergeImportCfg(genTxtPath, originalCfg, cfg string, isCompilePhase bool) error {
	b001txt, err := os.ReadFile(genTxtPath)
	if err != nil {
		return err
	}
	b001Path := strings.TrimSpace(string(b001txt))
	b001SlicePath := strings.Split(b001Path, "\n")
	tempPaths := make([]string, 0, len(b001SlicePath))
	if isCompilePhase {
		tempPaths = append(tempPaths, path.Join(b001SlicePath[len(b001SlicePath)-1], cfg))
	} else {
		for _, p := range b001SlicePath {
			tempPaths = append(tempPaths, path.Join(p, cfg))
		}
	}

	parseCfg := func(r io.Reader) map[string]string {
		res := make(map[string]string)
		s := bufio.NewScanner(r)
		for s.Scan() {
			line := s.Text()
			if !strings.HasPrefix(line, "packagefile") {
				continue
			}
			index := strings.Index(line, "=")
			if index < 11 {
				continue
			}
			res[line[:index-1]] = line
		}
		return res
	}

	f, err := os.OpenFile(originalCfg, os.O_RDWR|os.O_APPEND, 0o777)
	if err != nil {
		return err
	}
	defer f.Close()

	fns := make([]*os.File, 0, len(tempPaths))
	for _, tmpPath := range tempPaths {
		fn, err := os.Open(tmpPath)
		if err != nil {
			return err
		}
		fns = append(fns, fn)
	}
	defer func() {
		for _, fn := range fns {
			fn.Close()
		}
	}()

	w := bufio.NewWriter(f)
	originalPkgs := parseCfg(f)
	for _, fn := range fns {
		tempPkg := parseCfg(fn)
		for name, line := range tempPkg {
			if _, ok := originalPkgs[name]; !ok {
				w.WriteString(line + "\n")
			}
		}
	}
	w.Flush()

	return nil
}
