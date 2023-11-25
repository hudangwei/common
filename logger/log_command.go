package logger

import (
	"fmt"
)

type LogCommand struct{}

func (l *LogCommand) Name() string {
	return "logger"
}

func (l *LogCommand) Help() string {
	return fmt.Sprintln("logger debug|info|warn|error|dpanic|panic|fatal|now")
}

func (l *LogCommand) Call(args []string) string {
	if len(args) == 0 {
		return l.Help()
	}

	level := args[0]
	if level == "now" {
		return fmt.Sprintln("current log level:", GetLoggerLevel())
	}
	SetLoggerLevel(level)
	return fmt.Sprintln("SetLogLevel", level, "NewLevel", GetLoggerLevel())
}
