package internal

import (
	"github.com/chaosnote/go-kernel/helper"

	"go.uber.org/zap"
)

var _logger *zap.Logger

//-------------------------------------------------------------------------------------------------

/*
Console
*/
func Console(msg string, fields ...zap.Field) {
	helper.Console.Debug(msg, fields...)
}

/*
File ...
*/
func File(msg string, fields ...zap.Field) {
	helper.Console.Debug(msg, fields...)
	_logger.Info(msg, fields...)
}

/*
Fatal ...
*/
func Fatal(msg string, fields ...zap.Field) {
	helper.Console.Debug(msg, fields...)
	_logger.Fatal(msg, fields...)
}

//-------------------------------------------------------------------------------------------------

func init() {
	_logger = helper.NewFileLogger("./dist", "kernel")
}
