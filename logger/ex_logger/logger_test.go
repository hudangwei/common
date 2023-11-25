package ex_logger

import (
	"testing"
)

func ImportantWarp(v ...interface{}) {
	Important(v...)
}

func TestLog(t *testing.T) {
	SetRollingFile(".", "server.log", 1000, 100, MB)
	SetFlag(LstdFlags | Lmicroseconds)
	SetConsole(false)
	SetLevel(ALL)

	ImportantWarp("hahah", 1, 2, 3, "ddddd")
}
