package log

import (
	"fmt"
	"io"
	"log"
	"os"
)

var logger = log.New(os.Stdout, "", log.LstdFlags)

func init() {
	log.SetOutput(io.Discard)
}

func Info(format string, a ...any) {
	Log("INFO", format, a...)
}

func Err(format string, a ...any) {
	Log("ERR", format, a...)
}

func Fatal(format string, a ...any) {
	Log("FATAL", format, a...)
	os.Exit(1)
}

func Log(level string, format string, a ...any) {
	logger.Println(fmt.Errorf(level+" "+format, a...))
}
