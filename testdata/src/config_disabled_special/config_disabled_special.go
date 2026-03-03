package config_disabled_special

import "log/slog"

func test() {
	slog.Info("warning: pending")
	slog.Info("server started!")
}
