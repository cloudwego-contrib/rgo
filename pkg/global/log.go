package global

import (
	"fmt"
	"github.com/cloudwego-contrib/rgo/pkg/global/consts"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"path/filepath"
	"time"
)

var Logger *zap.Logger

func InitLogger(rgoBasePath, curWorkPath string) {
	currentTime := time.Now().Format("2006-01-02")
	logFileName := filepath.Join(rgoBasePath, consts.LogPath, fmt.Sprintf("rgo_%s", curWorkPath), fmt.Sprintf("%s.log", currentTime))

	lumberjackLogger := &lumberjack.Logger{
		Filename:   logFileName,
		MaxSize:    100,
		MaxBackups: 30,
		MaxAge:     30,
		Compress:   true,
	}

	writeSyncer := zapcore.AddSync(lumberjackLogger)
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		writeSyncer,
		zapcore.InfoLevel,
	)

	Logger = zap.New(core, zap.AddCaller())
	defer Logger.Sync()

	Logger.Info("test log info")
}
