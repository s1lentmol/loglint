package lowercase

import "log/slog"

func test() {
	slog.Info("Starting server")    // want "log message must start with a lowercase letter"
	slog.Error(" Failed to launch") // want "log message must start with a lowercase letter"
	slog.Warn("X value updated")    // want "log message must start with a lowercase letter"

	slog.Info("123 started")
	slog.Info("   404 handled")
}
