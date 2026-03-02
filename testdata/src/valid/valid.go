package valid

import (
	"log/slog"

	"go.uber.org/zap"
)

func test() {
	slog.Info("starting server")
	slog.Info(" 123 ready")

	logger := zap.NewNop()
	logger.Info("request completed")

	sugar := zap.S()
	sugar.Info("request completed")
}
