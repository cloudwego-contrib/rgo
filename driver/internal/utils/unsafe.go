package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/packages"
	"os"
	"path/filepath"

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

//go:linkname driver golang.org/x/tools/go/packages.driver
type driver func(cfg *packages.Config, patterns ...string) (*packages.DriverResponse, error)

//go:linkname defaultDriver golang.org/x/tools/go/packages.defaultDriver
func defaultDriver(cfg *packages.Config, patterns ...string) (*packages.DriverResponse, bool, error)

func UnsafeGetDefaultDriverResponse(cfg *packages.Config, patterns ...string) (*packages.DriverResponse, bool, error) {
	ld := newLoader(cfg)

	return defaultDriver(&ld.Config, patterns...)
}

func UnsafeGetDriverResponse(cfg *packages.Config, patterns ...string) (*packages.DriverResponse, error) {
	ld := newLoader(cfg)
	const (
		// windowsArgMax specifies the maximum command line length for
		// the Windows' CreateProcess function.
		windowsArgMax = 32767
		// maxEnvSize is a very rough estimation of the maximum environment
		// size of a user.
		maxEnvSize = 16384
		// safeArgMax specifies the maximum safe command line length to use
		// by the underlying driver excl. the environment. We choose the Windows'
		// ARG_MAX as the starting point because it's one of the lowest ARG_MAX
		// constants out of the different supported platforms,
		// e.g., https://www.in-ulm.de/~mascheck/various/argmax/#results.
		safeArgMax = windowsArgMax - maxEnvSize
	)

	//// 添加当前项目中的所有包
	//currentModulePatterns := append(patterns, "./...")
	//
	//// 获取 GOPATH 下的所有包
	//gopath := build.Default.GOPATH
	//if gopath != "" {
	//	gopathPatterns := filepath.Join(gopath, "src", "...")
	//	currentModulePatterns = append(currentModulePatterns, gopathPatterns)
	//}

	//// go list fallback
	////
	//// Write overlays once, as there are many calls
	//// to 'go list' (one per chunk plus others too).
	//overlay, cleanupOverlay, err := writeOverlays(cfg.Overlay)
	//if err != nil {
	//	return nil, err
	//}
	//defer cleanupOverlay()
	//ld.Config.goListOverlayFile = overlay

	chunks, err := splitIntoChunks(patterns, safeArgMax)
	if err != nil {
		return nil, err
	}

	response, err := callDriverOnChunks(goListDriver, &ld.Config, chunks)
	if err != nil {
		return nil, err
	}
	return response, err
}

//go:linkname newLoader golang.org/x/tools/go/packages.newLoader
func newLoader(cfg *packages.Config) *loader

//go:linkname callDriverOnChunks golang.org/x/tools/go/packages.callDriverOnChunks
func callDriverOnChunks(driver driver, cfg *packages.Config, chunks [][]string) (*packages.DriverResponse, error)

//go:linkname goListDriver golang.org/x/tools/go/packages.goListDriver
func goListDriver(cfg *packages.Config, patterns ...string) (_ *packages.DriverResponse, err error)

func splitIntoChunks(patterns []string, argMax int) ([][]string, error) {
	if argMax <= 0 {
		return nil, errors.New("failed to split patterns into chunks, negative safe argMax value")
	}
	var chunks [][]string
	charsInChunk := 0
	nextChunkStart := 0
	for i, v := range patterns {
		vChars := len(v)
		if vChars > argMax {
			// a single pattern is longer than the maximum safe ARG_MAX, hardly should happen
			return nil, errors.New("failed to split patterns into chunks, a pattern is too long")
		}
		charsInChunk += vChars + 1 // +1 is for a whitespace between patterns that has to be counted too
		if charsInChunk > argMax {
			chunks = append(chunks, patterns[nextChunkStart:i])
			nextChunkStart = i
			charsInChunk = vChars
		}
	}
	// add the last chunk
	if nextChunkStart < len(patterns) {
		chunks = append(chunks, patterns[nextChunkStart:])
	}
	return chunks, nil
}

func writeOverlays(overlay map[string][]byte) (filename string, cleanup func(), err error) {
	// Do nothing if there are no overlays in the config.
	if len(overlay) == 0 {
		return "", func() {}, nil
	}

	dir, err := os.MkdirTemp("", "gocommand-*")
	if err != nil {
		return "", nil, err
	}

	// The caller must clean up this directory,
	// unless this function returns an error.
	// (The cleanup operand of each return
	// statement below is ignored.)
	defer func() {
		cleanup = func() {
			os.RemoveAll(dir)
		}
		if err != nil {
			cleanup()
			cleanup = nil
		}
	}()

	// Write each map entry to a temporary file.
	overlays := make(map[string]string)
	for k, v := range overlay {
		// Use a unique basename for each file (001-foo.go),
		// to avoid creating nested directories.
		base := fmt.Sprintf("%d-%s", 1+len(overlays), filepath.Base(k))
		filename := filepath.Join(dir, base)
		err := os.WriteFile(filename, v, 0666)
		if err != nil {
			return "", nil, err
		}
		overlays[k] = filename
	}

	// Write the JSON overlay file that maps logical file names to temp files.
	//
	// OverlayJSON is the format overlay files are expected to be in.
	// The Replace map maps from overlaid paths to replacement paths:
	// the Go command will forward all reads trying to open
	// each overlaid path to its replacement path, or consider the overlaid
	// path not to exist if the replacement path is empty.
	//
	// From golang/go#39958.
	type OverlayJSON struct {
		Replace map[string]string `json:"replace,omitempty"`
	}
	b, err := json.Marshal(OverlayJSON{Replace: overlays})
	if err != nil {
		return "", nil, err
	}
	filename = filepath.Join(dir, "overlay.json")
	if err := os.WriteFile(filename, b, 0666); err != nil {
		return "", nil, err
	}

	return filename, nil, nil
}
