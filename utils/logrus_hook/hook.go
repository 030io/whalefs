package logrus_hook

import (
	"runtime"
	"strings"
	"path/filepath"
	log "github.com/Sirupsen/logrus"
)

type ContextHook struct {
}

func (hook ContextHook)Levels() []log.Level {
	return log.AllLevels
}

func (hook ContextHook)Fire(entry *log.Entry) error {
	pc := make([]uintptr, 10)
	runtime.Callers(6, pc)
	frames := runtime.CallersFrames(pc)
	frame, _ := frames.Next()

	funcName := frame.Func.Name()
	funcName = funcName[strings.LastIndexByte(funcName, filepath.Separator) + 1 :]
	fileName := frame.File[strings.LastIndexByte(frame.File, filepath.Separator) + 1:]

	entry.Data["file"] = fileName
	entry.Data["func"] = funcName
	entry.Data["line"] = frame.Line

	//for {
	//	frame, more := frames.Next()
	//	println(frame.File)
	//	println(frame.Func.Name())
	//	println(frame.Line)
	//	println("")
	//
	//	if !more{
	//		break
	//	}
	//}

	return nil
}

func init() {
	log.AddHook(ContextHook{})
}
