package internal

import (
	"github.com/chaosnote/go-kernel/helper"

	"go.uber.org/zap"
)

var _logger *helper.Logger

//-------------------------------------------------------------------------------------------------

/*
Console
*/
func Console(msg string, fields ...zap.Field) {
	_logger.Console.Debug(msg, fields...)
}

/*
File ...
*/
func File(msg string, fields ...zap.Field) {
	_logger.Console.Debug(msg, fields...)
	_logger.File.Info(msg, fields...)
}

/*
Fatal ...
*/
func Fatal(msg string, fields ...zap.Field) {
	_logger.Console.Debug(msg, fields...)
	_logger.File.Fatal(msg, fields...)
}

//-------------------------------------------------------------------------------------------------

func init() {
	_logger = helper.NewLogger("./dist", "build", false)
}
