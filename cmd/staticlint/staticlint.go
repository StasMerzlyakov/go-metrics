// Package main contains project multicheck linter:
// - standart golang.org/x/tools/go/analysis
// - SAXXX + QF1004 staticcheck.io
// - github.com/jingyugao/rowserrcheck/passes/rowserr
// - github.com/kisielk/errcheck/errcheck
//
// If `config.json` exists in the current project directory it will be used for reading excludedchecks.
//
// content: {
//
//		  "excludedcheck": ["appends", "QF1004"]
//	}
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jingyugao/rowserrcheck/passes/rowserr"
	"github.com/kisielk/errcheck/errcheck"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/appends"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/slog"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"honnef.co/go/tools/staticcheck"
)

type ConfigData struct {
	ExcludedChecks []string `json:"excludedChecks"`
}

const ConfigX = "config.json"

func main() {

	var analyzers []*analysis.Analyzer
	var cfg ConfigData

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}

	exPath := filepath.Dir(ex)
	confFilePath := filepath.Join(exPath, ConfigX)
	fmt.Println(confFilePath)

	data, err := os.ReadFile(ConfigX)
	if err != nil {
		fmt.Printf("can't find conf file: %s; used default conifguration\n", ConfigX)
	} else {
		if err = json.Unmarshal(data, &cfg); err != nil {
			panic(err)
		}
	}

	excludedChecks := make(map[string]bool)
	for _, v := range cfg.ExcludedChecks {
		excludedChecks[v] = true
	}

	standardAnalyzers := []*analysis.Analyzer{
		appends.Analyzer,          // check for missing values after append
		asmdecl.Analyzer,          // report mismatches between assembly files and Go declarations
		assign.Analyzer,           // check for useless assignments
		atomic.Analyzer,           // check for common mistakes using the sync/atomic package
		bools.Analyzer,            // check for common mistakes involving boolean operators
		buildtag.Analyzer,         // check //go:build and // +build directives
		cgocall.Analyzer,          // detect some violations of the cgo pointer passing rules
		composite.Analyzer,        // check for unkeyed composite literals
		copylock.Analyzer,         // check for locks erroneously passed by value
		defers.Analyzer,           // report common mistakes in defer statements
		directive.Analyzer,        // check Go toolchain directives such as //go:debug
		errorsas.Analyzer,         // report passing non-pointer or non-error values to errors.As
		framepointer.Analyzer,     // report assembly that clobbers the frame pointer before saving it
		httpresponse.Analyzer,     // check for mistakes using HTTP responses
		ifaceassert.Analyzer,      // detect impossible interface-to-interface type assertions
		loopclosure.Analyzer,      // check references to loop variables from within nested functions
		lostcancel.Analyzer,       // check cancel func returned by context.WithCancel is called
		nilfunc.Analyzer,          // check for useless comparisons between functions and nil
		printf.Analyzer,           // check consistency of Printf format strings and arguments
		shift.Analyzer,            // check for shifts that equal or exceed the width of the integer
		sigchanyzer.Analyzer,      // check for unbuffered channel of os.Signal
		slog.Analyzer,             // check for invalid structured logging calls
		stdmethods.Analyzer,       // check signature of methods of well-known interfaces
		stringintconv.Analyzer,    // check for string(int) conversions
		structtag.Analyzer,        // check that struct field tags conform to reflect.StructTag.Get
		testinggoroutine.Analyzer, // report calls to (*testing.T).Fatal from goroutines started by a test
		tests.Analyzer,            // check for common mistaken usages of tests and examples
		timeformat.Analyzer,       // check for calls of (time.Time).Format or time.Parse with 2006-02-01
		unmarshal.Analyzer,        // report passing non-pointer or non-interface values to unmarshal
		unreachable.Analyzer,      // check for unreachable code
		unsafeptr.Analyzer,        // check for invalid conversions of uintptr to unsafe.Pointer
		unusedresult.Analyzer,     // check for unused results of calls to some functions
	}

	// golang.org/x/tools/go/analysis/passes analazers
	for _, v := range standardAnalyzers {
		if !excludedChecks[v.Name] {
			analyzers = append(analyzers, v)
		}
	}

	// SA + QF1004 staticcheck.io
	for _, v := range staticcheck.Analyzers {
		if !excludedChecks[v.Analyzer.Name] {
			if strings.HasPrefix(v.Analyzer.Name, "SA") ||
				v.Analyzer.Name == "QF1004" { // use strings.ReplaceAll instead of strings.Replace with n == -1
				analyzers = append(analyzers, v.Analyzer)
			}
		}
	}

	// errcheck
	errCheck := errcheck.Analyzer
	if !excludedChecks[errCheck.Name] {
		analyzers = append(analyzers, errCheck) // check for unchecked errors in Go code.
	}

	// rowserrcheck
	rser := rowserr.NewAnalyzer(
		"github.com/StasMerzlyakov/go-metrics/internal/server/adapter/storage/postgres",
	)

	if !excludedChecks[rser.Name] {
		analyzers = append(analyzers, rser) // rowserrcheck is a static analysis tool which checks whether sql.Rows.Err is correctly checked
	}

	multichecker.Main(
		analyzers...,
	)
}
