/*
   @Time : 2020/10/31 3:04 下午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : logger
   @Software: GoLand
*/

package logger

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/offcn-jl/gaea-back-end/commons/config"
	"runtime/debug"
	"time"
)

// Log 打印日志
func Log(log string) {
	_, _ = fmt.Fprintf(gin.DefaultWriter, "[ GAEA - 日志 ] %v | %s\n",
		time.Now().Format("2006/01/02 - 15:04:05"),
		log,
	)
}

// Error 打印错误日志及调用堆栈
func Error(err error) {
	_, _ = fmt.Fprintf(gin.DefaultWriter, "[ GAEA - 错误 ] %v | %s\n",
		time.Now().Format("2006/01/02 - 15:04:05"),
		err,
	)
	_, _ = fmt.Fprintf(gin.DefaultWriter, "[ GAEA - 错误 - 调用堆栈 ] %v\n%s\n",
		time.Now().Format("2006/01/02 - 15:04:05"),
		debug.Stack(), // 获取调用堆栈
	)
	// debug.PrintStack() // 打印调用堆栈
}

// Panic 打印错误后抛出 panic
func Panic(err error) {
	_, _ = fmt.Fprintf(gin.DefaultWriter, "[ GAEA - 异常 - PANIC ] %v | %s\n",
		time.Now().Format("2006/01/02 - 15:04:05"),
		err,
	)
	_, _ = fmt.Fprintf(gin.DefaultWriter, "[ GAEA - 异常 - PANIC - 调用堆栈 ] %v\n%s\n",
		time.Now().Format("2006/01/02 - 15:04:05"),
		debug.Stack(), // 输出调用堆栈
	)
	panic(err)
}

// DebugToJson 调试输出为 Json 字符串
func DebugToJson(key string, value interface{}) {
	if !config.Get().DisableDebug {
		jsonStrings, _ := json.Marshal(value)
		_, _ = fmt.Fprintf(gin.DefaultWriter, "[ GAEA - 调试 - JSON ] %v | %s --> %s\n",
			time.Now().Format("2006/01/02 - 15:04:05"),
			key,
			jsonStrings,
		)
	}
}

// DebugToString 调试输出为字符串
func DebugToString(key string, value interface{}) {
	if !config.Get().DisableDebug {
		_, _ = fmt.Fprintf(gin.DefaultWriter, "[ GAEA - 调试 - 字符串 ] %v | %s --> %s\n",
			time.Now().Format("2006/01/02 - 15:04:05"),
			key,
			fmt.Sprint(value),
		)
	}
}
