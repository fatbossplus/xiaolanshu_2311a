package logger

import (
	"backend/pkg/config"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

// Init 根据配置初始化全局 logger，应在服务启动时调用一次
func Init(cfg config.LoggerConfig) {
	var level zapcore.Level
	_ = level.UnmarshalText([]byte(cfg.Level)) // 解析日志级别，如 "debug"/"info"/"warn"/"error"

	encCfg := zap.NewProductionEncoderConfig()
	encCfg.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05") // 时间格式
	encCfg.EncodeLevel = zapcore.CapitalLevelEncoder                       // 级别大写，如 INFO、ERROR

	// 默认输出到控制台；配置了 file 目录则同时写入文件
	ws := zapcore.AddSync(os.Stdout) // 输出到控制台
	if cfg.File != "" {
		if dw, err := newDailyWriter(cfg.File, cfg.Service); err == nil {
			ws = zapcore.NewMultiWriteSyncer(ws, zapcore.AddSync(dw))
		}
	}

	log = zap.New(
		zapcore.NewCore(zapcore.NewJSONEncoder(encCfg), ws, level),
		zap.AddCaller(),      // 记录调用位置，如 handler/auth.go:42
		zap.AddCallerSkip(1), // 跳过 logger 包本身，指向真实调用方
		zap.Fields(zap.String("service", cfg.Service)), // 每条日志自动携带服务名
	)
}

// Debug/Info/Warn/Error/Fatal 全局日志方法，fields 传结构化字段
func Debug(msg string, fields ...zap.Field) { log.Debug(msg, fields...) }
func Info(msg string, fields ...zap.Field)  { log.Info(msg, fields...) }
func Warn(msg string, fields ...zap.Field)  { log.Warn(msg, fields...) }
func Error(msg string, fields ...zap.Field) { log.Error(msg, fields...) }
func Fatal(msg string, fields ...zap.Field) { log.Fatal(msg, fields...) }

// InjectCtx 将携带指定字段的派生 logger 注入 gin.Context
// 通常在 TraceMiddleware 中调用，注入 traceID，使后续日志自动携带
func InjectCtx(c *gin.Context, fields ...zap.Field) { c.Set("_logger", log.With(fields...)) }

// FromCtx 从 gin.Context 取出 logger，取不到则返回全局 logger
// 在 handler 中使用：logger.FromCtx(c).Info("xxx")，日志自动带 traceID
func FromCtx(c *gin.Context) *zap.Logger {
	if l, ok := c.Get("_logger"); ok {
		return l.(*zap.Logger)
	}
	return log
}

// dailyWriter 按天滚动的文件 writer
// 每次写入时检查日期，跨天后自动切换到新文件，无需重启服务
// 文件路径格式：{dir}/{service}/YYYY-MM-DD.log
type dailyWriter struct {
	mu      sync.Mutex
	dir     string   // 日志根目录
	service string   // 服务名，作为子目录
	curDate string   // 当前文件对应的日期
	file    *os.File // 当前打开的文件句柄
}

func newDailyWriter(dir, service string) (*dailyWriter, error) {
	w := &dailyWriter{dir: dir, service: service}
	return w, w.rotate(today())
}

func (w *dailyWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	// 日期变更时滚动到新文件，rotate 失败则继续写旧文件
	if d := today(); d != w.curDate {
		_ = w.rotate(d)
	}
	return w.file.Write(p)
}

func (w *dailyWriter) Sync() error { return w.file.Sync() }

// rotate 关闭当前文件，按日期创建新文件
func (w *dailyWriter) rotate(date string) error {
	if w.file != nil {
		_ = w.file.Close()
	}
	dir := filepath.Join(w.dir, w.service)
	_ = os.MkdirAll(dir, 0755)
	f, err := os.OpenFile(filepath.Join(dir, date+".log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	w.file, w.curDate = f, date
	return nil
}

func today() string { return time.Now().Format("2006-01-02") }
