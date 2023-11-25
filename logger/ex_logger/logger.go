package ex_logger

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	_VER string = "1.0.0"
)

type LEVEL int32

var logLevel LEVEL = 1
var maxFileSize int64
var maxFileCount int32
var dailyRolling bool = true
var consoleAppender bool = true
var RollingFile bool = false
var logObj *_FILE

const DATEFORMAT = "2006-01-02"

type UNIT int64

const (
	_       = iota
	KB UNIT = 1 << (iota * 10)
	MB
	GB
	TB
)

const (
	ALL LEVEL = iota
	DEBUG
	INFO
	IMPORTANT
	WARN
	ERROR
	FATAL
	SYS
	OFF
)

//type ExLog struct {
//	*log.Logger
//}
//
//// to output function name
//func (l *ExLog) Output(calldepth int, s string) error {
//	now := time.Now() // get this early.
//	var file string
//	var line int
//	l.mu.Lock()
//	defer l.mu.Unlock()
//	if l.flag&(Lshortfile|Llongfile) != 0 {
//		// release lock while getting caller info - it's expensive.
//		l.mu.Unlock()
//		var ok bool
//		_, file, line, ok = runtime.Caller(calldepth)
//		if !ok {
//			file = "???"
//			line = 0
//		}
//		l.mu.Lock()
//	}
//	l.buf = l.buf[:0]
//	l.formatHeader(&l.buf, now, file, line)
//	l.buf = append(l.buf, s...)
//	if len(s) == 0 || s[len(s)-1] != '\n' {
//		l.buf = append(l.buf, '\n')
//	}
//	_, err := l.out.Write(l.buf)
//	return err
//}
//
//func NewExlog(out io.Writer, prefix string, flag int) *ExLog {
//	return &ExLog{Logger:log.New(out, prefix, flag)}
//}

func (l LEVEL) String() string {
	switch l {
	case ALL:
		return "all"
	case DEBUG:
		return "debug"
	case INFO:
		return "info"
	case IMPORTANT:
		return "important"
	case WARN:
		return "warn"
	case ERROR:
		return "error"
	case FATAL:
		return "fatal"
	case SYS:
		return "sys"
	case OFF:
		return "off"
	default:
		return fmt.Sprintf("Level(%d)", l)
	}
}

type _FILE struct {
	dir      string
	filename string
	_suffix  int
	isCover  bool
	_date    *time.Time
	mu       *sync.RWMutex
	logfile  *os.File
	lg       *ExLogger
}

func SetConsole(isConsole bool) {
	consoleAppender = isConsole
}

func SetLevel(_level LEVEL) {
	logLevel = _level
}

func LogLevel() LEVEL {
	return logLevel
}

func GetLevel() string {
	return logLevel.String()
}

func StrToLEVEL(level string) LEVEL {
	var lv LEVEL
	switch level {
	case "all", "ALL":
		lv = ALL
	case "debug", "DEBUG":
		lv = DEBUG
	case "info", "INFO":
		lv = INFO
	case "important", "IMPORTANT":
		lv = IMPORTANT
	case "warn", "WARN":
		lv = WARN
	case "error", "ERROR":
		lv = ERROR
	case "fatal", "FATAL":
		lv = FATAL
	case "sys", "SYS":
		lv = SYS
	case "off", "OFF":
		lv = OFF
	}
	return lv
}

func SetFlag(_flag int) {
	logObj.mu.Lock()
	defer logObj.mu.Unlock()
	logObj.lg.SetFlags(_flag)
}

func SetRollingFile(fileDir, fileName string, maxNumber int32, maxSize int64, _unit UNIT) {
	maxFileCount = maxNumber
	maxFileSize = maxSize * int64(_unit)
	RollingFile = true
	dailyRolling = false

	if strings.HasSuffix(fileName, ".log") {
		fileName = fileName[0 : len(fileName)-4]
	}

	logObj = &_FILE{dir: fileDir, filename: fileName, isCover: false, mu: new(sync.RWMutex)}
	logObj.mu.Lock()
	defer logObj.mu.Unlock()

	for i := 1; i <= int(maxNumber); i++ {
		if isExist(fileDir + "/" + fileName + "." + strconv.Itoa(i) + ".log") {
			logObj._suffix = i
		} else {
			break
		}
	}
	if !logObj.isMustRename() {
		var err error
		logObj.logfile, err = os.OpenFile(fileDir+"/"+fileName+".log", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			logObj.logfile.Chmod(644)
		}
		old := logObj.lg
		if old != nil {
			old.Close()
		}
		logObj.lg = NewExLogger(logObj.logfile, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	} else {
		logObj.rename()
	}
	go fileMonitor()
}

func SetRollingDaily(fileDir, fileName string) {
	RollingFile = false
	dailyRolling = true
	t, _ := time.Parse(DATEFORMAT, time.Now().Format(DATEFORMAT))

	if strings.HasSuffix(fileName, ".log") {
		fileName = fileName[0 : len(fileName)-4]
	}

	logObj = &_FILE{dir: fileDir, filename: fileName, _date: &t, isCover: false, mu: new(sync.RWMutex)}
	logObj.mu.Lock()
	defer logObj.mu.Unlock()

	if !logObj.isMustRename() {
		var err error
		logObj.logfile, err = os.OpenFile(fileDir+"/"+fileName+".log", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			logObj.logfile.Chmod(644)
		}
		old := logObj.lg
		if old != nil {
			old.Close()
		}
		logObj.lg = NewExLogger(logObj.logfile, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
	} else {
		logObj.rename()
	}
}

func console(stack int, s ...interface{}) {
	if consoleAppender {
		pc, file, line, _ := runtime.Caller(stack)
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		f := runtime.FuncForPC(pc)
		file = short
		fname := strings.Split(f.Name(), "/")
		if len(fname) > 1 {
			log.Println(file+":"+strconv.Itoa(line), fname[len(fname)-1], s)
		} else {
			log.Println(file+":"+strconv.Itoa(line), f.Name(), s)
		}
	}
}

func catchError() {
	if err := recover(); err != nil {
		log.Println("err", err)
	}
}

func Debug(v ...interface{}) {
	DebugS(DEBUG, 3, v)
}

func DebugS(level LEVEL, stack int, v ...interface{}) {
	//if logLevel > DEBUG {
	//	return
	//}
	if level > DEBUG {
		return
	}
	if dailyRolling {
		fileCheck()
	}
	defer catchError()
	logObj.mu.RLock()
	defer logObj.mu.RUnlock()

	lg := logObj.lg
	lg.Output(stack, fmt.Sprintln("debug", v))
	console(stack, "debug", v)
}

func Info(v ...interface{}) {
	InfoS(INFO, 3, v)
}

func InfoS(level LEVEL, stack int, v ...interface{}) {
	//if logLevel > INFO {
	//	return
	//}
	if level > INFO {
		return
	}
	if dailyRolling {
		fileCheck()
	}
	defer catchError()
	logObj.mu.RLock()
	defer logObj.mu.RUnlock()

	lg := logObj.lg
	lg.Output(stack, fmt.Sprintln("info", v))
	console(stack, "info", v)
}

func Warn(v ...interface{}) {
	WarnS(WARN, 3, v)
}

func WarnS(level LEVEL, stack int, v ...interface{}) {
	//if logLevel > WARN {
	//	return
	//}
	if level > WARN {
		return
	}
	if dailyRolling {
		fileCheck()
	}
	defer catchError()
	logObj.mu.RLock()
	defer logObj.mu.RUnlock()

	lg := logObj.lg
	lg.Output(stack, fmt.Sprintln("warn", v))
	console(stack, "warn", v)
}

func Error(v ...interface{}) {
	ErrorS(ERROR, 3, v)
}

func ErrorS(level LEVEL, stack int, v ...interface{}) {
	//if logLevel > ERROR {
	//	return
	//}
	if level > ERROR {
		return
	}
	if dailyRolling {
		fileCheck()
	}
	defer catchError()
	logObj.mu.RLock()
	defer logObj.mu.RUnlock()

	lg := logObj.lg
	lg.Output(stack, fmt.Sprintln("error", v))
	console(stack, "error", v)
}

func Important(v ...interface{}) {
	ImportantS(IMPORTANT, 3, v...)
}

func ImportantS(level LEVEL, stack int, v ...interface{}) {
	//if logLevel > IMPORTANT {
	//	return
	//}
	if level > IMPORTANT {
		return
	}

	if dailyRolling {
		fileCheck()
	}
	defer catchError()
	logObj.mu.RLock()
	defer logObj.mu.RUnlock()

	lg := logObj.lg
	lg.Output(stack, fmt.Sprintln(v...))
	console(stack, "important", v)
}

func Fatal(v ...interface{}) {
	FatalS(FATAL, 3, v)
}

func FatalS(level LEVEL, stack int, v ...interface{}) {
	//if logLevel > FATAL {
	//	return
	//}
	if level > FATAL {
		return
	}

	if dailyRolling {
		fileCheck()
	}
	defer catchError()
	logObj.mu.RLock()
	defer logObj.mu.RUnlock()

	lg := logObj.lg
	lg.Output(stack, fmt.Sprintln("fatal", v))
	console(stack, "fatal", v)

}

func Sys(v ...interface{}) {
	SysS(SYS, 3, v)
}

func SysS(level LEVEL, stack int, v ...interface{}) {
	//if logLevel > SYS {
	//	return
	//}
	if level > SYS {
		return
	}

	if dailyRolling {
		fileCheck()
	}
	defer catchError()
	logObj.mu.RLock()
	defer logObj.mu.RUnlock()

	lg := logObj.lg
	lg.Output(stack, fmt.Sprint("sys", v))
	console(stack, "sys", v)
}

func (f *_FILE) isMustRename() bool {
	if dailyRolling {
		t, _ := time.Parse(DATEFORMAT, time.Now().Format(DATEFORMAT))
		if t.After(*f._date) {
			return true
		}
	} else {
		if maxFileCount > 1 {
			if fileSize(f.dir+"/"+f.filename+".log") >= maxFileSize {
				return true
			}
		}
	}
	return false
}

func (f *_FILE) rename() {
	if dailyRolling {
		fn := f.dir + "/" + f.filename + "." + f._date.Format(DATEFORMAT)
		if !isExist(fn) && f.isMustRename() {
			if f.logfile != nil {
				f.logfile.Close()
			}
			err := os.Rename(f.dir+"/"+f.filename, fn)
			if err != nil {
				f.lg.Println("rename err", err.Error())
			}
			t, _ := time.Parse(DATEFORMAT, time.Now().Format(DATEFORMAT))
			f._date = &t
			f.logfile, _ = os.Create(f.dir + "/" + f.filename)

			old := logObj.lg
			if old != nil {
				old.Close()
			}
			f.lg = NewExLogger(logObj.logfile, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
		}
	} else {
		f.coverNextOne()
	}
}

func (f *_FILE) nextSuffix() int {
	return int(f._suffix%int(maxFileCount) + 1)
}

func (f *_FILE) convertOld() {
	max_file := f.dir + "/" + f.filename + "." + strconv.Itoa(int(maxFileCount)) + ".log"
	if isExist(max_file) {
		os.Remove(max_file)
	}
	for i := maxFileCount; i > 0; i-- {
		t_file := f.dir + "/" + f.filename + "." + strconv.Itoa(int(i)) + ".log"
		n_file := f.dir + "/" + f.filename + "." + strconv.Itoa(int(i)+1) + ".log"

		if isExist(t_file) {
			os.Rename(t_file, n_file)
		}
	}
	file_name := f.dir + "/" + f.filename + ".log"
	if isExist(file_name) {
		new_file := f.dir + "/" + f.filename + ".1" + ".log"
		os.Rename(file_name, new_file)
	}
}

func (f *_FILE) coverNextOne() {
	f._suffix = f.nextSuffix()
	if f.logfile != nil {
		f.logfile.Close()
	}
	file_name := f.dir + "/" + f.filename + ".log"
	var err error
	f.logfile, err = os.Create(file_name)
	if err != nil {
		f.logfile.Chmod(644)
	}
	old := logObj.lg
	if old != nil {
		old.Close()
	}
	f.lg = NewExLogger(logObj.logfile, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)
}

func fileSize(file string) int64 {
	f, e := os.Stat(file)
	if e != nil {
		fmt.Println(e.Error())
		return 0
	}
	return f.Size()
}

func isExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func fileMonitor() {
	timer := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-timer.C:
			fileCheck()
		}
	}
}

func fileCheck() {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()
	if logObj != nil && logObj.isMustRename() {
		logObj.convertOld()
		logObj.mu.Lock()
		defer logObj.mu.Unlock()
		logObj.rename()
	}
}
