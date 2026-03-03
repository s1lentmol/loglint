package specialchars

import "log/slog"

func test() {
	slog.Info("server started!")   // want "log message must not contain special characters or emoji"
	slog.Error("warning: pending") // want "log message must not contain special characters or emoji"
	slog.Warn("something...")      // want "log message must not contain special characters or emoji"
	slog.Info("slog.Info")         // want "log message must not contain special characters or emoji"
	slog.Info("done 🚀")            // want "log message must not contain special characters or emoji"
}
