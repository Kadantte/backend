package log

import (
	"log"
	"os"
)

var (
	flag  = log.LstdFlags | log.Lmicroseconds
	Info  = log.New(os.Stderr, " info: ", flag)
	Error = log.New(os.Stderr, "error: ", flag)
)
