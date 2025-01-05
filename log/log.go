package log

import (
	"log"
	"os"
)

type Logger struct {
	*log.Logger

	postfix string
}

func NewLogger(prefix string, postfix string) Logger {
	if postfix != "" {
		postfix = "[" + postfix + "]"
	}
	flag := log.LstdFlags | log.Lmicroseconds
	return Logger{log.New(os.Stderr, prefix, flag), postfix}
}

func (logger *Logger) Fatalln(v ...any) {
	if logger.postfix == "" {
		logger.Logger.Fatalln(v...)
		return
	}
	args := append([]any{logger.postfix}, v...)
	logger.Logger.Fatalln(args...)
}

func (logger *Logger) Println(v ...any) {
	if logger.postfix == "" {
		logger.Logger.Println(v...)
		return
	}
	args := append([]any{logger.postfix}, v...)
	logger.Logger.Println(args...)
}

func (logger *Logger) Printf(format string, v ...any) {
	if logger.postfix == "" {
		logger.Logger.Printf(format, v...)
		return
	}
	format = logger.postfix + " " + format
	logger.Logger.Printf(format, v...)
}

type Log struct {
	Info  Logger
	Error Logger
}

func NewLog(postfix string) Log {
	return Log{
		Info:  NewLogger(" info: ", postfix),
		Error: NewLogger("error: ", postfix),
	}
}
