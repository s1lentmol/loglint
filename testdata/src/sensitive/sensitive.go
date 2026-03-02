package sensitive

import "log/slog"

func test(password, apiKey string, user struct{ Token string }) {
	slog.Info("password rotated")                // want "log message must not contain potential sensitive data"
	slog.Info("request completed " + apiKey)     // want "log message must not contain potential sensitive data"
	slog.Info("request completed " + user.Token) // want "log message must not contain potential sensitive data"
	slog.Info("auth completed")                  // want "log message must not contain potential sensitive data"
	slog.Info("request completed")

	_ = password
}
