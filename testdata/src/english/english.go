package english

import "log/slog"

func test() {
	slog.Info("запуск сервера")      // want "log message must contain only English letters" "log message must not contain special characters or emoji"
	slog.Error("ошибка подключения") // want "log message must contain only English letters" "log message must not contain special characters or emoji"
	slog.Warn("完成")                  // want "log message must start with a lowercase letter" "log message must contain only English letters" "log message must not contain special characters or emoji"
}
