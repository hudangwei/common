package app

import (
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	_sigs       = make(chan os.Signal, 1)
	_signalFunc = make(map[os.Signal]func())
)

func RegSignalFunc(sig os.Signal, f func()) {
	signal.Notify(_sigs, sig)
	_signalFunc[sig] = f
}

func Wait() {
	signal.Notify(_sigs, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGQUIT)

	callfunc := func(sig os.Signal) {
		if fu, ok := _signalFunc[sig]; ok {
			fu()
		}
	}

	for {
		msg := <-_sigs
		switch msg {
		default:
			callfunc(msg)
		case os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGQUIT:
			callfunc(msg)
			time.Sleep(time.Second)
			return
		}
	}
}
