package log

import (
	"log"
	"os"
)

type Logger struct {
	*log.Logger

	idPrefix string
}

func New(levelPrefix string, id string) Logger {
	var idPrefix string
	if id != "" {
		idPrefix = "[" + id + "]"
	}

	flag := log.LstdFlags | log.Lmicroseconds
	return Logger{log.New(os.Stderr, levelPrefix, flag), idPrefix}
}

func (logger *Logger) Println(v ...any) {
	if logger.idPrefix == "" {
		logger.Logger.Println(v...)
		return
	}
	args := append([]any{logger.idPrefix}, v...)
	logger.Logger.Println(args...)
}

func (logger *Logger) Printf(format string, v ...any) {
	if logger.idPrefix == "" {
		logger.Logger.Printf(format, v...)
		return
	}
	format = logger.idPrefix + " " + format
	logger.Logger.Printf(format, v...)
}

var (
	Info  = New(" info: ", "")
	Error = New("error: ", "")
)
