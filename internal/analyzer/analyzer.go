package analyzer

import (
	"go/ast"
	"go/token"
	"go/types"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"
)

const (
	diagLowercaseStart = "log message must start with a lowercase letter"
	diagEnglishOnly    = "log message must contain only English letters"
	diagNoSpecial      = "log message must not contain special characters or emoji"
	diagNoSensitive    = "log message must not contain potential sensitive data"
)

var sensitiveKeywords = []string{
	"password",
	"passwd",
	"pwd",
	"token",
	"secret",
	"api_key",
	"apikey",
	"auth",
	"credential",
}

var (
	slogMethods = map[string]struct{}{
		"Debug": {},
		"Info":  {},
		"Warn":  {},
		"Error": {},
	}
	zapLoggerMethods = map[string]struct{}{
		"Debug":  {},
		"Info":   {},
		"Warn":   {},
		"Error":  {},
		"DPanic": {},
		"Panic":  {},
		"Fatal":  {},
	}
	zapSugaredMethods = map[string]struct{}{
		"Debug":   {},
		"Info":    {},
		"Warn":    {},
		"Error":   {},
		"DPanic":  {},
		"Panic":   {},
		"Fatal":   {},
		"Debugf":  {},
		"Infof":   {},
		"Warnf":   {},
		"Errorf":  {},
		"DPanicf": {},
		"Panicf":  {},
		"Fatalf":  {},
	}
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

			msgExpr, ok := extractLogMessageExpr(pass, call, loggerAliases)
			if !ok {
				return true
			}

			msgText, ok := normalizeMessageText(msgExpr, false)
			if !ok {
				return true
			}
			ruleText := stripDynamicMarkers(msgText)

			pos := msgExpr.Pos()
			if pos == token.NoPos {
				pos = call.Pos()
			}

			if !checkLowercaseStart(ruleText) {
				pass.Reportf(pos, diagLowercaseStart)
			}
			if !checkEnglishOnly(ruleText) {
				pass.Reportf(pos, diagEnglishOnly)
			}
			if !checkNoSpecialCharsOrEmoji(ruleText) {
				pass.Reportf(pos, diagNoSpecial)
			}
			if !checkNoSensitiveData(msgText, msgExpr) {
				pass.Reportf(pos, diagNoSensitive)
			}

			return true
		})
	}

	return nil, nil
}

func loggerImportAliases(file *ast.File) map[string]string {
	aliases := make(map[string]string)

	for _, imp := range file.Imports {
		path := strings.Trim(imp.Path.Value, `"`)

		switch path {
		case "log/slog":
			addImportAlias(aliases, imp, "slog", path)
		case "go.uber.org/zap":
			addImportAlias(aliases, imp, "zap", path)
		}
	}

	return aliases
}

func addImportAlias(dst map[string]string, imp *ast.ImportSpec, fallback, path string) {
	if imp.Name == nil {
		dst[fallback] = path
		return
	}

	if imp.Name.Name == "." || imp.Name.Name == "_" {
		return
	}

	dst[imp.Name.Name] = path
}

func extractLogMessageExpr(pass *analysis.Pass, call *ast.CallExpr, loggerAliases map[string]string) (ast.Expr, bool) {
	if len(call.Args) == 0 {
		return nil, false
	}

	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil, false
	}

	if pkgIdent, ok := sel.X.(*ast.Ident); ok {
		if isImportedPkgIdent(pass, pkgIdent, loggerAliases, "log/slog") && isAllowedSlogMethod(sel.Sel.Name) {
			return call.Args[0], true
		}
	}

	zapType, ok := zapLoggerType(pass, sel.X)
	if !ok || !isAllowedZapMethod(sel.Sel.Name, zapType) {
		return nil, false
	}

	return call.Args[0], true
}

func isImportedPkgIdent(pass *analysis.Pass, ident *ast.Ident, aliases map[string]string, expectedPath string) bool {
	path, ok := aliases[ident.Name]
	if !ok || path != expectedPath {
		return false
	}

	obj, ok := pass.TypesInfo.Uses[ident]
	if !ok {
		return false
	}

	pkgObj, ok := obj.(*types.PkgName)
	if !ok || pkgObj.Imported() == nil {
		return false
	}

	return pkgObj.Imported().Path() == expectedPath
}

func zapLoggerType(pass *analysis.Pass, expr ast.Expr) (string, bool) {
	tv, ok := pass.TypesInfo.Types[expr]
	if !ok || tv.Type == nil {
		return "", false
	}

	typ := tv.Type
	if ptr, ok := typ.(*types.Pointer); ok {
		typ = ptr.Elem()
	}

	named, ok := typ.(*types.Named)
	if !ok || named.Obj() == nil || named.Obj().Pkg() == nil {
		return "", false
	}

	pkgPath := named.Obj().Pkg().Path()
	typeName := named.Obj().Name()

	if pkgPath != "go.uber.org/zap" {
		return "", false
	}

	if typeName != "Logger" && typeName != "SugaredLogger" {
		return "", false
	}

	return typeName, true
}

func isAllowedSlogMethod(method string) bool {
	_, ok := slogMethods[method]
	return ok
}

func isAllowedZapMethod(method, loggerType string) bool {
	switch loggerType {
	case "Logger":
		_, ok := zapLoggerMethods[method]
		return ok
	case "SugaredLogger":
		_, ok := zapSugaredMethods[method]
		return ok
	default:
		return false
	}
}

func normalizeMessageText(expr ast.Expr, inConcat bool) (string, bool) {
	switch e := expr.(type) {
	case *ast.BasicLit:
		if e.Kind != token.STRING {
			if inConcat {
				return "<expr>", true
			}
			return "", false
		}

		unquoted, err := strconv.Unquote(e.Value)
		if err != nil {
			return "", false
		}

		return unquoted, true
	case *ast.BinaryExpr:
		if e.Op != token.ADD {
			if inConcat {
				return "<expr>", true
			}
			return "", false
		}

		left, ok := normalizeMessageText(e.X, true)
		if !ok {
			return "", false
		}
		right, ok := normalizeMessageText(e.Y, true)
		if !ok {
			return "", false
		}

		return left + right, true
	default:
		if inConcat {
			return "<expr>", true
		}
		return "", false
	}
}

func checkLowercaseStart(msg string) bool {
	for _, r := range msg {
		if unicode.IsSpace(r) {
			continue
		}

		if unicode.IsDigit(r) {
			return true
		}
		if unicode.IsLetter(r) {
			return unicode.IsLower(r)
		}

		return true
	}

	return true
}

func checkEnglishOnly(msg string) bool {
	for _, r := range msg {
		if !unicode.IsLetter(r) {
			continue
		}

		if !isASCIIAlpha(r) {
			return false
		}
	}
	return true
}

func checkNoSpecialCharsOrEmoji(msg string) bool {
	for _, r := range msg {
		if isASCIIAlphaNumSpace(r) {
			continue
		}
		return false
	}
	return true
}

func isASCIIAlphaNumSpace(r rune) bool {
	return r == ' ' ||
		(r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9')
}

func isASCIIAlpha(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func checkNoSensitiveData(msgText string, msgExpr ast.Expr) bool {
	if hasSensitiveKeyword(msgText) {
		return false
	}

	found := false
	ast.Inspect(msgExpr, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.Ident:
			if hasSensitiveKeyword(x.Name) {
				found = true
				return false
			}
		case *ast.SelectorExpr:
			if hasSensitiveKeyword(x.Sel.Name) {
				found = true
				return false
			}
		}
		return !found
	})

	return !found
}

func hasSensitiveKeyword(s string) bool {
	lower := strings.ToLower(s)
	for _, kw := range sensitiveKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

func stripDynamicMarkers(msg string) string {
	return strings.ReplaceAll(msg, "<expr>", "")
}
