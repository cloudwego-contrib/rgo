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

package internal

import (
	"go/ast"
	"go/types"
	"sync"
	_ "unsafe"

	"golang.org/x/tools/go/packages"
)

//nolint:unused
type loaderPackage struct {
	*packages.Package
	importErrors map[string]error // maps each bad import to its error
	loadOnce     sync.Once
	color        uint8 // for cycle detection
	needsrc      bool  // load from source (Mode >= LoadTypes)
	needtypes    bool  // type information is either requested or depended on
	initial      bool  // package was matched by a pattern
	goVersion    int   // minor version number of go command on PATH
}

//nolint:unused
type parseValue struct {
	f     *ast.File
	err   error
	ready chan struct{}
}

type loader struct {
	pkgs map[string]*loaderPackage //nolint:unused
	packages.Config
	// non-nil if needed by mode
	sizes        types.Sizes            //nolint:unused
	parseCache   map[string]*parseValue //nolint:unused
	parseCacheMu sync.Mutex             //nolint:unused
	// enforces mutual exclusion of exportdata operations
	exportMu sync.Mutex //nolint:unused

	// Config.Mode contains the implied mode (see impliedLoadMode).
	// Implied mode contains all the fields we need the data for.
	// In requestedMode there are the actually requested fields.
	// We'll zero them out before returning packages to the user.
	// This makes it easier for us to get the conditions where
	// we need certain modes right.
	requestedMode packages.LoadMode //nolint:unused
}

//go:linkname defaultDriver golang.org/x/tools/go/packages.defaultDriver
func defaultDriver(cfg *packages.Config, patterns ...string) (*packages.DriverResponse, bool, error)

func UnsafeGetDefaultDriverResponse(cfg *packages.Config, patterns ...string) (*packages.DriverResponse, bool, error) {
	ld := newLoader(cfg)

	return defaultDriver(&ld.Config, patterns...)
}

//go:linkname newLoader golang.org/x/tools/go/packages.newLoader
func newLoader(cfg *packages.Config) *loader
