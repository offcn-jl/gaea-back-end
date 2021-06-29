/*
   @Time : 2021/4/2 1:54 下午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : tools_url_shortener_test.go
   @Package : services
   @Description: 单元测试 工具 短链接生成器 ( 长链接转短链接 ) [ 临时使用 ]
*/

package services

import (
	"bytes"
	"github.com/offcn-jl/gaea-back-end/commons/config"
	"github.com/offcn-jl/gaea-back-end/commons/database/orm"
	"github.com/offcn-jl/gaea-back-end/commons/database/structs"
	"github.com/offcn-jl/gaea-back-end/commons/utt"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"os"
	"testing"
)

// 覆盖 orm 库中的 ORM 对象
func init() {
	utt.InitTest() // 初始化测试数据并获取测试所需的上下文
	orm.MySQL.Gaea = utt.ORM
}

// TestToolsUrlShortenerCreateShortLink 测试 ToolsUrlShortenerCreateShortLink 是否可以按照预期新建短链接
func TestToolsUrlShortenerCreateShortLink(t *testing.T) {
	Convey("测试 ToolsUrlShortenerCreateShortLink 是否可以按照预期新建短链接", t, func() {
		// 重置测试上下文
		utt.ResetContext()
		// 初始化 Request
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/?access-token=wrong-fake-access-token", nil)

		// 测试 校验 AccessToken 是否合法
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		ToolsUrlShortenerCreateShortLink(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"AccessToken 不正确\"}")

		// 修改 AccessToken
		currentConfig := config.Get()
		currentConfig.ServicesAccessToken = "fake-access-token"
		config.Update(orm.MySQL.Gaea, currentConfig)

		// 重置测试上下文
		utt.ResetContext()
		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/?access-token=fake-access-token", nil)

		// 测试 绑定参数
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		ToolsUrlShortenerCreateShortLink(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Error\":\"invalid request\",\"Message\":\"提交的 Json 数据不正确\"}")

		// 清空并重建数据库
		orm.MySQL.Gaea.DropTableIfExists(structs.ToolsUrlShortener{})
		orm.MySQL.Gaea.AutoMigrate(structs.ToolsUrlShortener{})

		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/?access-token=fake-access-token", bytes.NewBufferString("{\"URL\":\"https://host-1.domain\"}"))

		// 测试 创建记录并返回创建成功
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		ToolsUrlShortenerCreateShortLink(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Data\":{\"Repetitive\":false,\"ShortUrl\":\"https://offcn.ltd/test/b\"},\"Message\":\"Success\"}")

		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/?access-token=fake-access-token", bytes.NewBufferString("{\"URL\":\"https://host-1.domain\"}"))

		// 测试 检查是否已经存在短链记录
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		ToolsUrlShortenerCreateShortLink(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Data\":{\"Repetitive\":true,\"ShortUrl\":\"https://offcn.ltd/test/b\"},\"Message\":\"Success\"}")

		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/?access-token=fake-access-token", bytes.NewBufferString("{\"URL\":\"https://host-1.domain\"}"))

		// 修改运行环境为生产环境
		os.Setenv("GIN_MODE", "release")

		// 测试 生产环境
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		ToolsUrlShortenerCreateShortLink(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Data\":{\"Repetitive\":true,\"ShortUrl\":\"https://offcn.ltd/b\"},\"Message\":\"Success\"}")
	})
}
