package app

import (
	"flag"
	"fmt"
	"io/ioutil"
	"reflect"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/hudangwei/common/depends"
	"github.com/hudangwei/common/logger"
	"go.uber.org/zap"
)

type AppSetting struct {
	ConfigPath string
}

var defaultSetting = AppSetting{}

func init() {
	flag.StringVar(&defaultSetting.ConfigPath, "c", "/etc/server.toml", "server config file")
}

type GlobalInfo struct {
	LogLevel string `toml:"log_level"`
}
type App struct {
	configData string
}

func NewApp() *App {
	flag.Parse()
	app := &App{}
	// 读取配置文件
	bs, _ := ioutil.ReadFile(defaultSetting.ConfigPath)
	app.configData = string(bs)

	// 解析global info
	if globalInfo, err := app.LoadConfig(&GlobalInfo{}, "global"); err == nil && len(globalInfo.(*GlobalInfo).LogLevel) > 0 {
		logger.SetLoggerLevel(globalInfo.(*GlobalInfo).LogLevel)
	}

	depends.RangeIOModule(func(name string, module depends.IOModule) bool {
		err := module.Open(app, name)
		if err != nil {
			logger.Panic("IOModule Open", zap.String("module name", name), zap.Error(err))
			return false
		}
		return true
	})

	return app
}

func (app *App) LoadConfig(config interface{}, name string) (interface{}, error) {
	section := reflect.New(reflect.StructOf([]reflect.StructField{reflect.StructField{
		Name: "Conf",
		Type: reflect.TypeOf(config),
		Tag:  reflect.StructTag(fmt.Sprintf("toml:\"%s\"", name)),
	}})).Interface()
	_, err := toml.Decode(app.configData, section)
	if err != nil {
		return nil, err
	}
	return reflect.Indirect(reflect.ValueOf(section)).FieldByName("Conf").Interface(), nil
}

func (app *App) Run(fn func()) {
	exit := func() {
		if fn != nil {
			fn()
		}
	}
	RegSignalFunc(syscall.SIGTERM, exit)
	RegSignalFunc(syscall.SIGQUIT, exit)
	RegSignalFunc(syscall.SIGINT, exit)
	Wait()
}
