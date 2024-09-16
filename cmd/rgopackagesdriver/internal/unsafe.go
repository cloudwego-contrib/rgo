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

// nolint
type loaderPackage struct{}

// nolint
type parseValue struct {
	f     *ast.File
	err   error
	ready chan struct{}
}

// nolint
type loader struct {
	pkgs map[string]*loaderPackage
	packages.Config
	sizes        types.Sizes // non-nil if needed by mode
	parseCache   map[string]*parseValue
	parseCacheMu sync.Mutex
	exportMu     sync.Mutex // enforces mutual exclusion of exportdata operations

	// Config.Mode contains the implied mode (see impliedLoadMode).
	// Implied mode contains all the fields we need the data for.
	// In requestedMode there are the actually requested fields.
	// We'll zero them out before returning packages to the user.
	// This makes it easier for us to get the conditions where
	// we need certain modes right.
	requestedMode packages.LoadMode
}

//go:linkname defaultDriver golang.org/x/tools/go/packages.defaultDriver
func defaultDriver(cfg *packages.Config, patterns ...string) (*packages.DriverResponse, bool, error)

func UnsafeGetDefaultDriverResponse(cfg *packages.Config, patterns ...string) (*packages.DriverResponse, bool, error) {
	ld := newLoader(cfg)

	return defaultDriver(&ld.Config, patterns...)
}

//go:linkname newLoader golang.org/x/tools/go/packages.newLoader
func newLoader(cfg *packages.Config) *loader
