/*
   @Time : 2020/11/4 3:15 下午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : logger_test
   @Software: GoLand
   @Description: 日志组件 单元测试
*/

package logger

import (
	"bytes"
	"errors"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

// TestLog 测试 Log 函数是否输出期望内容
func TestLog(t *testing.T) {
	Convey("测试 Log 函数是否输出期望内容", t, func() {
		// 定义接收输出的 buffer
		buffer := new(bytes.Buffer)
		// 将默认的输出 writer 修改为 buffer
		DefaultWriter = buffer
		// 输出预期字符串到日志
		Log("期望内容")
		So(buffer.String(), ShouldContainSubstring, "[ GAEA - 日志 ]")
		So(buffer.String(), ShouldContainSubstring, "期望内容")
	})
}

// TestError 测试 Error 函数是否输出期望内容
func TestError(t *testing.T) {
	Convey("测试 Error 函数是否输出期望内容", t, func() {
		// 定义接收输出的 buffer
		buffer := new(bytes.Buffer)
		// 将默认的输出 writer 修改为 buffer
		DefaultWriter = buffer
		// 输出预期字符串到日志
		Error(errors.New("期望内容"))
		So(buffer.String(), ShouldContainSubstring, "[ GAEA - 错误 ]")
		So(buffer.String(), ShouldContainSubstring, "[ GAEA - 错误 - 调用堆栈 ]")
		So(buffer.String(), ShouldContainSubstring, "期望内容")
	})
}

// TestPanic 测试 Panic 函数是否输出期望内容后抛出 PANIC
func TestPanic(t *testing.T) {
	Convey("测试 Panic 函数是否输出期望内容后抛出 PANIC", t, func() {
		// 定义 err 变量用于断言，如果在调用函数和进行断言时分别进行定义，会出现错误的内容一致，但并不是同一个 "错误" 所以无法通过断言的情况
		err := errors.New("期望内容")
		// 定义接收输出的 buffer
		buffer := new(bytes.Buffer)
		// 将默认的输出 writer 修改为 buffer
		DefaultWriter = buffer
		// 输出预期字符串到日志
		So(func() { Panic(err) }, ShouldPanicWith, err)
		So(buffer.String(), ShouldContainSubstring, "[ GAEA - 异常 - PANIC ]")
		So(buffer.String(), ShouldContainSubstring, "[ GAEA - 异常 - PANIC - 调用堆栈 ]")
		So(buffer.String(), ShouldContainSubstring, "期望内容")
	})
}

// TestDebugToJson 测试 DebugToJson 函数是否在调试模式开启时将参数调试输出为 Json 字符串
func TestDebugToJson(t *testing.T) {
	// 定义接收输出的 buffer
	buffer := new(bytes.Buffer)
	// 将默认的输出 writer 修改为 buffer
	DefaultWriter = buffer

	Convey("测试 DebugToJson 函数是否在调试模式开启时将参数调试输出为 Json 字符串", t, func() {
		Convey("测试 未开启 调试模式时的情况", func() {
			// 禁用调试模式
			debugMod = false
			// 输出预期字符串到日志
			DebugToJson("期望参数名", []string{"期望参数一", "期望参数二"})
			// 判断日志是否唯恐
			So(buffer.String(), ShouldBeEmpty)
		})
		Convey("测试 开启 调试模式时的情况", func() {
			// 启用调试模式
			debugMod = true
			// 输出预期字符串到日志
			DebugToJson("期望参数名", []string{"期望参数一", "期望参数二"})
			// 判断日志内容是否符合预期
			So(buffer.String(), ShouldContainSubstring, "[ GAEA - 调试 - JSON ]")
			So(buffer.String(), ShouldContainSubstring, "期望参数名")
			So(buffer.String(), ShouldContainSubstring, "[\"期望参数一\",\"期望参数二\"]")
			So(buffer.String(), ShouldContainSubstring, "期望参数名 --> [\"期望参数一\",\"期望参数二\"]")
		})
	})
}

// TestDebugToString 测试 DebugToString 函数是否在调试模式开启时将参数调试输出为字符串
func TestDebugToString(t *testing.T) {
	// 定义接收输出的 buffer
	buffer := new(bytes.Buffer)
	// 将默认的输出 writer 修改为 buffer
	DefaultWriter = buffer
	Convey("测试 DebugToString 函数是否在调试模式开启时将参数调试输出为字符串", t, func() {
		Convey("测试 未开启 调试模式时的情况", func() {
			// 禁用调试模式
			debugMod = false
			// 输出预期字符串到日志
			DebugToString("期望参数名", []string{"期望参数一", "期望参数二"})
			// 判断日志是否唯恐
			So(buffer.String(), ShouldBeEmpty)
		})
		Convey("测试 开启 调试模式时的情况", func() {
			// 启用调试模式
			debugMod = true
			// 输出预期字符串到日志
			DebugToString("期望参数名", []string{"期望参数一", "期望参数二"})
			// 判断日志内容是否符合预期
			So(buffer.String(), ShouldContainSubstring, "[ GAEA - 调试 - 字符串 ]")
			So(buffer.String(), ShouldContainSubstring, "期望参数名")
			So(buffer.String(), ShouldContainSubstring, "[期望参数一 期望参数二]")
			So(buffer.String(), ShouldContainSubstring, "期望参数名 --> [期望参数一 期望参数二]")
		})
	})
}
