package basic

import (
	sl "log/slog"

	z "go.uber.org/zap"
)

type fakeLogger struct{}

func (fakeLogger) Info(msg string, args ...any) {}

func validCases() {
	sl.Info("starting server")

	logger := z.NewNop()
	logger.Info("request completed")

	sugar := z.S()
	sugar.Info("request completed")
}

func invalidCases(password, apiKey, token string, user struct{ Token string }) {
	sl.Error("Failed to connect")
	sl.Info("запуск сервера")
	sl.Info("server started!")
	sl.Info("user password: " + password)
	sl.Info("api_key=" + apiKey)
	sl.Info("token: " + token)
	sl.Info("Password: пароль!!!")

	logger := z.NewNop()
	logger.Info("Secret leaked")

	sugar := z.S()
	sugar.Info("done 🚀")

	sl.Info("user token " + user.Token)
}

func edgeCases() {
	sl := fakeLogger{}
	sl.Info("not a real logger")
}
