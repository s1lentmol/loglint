package mixed

import (
	"log/slog"

	"go.uber.org/zap"
)

func test(token string) {
	slog.Info("Password: пароль!!!") // want "log message must start with a lowercase letter" "log message must contain only English letters" "log message must not contain special characters or emoji" "log message must not contain potential sensitive data"
	slog.Info("token: " + token)     // want "log message must not contain special characters or emoji" "log message must not contain potential sensitive data"

	logger := zap.NewNop()
	logger.Info("Secret leaked!") // want "log message must start with a lowercase letter" "log message must not contain special characters or emoji" "log message must not contain potential sensitive data"

	sugar := zap.S()
	sugar.Info("Done 🚀") // want "log message must start with a lowercase letter" "log message must not contain special characters or emoji"
}
