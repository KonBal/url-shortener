// Adds analyzers to staticcheck and errcheck
// Used anaylyzers:
// passes/appends, passes/assign, bodyclose/passes/bodyclose (checks that body is closed),
// passes/bools, passes/defers, errcheck (checks that all errors are handled),
// passes/lostcancel, passes/nilfunc, passes/printf, passes/shadow, passes/stringintconv, passes/structtag,
// passes/unmarshal, passes/unreachable, passes/unusedresult
// Also checks that os.Exit is not called in main function in package main.
// To run the check, build the executable and run it on the package you want to check.
package main

import (
	"strings"

	"github.com/KonBal/url-shortener/cmd/staticlint/noexit"
	"github.com/kisielk/errcheck/errcheck"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/appends"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"honnef.co/go/tools/staticcheck"
)

func main() {
	statChecks := []*analysis.Analyzer{
		appends.Analyzer,
		assign.Analyzer,
		bodyclose.Analyzer,
		bools.Analyzer,
		defers.Analyzer,
		errcheck.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,
		stringintconv.Analyzer,
		structtag.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unusedresult.Analyzer,
		noexit.Analyzer,
	}

	checks := map[string]bool{
		"S1008":  true,
		"S1036":  true,
		"ST1005": true,
		"ST1015": true,
		"QF1001": true,
		"QF1003": true,
	}

	for _, v := range staticcheck.Analyzers {
		if strings.HasSuffix(v.Analyzer.Name, "SA") || checks[v.Analyzer.Name] {
			statChecks = append(statChecks, v.Analyzer)
		}
	}

	multichecker.Main(
		statChecks...,
	)
}
