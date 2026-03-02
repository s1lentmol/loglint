package zap

type Logger struct{}

type SugaredLogger struct{}

func NewNop() *Logger {
	return &Logger{}
}

func L() *Logger {
	return &Logger{}
}

func S() *SugaredLogger {
	return &SugaredLogger{}
}

func (l *Logger) Info(msg string, fields ...any) {}

func (l *Logger) Named(name string) *Logger {
	return l
}

func (l *SugaredLogger) Info(args ...any) {}
