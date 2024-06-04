package osexit

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "osexit",
	Doc:  "prohibits the use of os.Exit in the main function of the main package",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	expr := func(fnDecl *ast.FuncDecl) {
		if fnDecl.Name.Name == "main" {
			for _, stmt := range fnDecl.Body.List {
				if exptStmt, ok := stmt.(*ast.ExprStmt); ok {
					if call, ok := exptStmt.X.(*ast.CallExpr); ok {
						if isOsExitCall(call) {
							pass.Reportf(exptStmt.Pos(), "os.Exit in main function")
						}
					}
				}

			}
		}
	}

	for _, file := range pass.Files {
		if len(file.Comments) > 0 {
			firstComment := file.Comments[0]
			if strings.HasPrefix(firstComment.Text(), "Code generated by 'go test'. DO NOT EDIT") {
				// skip generated files
				continue
			}
		}
		if file.Name.Name == "main" {
			ast.Inspect(file, func(node ast.Node) bool {
				if fnDecl, ok := node.(*ast.FuncDecl); ok {
					expr(fnDecl)
				}
				return true
			})
		}
	}
	return nil, nil

}

func isOsExitCall(call *ast.CallExpr) bool {

	if selExpr, ok := call.Fun.(*ast.SelectorExpr); ok {
		if packIdent, isIdent := selExpr.X.(*ast.Ident); !isIdent || packIdent.Name != "os" {
			return false
		}

		if selExpr.Sel.Name != "Exit" {
			return false
		}
		return true
	}

	return false
}