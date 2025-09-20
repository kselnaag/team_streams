package types

type ILog interface {
	Start() func()
	LogTrace(format string, v ...any)
	LogDebug(format string, v ...any)
	LogInfo(format string, v ...any)
	LogWarn(format string, v ...any)
	LogError(err error)
	LogFatal(err error)
	LogPanic(err error)
}
