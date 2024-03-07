package sing2catconfig

import (
	"fmt"
	"os"
	"runtime"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)
var Sing2cat_logger *zap.Logger
func get_logger_format() zapcore.Encoder {
	logger_format := zap.NewProductionEncoderConfig()
	logger_format.EncodeTime = zapcore.ISO8601TimeEncoder
	logger_format.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(logger_format)
}
func get_logger_writer(level string) zapcore.WriteSyncer{
	project_dir,err := Get_value("project_dir")
	if err != nil{
		panic("Fatal!Get Working Dir failed!")
	}
	lumberJackLogger := &lumberjack.Logger{
		Filename:   fmt.Sprintf("%s/log/sing2cat-%s.log",project_dir.(string),level),
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   false,
	}
	return zapcore.AddSync(lumberJackLogger)
}

func Get_logger_Core(){
	// 获取日志格式
	logger_format := get_logger_format()
	// 获取日志写入句柄
	msg_writer := get_logger_writer("info")
	error_writer := get_logger_writer("error")
	consolo_core := zapcore.NewCore(logger_format, zapcore.AddSync(os.Stdout),zapcore.DebugLevel)
	// 获取日志实现接口
	msg_core := zapcore.NewCore(logger_format,msg_writer,zapcore.InfoLevel)
	error_core := zapcore.NewCore(logger_format,error_writer,zapcore.ErrorLevel)
	core := zapcore.NewTee(msg_core,error_core,consolo_core)
	Sing2cat_logger = zap.New(core)
}

func Logger_caller(msg string,err error,skip ...int){
	if err != nil{
		index := 1
		if len(skip) != 0{
			index = skip[0]
		}
		_,file,line,_ := runtime.Caller(index)
		Sing2cat_logger.Error(msg,zap.String("caller",fmt.Sprintf("%s:%d",file,line)),zap.Error(err))
	}else{
		Sing2cat_logger.Info(msg)
	}
	
}