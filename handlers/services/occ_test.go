/*
   @Time : 2021/6/29 10:51 上午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : occ_test.go
   @Package : services
   @Description: 单元测试 OCC 相关业务
*/

package services

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"github.com/offcn-jl/gaea-back-end/commons/config"
	"github.com/offcn-jl/gaea-back-end/commons/database/orm"
	"github.com/offcn-jl/gaea-back-end/commons/utt"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"testing"
	"time"
)

// TestOCCGetSign 测试 OCCGetSign 是否可以按照预期获取 OCC 接口调用 Sign
func TestOCCGetSign(t *testing.T) {
	Convey("测试 ToolsUrlShortenerCreateShortLink 是否可以按照预期新建短链接", t, func() {
		// 初始化 Request
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/?access-token=wrong-fake-access-token", nil)

		// 测试 校验 AccessToken 是否合法
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		OCCGetSign(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"AccessToken 不正确\"}")

		// 修改 AccessToken
		currentConfig := config.Get()
		currentConfig.ServicesAccessToken = "fake-access-token"
		config.Update(orm.MySQL.Gaea, currentConfig)

		// 重置测试上下文
		utt.ResetContext()
		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/?access-token=fake-access-token", bytes.NewBufferString("{\"URL\":\"https://host-1.domain\"}"))

		// 测试 获取 Sign
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		OCCGetSign(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Data\":\""+fmt.Sprintf("%x", md5.Sum([]byte(config.Get().OffcnOCCKey+time.Now().Format("20060102"))))+"\",\"Message\":\"Success\"}")
	})
}
