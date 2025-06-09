package middleware

import (
	"fmt"
	"log"
	"os"
	"sync/atomic"

	"github.com/guyigood/gyweb/core/gyarn"
)

var (
	// debugEnabled 调试模式开关
	debugEnabled uint32
	// debugLogger 调试日志记录器
	debugLogger = log.New(os.Stdout, "[GYWEB-DEBUG] ", log.LstdFlags|log.Lshortfile)
)

// SetDebug 设置调试模式
func SetDebug(enable bool) {
	if enable {
		atomic.StoreUint32(&debugEnabled, 1)
	} else {
		atomic.StoreUint32(&debugEnabled, 0)
	}
}

// IsDebugEnabled 检查调试模式是否启用
func IsDebugEnabled() bool {
	return atomic.LoadUint32(&debugEnabled) == 1
}

// init 初始化调试模式
func init() {
	// 从环境变量读取调试模式设置
	if os.Getenv("GYWEB_DEBUG") == "true" {
		SetDebug(true)
	}
}

// debugLog 输出调试日志
func debugLog(format string, args ...interface{}) {
	if IsDebugEnabled() {
		debugLogger.Output(2, fmt.Sprintf(format, args...))
	}
}

// debugAuth 输出认证相关的调试信息
func debugAuth(c *gyarn.Context, msg string, args ...interface{}) {
	if IsDebugEnabled() {
		debugLogger.Output(2, fmt.Sprintf("[Auth] %s - %s %s: %s",
			c.Request.RemoteAddr,
			c.Request.Method,
			c.Request.URL.Path,
			fmt.Sprintf(msg, args...),
		))
	}
}

// debugWhitelist 输出白名单匹配的调试信息
func debugWhitelist(path string, matched bool, matchType string) {
	if IsDebugEnabled() {
		debugLogger.Output(2, fmt.Sprintf("[Whitelist] Path: %s, Type: %s, Matched: %v",
			path,
			matchType,
			matched,
		))
	}
}

// debugAuthFunc 输出认证函数执行的调试信息
func debugAuthFunc(c *gyarn.Context, success bool) {
	if IsDebugEnabled() {
		debugLogger.Output(2, fmt.Sprintf("[AuthFunc] %s - %s %s: %v",
			c.Request.RemoteAddr,
			c.Request.Method,
			c.Request.URL.Path,
			success,
		))
	}
}

// debugUnauthorized 输出未授权处理的调试信息
func debugUnauthorized(c *gyarn.Context) {
	if IsDebugEnabled() {
		debugLogger.Output(2, fmt.Sprintf("[Unauthorized] %s - %s %s",
			c.Request.RemoteAddr,
			c.Request.Method,
			c.Request.URL.Path,
		))
	}
}

// DebugSQL 输出 SQL 语句和参数的调试信息
func DebugSQL(sql string, args ...interface{}) {
	if IsDebugEnabled() {
		debugLogger.Output(2, fmt.Sprintf("[SQL] Query: %s\nArgs: %v",
			sql,
			args,
		))
	}
}

func DebugVar(varname string, v interface{}) {
	if IsDebugEnabled() {
		debugLogger.Output(2, fmt.Sprintf("[VAR] %s: %v", varname, v))
	}
}
