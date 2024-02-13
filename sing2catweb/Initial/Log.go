package initial

import (
	"fmt"
	"os"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)
var Logger *zap.Logger
func InitialLogger() {
	encoder_config := zap.NewProductionEncoderConfig()
	// 设置日志记录中时间的格式
	encoder_config.EncodeTime = zapcore.ISO8601TimeEncoder
	// 日志Encoder 还是JSONEncoder，把日志行格式化成JSON格式的
	encoder := zapcore.NewConsoleEncoder(encoder_config)
	// 设置日志路径
	proxy_info := logWritter(fmt.Sprintf("%s/log/proxy_info.log",GetValue("base_dir")))
	proxy_error := logWritter(fmt.Sprintf("%s/log/proxy_error.log",GetValue("base_dir")))
	// 最终实现写日志
	core := zapcore.NewTee(
	// 同时向控制台和文件写日志， 生产环境记得把控制台写入去掉，日志记录的基本是Debug 及以上，生产环境记得改成Info
		zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout),zapcore.DebugLevel),
		zapcore.NewCore(encoder, proxy_info, zapcore.InfoLevel),
		zapcore.NewCore(encoder, proxy_error, zapcore.ErrorLevel),
	)
	Logger = zap.New(core,zap.AddCaller())
   }
func logWritter(path string) zapcore.WriteSyncer{	
	lumberWriteSyncer := &lumberjack.Logger{
	Filename:   path,
	MaxSize:    10, // megabytes
	MaxBackups: 100,
	MaxAge:     28,    // days
	Compress:   false, //Compress确定是否应该使用gzip压缩已旋转的日志文件。默认值是不执行压缩。
}
return zapcore.AddSync(lumberWriteSyncer)
}