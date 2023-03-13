package asynq_learn

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"testing"
)

// 如何在log封装一个日志服务

type ServerA struct {
	log *LoggerA
	// .... 其他成员忽略
}

type BaseInterface interface {
	// Debug logs a message at Debug level.
	Debug(args ...interface{})

	// Info logs a message at Info level.
	Info(args ...interface{})

	// Warn logs a message at Warning level.
	Warn(args ...interface{})

	// Error logs a message at Error level.
	Error(args ...interface{})

	// Fatal logs a message at Fatal level
	// and process will exit with status set to 1.
	Fatal(args ...interface{})
}

type LoggerA struct {
	base  BaseInterface // 任何满足了BaseInterface的日志服务都可以作为base的值
	level Level         // 日志等级
	mu    sync.Mutex    // 日志锁
}

// 定义一个默认的日志服务，当然也可以传入其他的日志服务
type DefaultLog struct {
	*log.Logger // 日志服务
}

// Debug logs a message at Debug level.
func (l *DefaultLog) Debug(args ...interface{}) {
	l.prefixPrint("DEBUG: ", args...)
}

// Info logs a message at Info level.
func (l *DefaultLog) Info(args ...interface{}) {
	l.prefixPrint("INFO: ", args...)
}

// Warn logs a message at Warning level.
func (l *DefaultLog) Warn(args ...interface{}) {
	l.prefixPrint("WARN: ", args...)
}

// Error logs a message at Error level.
func (l *DefaultLog) Error(args ...interface{}) {
	l.prefixPrint("ERROR: ", args...)
}

// Fatal logs a message at Fatal level
// and process will exit with status set to 1.
func (l *DefaultLog) Fatal(args ...interface{}) {
	l.prefixPrint("FATAL: ", args...)
	os.Exit(1)
}

func (l *DefaultLog) prefixPrint(prefix string, args ...interface{}) {
	args = append([]interface{}{prefix}, args...)
	l.Print(args...)
}

// 定义一个Level 类型
type Level int32

const (
	DebugLevelA Level = iota
	InfoLevelA
	WarnLevelA
	ErrorLevelA
	FatalLevelA
)

// 定义一个Level 类型的方法 输出字符串

func (l Level) String() string {
	switch l {
	case DebugLevelA:
		return "debug"
	case InfoLevelA:
		return "info"
	case WarnLevelA:
		return "warning"
	case ErrorLevelA:
		return "error"
	case FatalLevelA:
		return "fatal"
	default:
		return "unknown"
	}
}

// 公开创建日志的方法

func NewBaseLog(out io.Writer) *DefaultLog {
	prefix := fmt.Sprintf("asynq_learn: pid=%d ", os.Getpid())
	return &DefaultLog{
		log.New(out, prefix, log.Ldate|log.Ltime|log.Lmicroseconds|log.LUTC),
	}
}

// LoggerA是对DefaultLog的封装
func NewLoggerA(base BaseInterface) *LoggerA {
	// 如果传入的基础库是nil，则将日志输出到标准输出
	if base == nil {
		base = NewBaseLog(os.Stderr)
	}
	return &LoggerA{
		base:  base,
		level: DebugLevelA, // 默认的等级
	}
}

// 标准库->baseLog->默认log

// 日志等级不能一直是DebugLevelA，所以得公开一个方法用于设置LoggerA的日志等级

func (l *LoggerA) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	// 验证设置的等级是否在正常范围内
	if level < DebugLevelA || level > FatalLevelA {
		panic("等级不合法")
	}
	l.level = level
}

// 这个时候LoggerA已经基本可以用了，但是我们想在baseLog的基础上再封装一次，加入原子性限制，所以我们得重新实现各个方法

func (l *LoggerA) Debug(args ...interface{}) {
	if l.level != DebugLevelA {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.base.Debug(args)
}

func (l *LoggerA) Info(args ...interface{}) {
	if l.level != DebugLevelA {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.base.Debug(args)
}

func (l *LoggerA) Warn(args ...interface{}) {
	if l.level != DebugLevelA {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.base.Debug(args)
}

func (l *LoggerA) Error(args ...interface{}) {
	if l.level != DebugLevelA {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.base.Debug(args)
}

func (l *LoggerA) Fatal(args ...interface{}) {
	if l.level != DebugLevelA {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.base.Debug(args)
}

// 其他方法类似

// asynq中还封装了一个方法，用于格式化输出日志

func (l *LoggerA) DebugF(format string, args ...interface{}) {
	l.mu.Lock()
	l.mu.Unlock()
	l.Debug(fmt.Sprintf(format, args...))
}

// 至此，日志的封装大概结束了

var _ BaseInterface = (*LoggerA)(nil)

func TestLog(t *testing.T) {
	s := &ServerA{
		log: NewLoggerA(NewBaseLog(os.Stdout)),
	}
	s.log.Debug("tttt")
}

// 1、基本日志封装底层日志的方法，设置一些参数（前缀等），基本日志公开一个方法，用于设置日志的输出位置 io.Writer
// 2、然后我们得项目需要再次对基础日志封装，用于设置一些限制（如）日志等级之类的东西
// 3、然后将日志封装到我们得结构体中用于使用
