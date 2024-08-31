package internal

import (
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/packages"
	"sync"
	_ "unsafe"
)

//go:linkname loaderPackage golang.org/x/tools/go/packages.loaderPackage
type loaderPackage struct{}

//go:linkname parseValue golang.org/x/tools/go/packages.parseValue
type parseValue struct {
	f     *ast.File
	err   error
	ready chan struct{}
}

//go:linkname loader golang.org/x/tools/go/packages.loader
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
