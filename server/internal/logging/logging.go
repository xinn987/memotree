// Package logging 负责后端进程自己的日志输出初始化。
//
// 本地开发阶段继续使用标准库 log，统一把 API/Worker 日志同时写到终端和文件；
// 启动脚本只负责拉起进程，不负责代管后端日志文件。
package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

const defaultLogDir = "logs"

// ServiceConfig 描述单个后端服务的日志输出位置。
type ServiceConfig struct {
	ServiceName   string
	LogDir        string
	ConsoleWriter io.Writer
}

// ConfigureFromEnv 使用 LOG_DIR 初始化服务日志。
// LOG_DIR 为空时默认写入当前工作目录下的 logs/。
func ConfigureFromEnv(serviceName string) (func(), error) {
	return Configure(ServiceConfig{
		ServiceName:   serviceName,
		LogDir:        os.Getenv("LOG_DIR"),
		ConsoleWriter: os.Stdout,
	})
}

// Configure 把标准库 log 的输出切到“终端 + 服务日志文件”。
// 返回的 cleanup 会关闭文件并恢复原来的标准库 log 输出，主要供测试和短生命周期命令使用。
func Configure(config ServiceConfig) (func(), error) {
	if config.ServiceName == "" {
		return nil, fmt.Errorf("service name is required")
	}
	logDir := config.LogDir
	if logDir == "" {
		logDir = defaultLogDir
	}
	if config.ConsoleWriter == nil {
		config.ConsoleWriter = os.Stdout
	}
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, err
	}

	logFilePath := filepath.Join(logDir, config.ServiceName+".log")
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}

	previousOutput := log.Writer()
	previousFlags := log.Flags()
	previousPrefix := log.Prefix()
	log.SetOutput(io.MultiWriter(config.ConsoleWriter, logFile))
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.SetPrefix("")

	return func() {
		log.SetOutput(previousOutput)
		log.SetFlags(previousFlags)
		log.SetPrefix(previousPrefix)
		_ = logFile.Close()
	}, nil
}
