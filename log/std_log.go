package stdlog

import "github.com/uber-go/zap"

type StdLogger struct {
	Logger zap.Logger
	Level  zap.Level
}

func NewStdLogger(logger zap.Logger, level zap.Level, key, value string) *StdLogger{
	stdLogger := logger.With(zap.String(key, value))

	return &StdLogger{
		Logger:  stdLogger,
		Level: level,
	}
}
func (log *StdLogger) Write(p []byte) (n int, err error){
	switch log.Level {
	case zap.DebugLevel:
		log.Logger.Debug(string(p[:]))
	case zap.InfoLevel:
		log.Logger.Info(string(p[:]))
	case zap.WarnLevel:
		log.Logger.Warn(string(p[:]))
	case zap.ErrorLevel:
		log.Logger.Error(string(p[:]))
	default:
		log.Logger.Debug(string(p[:]))
	}

	return len(p), nil
}