package noexit

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// Analyzer
var Analyzer = &analysis.Analyzer{
	Name: "osexitcheck",
	Doc:  "checks the calls to os.Exit in function main of package main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name.Name == "main" {
			ast.Inspect(file, func(node ast.Node) bool {
				switch x := node.(type) {
				case *ast.FuncDecl:
					if x.Name.String() != "main" {
						return false
					}

				case *ast.FuncLit:
					return false

				case *ast.CallExpr:
					if sel, ok := x.Fun.(*ast.SelectorExpr); ok {
						if ident, ok := sel.X.(*ast.Ident); ok {
							if ident.Name == "os" && sel.Sel.Name == "Exit" {
								pass.Reportf(sel.Pos(), "call to os.Exit in function main of package main")
							}
						}
					}
				}

				return true
			})
		}
	}

	return nil, nil
}
