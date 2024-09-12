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
	_ "unsafe"

	"golang.org/x/tools/go/packages"
)

type loader struct {
	packages.Config
}

//go:linkname defaultDriver golang.org/x/tools/go/packages.defaultDriver
func defaultDriver(cfg *packages.Config, patterns ...string) (*packages.DriverResponse, bool, error)

func UnsafeGetDefaultDriverResponse(cfg *packages.Config, patterns ...string) (*packages.DriverResponse, bool, error) {
	ld := newLoader(cfg)

	return defaultDriver(&ld.Config, patterns...)
}

//go:linkname newLoader golang.org/x/tools/go/packages.newLoader
func newLoader(cfg *packages.Config) *loader
