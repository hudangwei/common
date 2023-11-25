package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// error logger
var errorLogger *zap.Logger
var logLevel = zap.NewAtomicLevel()

var levelMap = map[string]zapcore.Level{
	"debug":  zapcore.DebugLevel,
	"info":   zapcore.InfoLevel,
	"warn":   zapcore.WarnLevel,
	"error":  zapcore.ErrorLevel,
	"dpanic": zapcore.DPanicLevel,
	"panic":  zapcore.PanicLevel,
	"fatal":  zapcore.FatalLevel,
}

func GetLoggerLevel() string {
	return logLevel.String()
}

func SetLoggerLevel(lv string) {
	logLevel.SetLevel(getLoggerLevel(lv))
}

func getLoggerLevel(lvl string) zapcore.Level {
	if level, ok := levelMap[lvl]; ok {
		return level
	}
	return zapcore.InfoLevel
}

func init() {
	// Init("./logs", "server.log", "server_elk.log", "debug")
	Init("./logs", "server.log", "debug")
}

func Init(logPath, fileName string, level string) {
	// syncWriter := zapcore.AddSync(&lumberjack.Logger{
	// 	Filename:   fileName,
	// 	MaxSize:    100 << 20, //100mb
	// 	MaxBackups: 999,       // 最多保留999个备份
	// 	LocalTime:  true,
	// })

	encoder := zap.NewProductionEncoderConfig()
	encoder.EncodeTime = zapcore.ISO8601TimeEncoder
	logOut := NewRotateFile(logPath, fileName, 20, 100000, 0)
	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoder), zapcore.AddSync(logOut), logLevel)
	logLevel.SetLevel(getLoggerLevel(level))
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	errorLogger = logger

	//ex_logger
	// ex_logger.SetRollingFile(logPath, otherFileName, 1000, 100, ex_logger.MB)
	// ex_logger.SetFlag(ex_logger.LstdFlags | ex_logger.Lmicroseconds)
	// ex_logger.SetConsole(false)
	// ex_logger.SetLevel(ex_logger.ALL)
}

func Debug(msg string, fields ...zap.Field) {
	errorLogger.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	errorLogger.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	errorLogger.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	errorLogger.Error(msg, fields...)
}

func Panic(msg string, fields ...zap.Field) {
	errorLogger.Panic(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	errorLogger.Fatal(msg, fields...)
}

// func Important(v ...interface{}) {
// 	ex_logger.Important(v)
// }
