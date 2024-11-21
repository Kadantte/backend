package log

import (
	"log"
	"os"
)

var (
	Info  = log.New(os.Stderr, " info: ", log.LstdFlags|log.Lmsgprefix)
	Error = log.New(os.Stderr, "error: ", log.LstdFlags|log.Lmsgprefix)
)
