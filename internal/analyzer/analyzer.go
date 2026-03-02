package analyzer

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "loglint",
	Doc:  "checks log message quality in Go code",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		loggerAliases := loggerImportAliases(file)
		if len(loggerAliases) == 0 {
			continue
		}

		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			_ = isLoggerCall(call, loggerAliases)
			return true
		})
	}

	return nil, nil
}

func loggerImportAliases(file *ast.File) map[string]struct{} {
	aliases := make(map[string]struct{})

	for _, imp := range file.Imports {
		path := strings.Trim(imp.Path.Value, `"`)

		switch path {
		case "log/slog":
			addImportAlias(aliases, imp, "slog")
		case "go.uber.org/zap":
			addImportAlias(aliases, imp, "zap")
		}
	}

	return aliases
}

func addImportAlias(dst map[string]struct{}, imp *ast.ImportSpec, fallback string) {
	if imp.Name == nil {
		dst[fallback] = struct{}{}
		return
	}

	if imp.Name.Name == "." || imp.Name.Name == "_" {
		return
	}

	dst[imp.Name.Name] = struct{}{}
}

func isLoggerCall(call *ast.CallExpr, loggerAliases map[string]struct{}) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	pkgIdent, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}

	_, ok = loggerAliases[pkgIdent.Name]
	return ok
}
