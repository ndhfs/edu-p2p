package log

import (
	"fmt"
	"log"
	"os"
)

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
	log.Println(fmt.Errorf(level+" "+format, a...))
}
