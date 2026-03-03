package analyzer

import (
	"go/parser"
	"testing"
)

func TestCheckLowercaseStart(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		msg  string
		want bool
	}{
		{name: "lowercase", msg: "starting server", want: true},
		{name: "uppercase", msg: "Starting server", want: false},
		{name: "leading spaces", msg: "  starting server", want: true},
		{name: "leading spaces uppercase", msg: "   Server started", want: false},
		{name: "starts with digit", msg: "123 started", want: true},
		{name: "empty", msg: "", want: true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := checkLowercaseStart(tc.msg)
			if got != tc.want {
				t.Fatalf("checkLowercaseStart(%q) = %v, want %v", tc.msg, got, tc.want)
			}
		})
	}
}

func TestCheckEnglishOnly(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		msg  string
		want bool
	}{
		{name: "letters and spaces", msg: "starting server", want: true},
		{name: "letters and digits", msg: "server 8080", want: true},
		{name: "dot punctuation", msg: "slog.Info", want: true},
		{name: "colon punctuation", msg: "warning: pending", want: true},
		{name: "cyrillic", msg: "запуск сервера", want: false},
		{name: "cjk", msg: "完成", want: false},
		{name: "emoji with english", msg: "done 🚀", want: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := checkEnglishOnly(tc.msg)
			if got != tc.want {
				t.Fatalf("checkEnglishOnly(%q) = %v, want %v", tc.msg, got, tc.want)
			}
		})
	}
}

func TestCheckNoSpecialCharsOrEmoji(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		msg  string
		want bool
	}{
		{name: "letters and spaces", msg: "request completed", want: true},
		{name: "digits", msg: "server 8080", want: true},
		{name: "exclamation", msg: "failed!", want: false},
		{name: "ellipsis", msg: "something...", want: false},
		{name: "underscore", msg: "api_key", want: false},
		{name: "emoji", msg: "done 🚀", want: false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := checkNoSpecialCharsOrEmoji(tc.msg)
			if got != tc.want {
				t.Fatalf("checkNoSpecialCharsOrEmoji(%q) = %v, want %v", tc.msg, got, tc.want)
			}
		})
	}
}

func TestHasSensitiveKeyword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want bool
	}{
		{name: "password", in: "password", want: true},
		{name: "apikey mixed case", in: "apiKey", want: true},
		{name: "credential substring", in: "credentials", want: true},
		{name: "safe text", in: "request completed", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := hasSensitiveKeyword(tc.in, defaultSensitiveKeywords)
			if got != tc.want {
				t.Fatalf("hasSensitiveKeyword(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestCheckNoSensitiveData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		msgText string
		expr    string
		want    bool
	}{
		{
			name:    "safe literal",
			msgText: "request completed",
			expr:    `"request completed"`,
			want:    true,
		},
		{
			name:    "keyword in text",
			msgText: "password rotated",
			expr:    `"password rotated"`,
			want:    false,
		},
		{
			name:    "keyword in ident",
			msgText: "request completed <expr>",
			expr:    `"request completed " + apiKey`,
			want:    false,
		},
		{
			name:    "keyword in selector",
			msgText: "request completed <expr>",
			expr:    `"request completed " + user.Token`,
			want:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			expr, err := parser.ParseExpr(tc.expr)
			if err != nil {
				t.Fatalf("parse expr %q: %v", tc.expr, err)
			}

			got := checkNoSensitiveData(tc.msgText, expr, defaultSensitiveKeywords)
			if got != tc.want {
				t.Fatalf("checkNoSensitiveData(%q, %q) = %v, want %v", tc.msgText, tc.expr, got, tc.want)
			}
		})
	}
}

func TestResolveSensitiveKeywords(t *testing.T) {
	t.Parallel()

	appendMode, err := resolveSensitiveKeywords("append", []string{"sessionid"})
	if err != nil {
		t.Fatalf("append mode: %v", err)
	}
	if !hasSensitiveKeyword("sessionid", appendMode) || !hasSensitiveKeyword("password", appendMode) {
		t.Fatalf("append mode did not merge defaults and custom")
	}

	overrideMode, err := resolveSensitiveKeywords("override", []string{"sessionid"})
	if err != nil {
		t.Fatalf("override mode: %v", err)
	}
	if !hasSensitiveKeyword("sessionid", overrideMode) {
		t.Fatalf("override mode missing custom keyword")
	}
	if hasSensitiveKeyword("password", overrideMode) {
		t.Fatalf("override mode should drop default keywords")
	}
}

func TestNormalizeMessageText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		expr string
		want string
		ok   bool
	}{
		{name: "string literal", expr: `"hello"`, want: "hello", ok: true},
		{name: "concat with ident", expr: `"token " + token`, want: "token <expr>", ok: true},
		{name: "concat literals", expr: `"a" + "b"`, want: "ab", ok: true},
		{name: "unsupported top-level ident", expr: "msg", want: "", ok: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			expr, err := parser.ParseExpr(tc.expr)
			if err != nil {
				t.Fatalf("parse expr %q: %v", tc.expr, err)
			}

			got, ok := normalizeMessageText(expr, false)
			if ok != tc.ok || got != tc.want {
				t.Fatalf("normalizeMessageText(%q) = (%q, %v), want (%q, %v)", tc.expr, got, ok, tc.want, tc.ok)
			}
		})
	}
}

func TestStripDynamicMarkers(t *testing.T) {
	t.Parallel()

	got := stripDynamicMarkers("user <expr> token <expr>")
	if got != "user  token " {
		t.Fatalf("stripDynamicMarkers() = %q, want %q", got, "user  token ")
	}
}
