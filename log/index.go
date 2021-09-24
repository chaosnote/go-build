package log

import (
	"github.com/chaosnote/go-kernel/helper"

	"go.uber.org/zap"
)

var logger *helper.Logger

// Set ...
func Set(path, name string, b bool) {
	logger = helper.NewLogger(path, name, b)
}

func check() {
	if logger == nil {
		panic("unset log.Set('name')")
	}
}

/*
Console
*/
func Console(msg string, fields ...zap.Field) {
	check()
	logger.Console.Debug(msg, fields...)
}

/*
File ...
*/
func File(msg string, fields ...zap.Field) {
	Console(msg, fields...)
	logger.File.Info(msg, fields...)
}

/*
Fatal ...
*/
func Fatal(msg string, fields ...zap.Field) {
	Console(msg, fields...)
	logger.File.Fatal(msg, fields...)
}
