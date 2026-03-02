package edgecases

import (
	sl "log/slog"
	"strings"

	z "go.uber.org/zap"
)

type fakeLogger struct{}

func (fakeLogger) Info(msg string, args ...any) {}

func test() {
	sl.Info("Failed run") // want "log message must start with a lowercase letter"

	logger := z.NewNop()
	logger.Info(strings.TrimSpace("request completed"))
	logger.Named("etcd-client")

	sugar := z.S()
	sugar.Info("session started")

	notLogger := fakeLogger{}
	notLogger.Info("Password leaked!")
}
