package webservice

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	// LogLevelDebug marks debug logging level
	logLevelDebug uint8 = iota
	// LogLevelTrace marks trace logging level
	logLevelTrace
	// LogLevelWarn marks warn logging level
	logLevelWarn
	// LogLevelError marks error logging level
	logLevelError
)

// Logger defines logger interface
type Logger interface {
	// SetLevel
	SetLevel(uint8)
	CheckLevel(uint8) bool
	Write(string, bool, ...interface{})
	Debug(...interface{})
	Trace(...interface{})
	Warn(...interface{})
	Error(...interface{})
}

// ConvertLoggerMust converts a object to Logger interface, build built-in logger if convert failed
func ConvertLoggerMust(log interface{}) Logger {
	if log == nil {
		return &logger{}
	}

	if l := log.(Logger); l != nil {
		return l
	}

	return &logger{}
}

type logger struct {
	level uint8
}

func (l *logger) SetLevel(lv uint8) {
	l.level = lv
}

func (l logger) CheckLevel(lv uint8) bool {
	return l.level <= lv
}

func addFuncNameTologs(args []interface{}) []interface{} {
	pc := make([]uintptr, 1)
	runtime.Callers(3, pc)
	f := runtime.FuncForPC(pc[0])
	logs := make([]interface{}, 0, len(args)+1)
	logs = append(logs, f.Name())
	logs = append(logs, args...)

	return logs
}

func (l logger) Write(suffixInfo string, suffix bool, args ...interface{}) {
	contents := format("trace", suffix, suffixInfo, args...)
	fmt.Println(contents)
}

func (l logger) Debug(args ...interface{}) {
	if l.CheckLevel(logLevelDebug) {
		logs := addFuncNameTologs(args)
		contents := format("debug", false, "", logs...)
		fmt.Println(contents)
	}
}

func (l logger) Trace(args ...interface{}) {
	if l.CheckLevel(logLevelTrace) {
		logs := addFuncNameTologs(args)
		contents := format("trace", false, "", logs...)
		fmt.Println(contents)
	}
}

func (l logger) Warn(args ...interface{}) {
	if l.CheckLevel(logLevelWarn) {
		logs := addFuncNameTologs(args)
		contents := format("warn", false, "", logs...)
		fmt.Println(contents)
	}
}

func (l logger) Error(args ...interface{}) {
	if l.CheckLevel(logLevelError) {
		logs := addFuncNameTologs(args)
		contents := format("error", false, "", logs...)
		fmt.Println(contents)
	}
}

func getDatetime() string {
	return time.Now().Format("2006-01-02 15:04:05.000")
}

func format(prefix string, suffix bool, suffixInfo string, args ...interface{}) string {
	content := "|" + prefix
	for _, arg := range args {
		switch arg.(type) {
		case int:
			content = content + "|" + strconv.Itoa(arg.(int))
			break
		case string:
			content = content + "|" + strings.TrimRight(arg.(string), "\n")
			break
		case int64:
			str := strconv.FormatInt(arg.(int64), 10)
			content = content + "|" + str
			break
		default:
			content = content + "|" + fmt.Sprintf("%v", arg)
			break
		}
	}
	if suffix {
		content = getDatetime() + content + "|" + suffixInfo
	} else {
		content = getDatetime() + content
	}
	return content
}
