package log

import (
	"kernel/helper"

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
	check()
	logger.Console.Debug(msg, fields...)
	logger.File.Info(msg, fields...)
}
