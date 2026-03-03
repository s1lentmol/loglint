package config_sensitive_override

import "log/slog"

func test() {
	slog.Info("password rotated")
	slog.Info("sessionid rotated") // want "log message must not contain potential sensitive data"
}
