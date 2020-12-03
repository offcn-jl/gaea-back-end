/*
   @Time : 2020/11/6 4:04 下午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : unit_test_tools
   @Software: GoLand
   @Description: 验证工具 单元测试
*/

package verify

import (
	"github.com/jarcoal/httpmock"
	"github.com/offcn-jl/gaea-back-end/commons/response"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"testing"
)

// TestPhone 测试 Phone 函数是否可以用来验证手机号码是否有效
func TestPhone(t *testing.T) {
	Convey("测试 Phone 函数是否可以用来验证手机号码是否有效", t, func() {
		So(Phone("17866668888"), ShouldBeTrue)      // 178 号段
		So(Phone("17166668888"), ShouldBeTrue)      // 171 号段
		So(Phone("+8617166668888"), ShouldBeTrue)   // 带国家码 +86
		So(Phone("008617166668888"), ShouldBeTrue)  // 带国家码 0086
		So(Phone("27866668888"), ShouldBeFalse)     // 287 号段
		So(Phone("+0117866668888"), ShouldBeFalse)  // 带国家码 +01
		So(Phone("000117866668888"), ShouldBeFalse) // 带国家码 0001
	})
}

// TestMisToken 测试 MisToken 是否可以校验 MIS 口令码 是否合法
func TestMisToken(t *testing.T) {
	Convey("测试 MisToken 是否可以校验 MIS 口令码 是否合法", t, func() {
		// 测试请求失败的情况
		// 此时未配置接口地址, 会直接返回请求出错
		pass, err := MisToken("")
		So(pass, ShouldBeFalse)
		So(err.Error(), ShouldContainSubstring, "unsupported protocol scheme \"\"")

		// 配置 httpmock 进行拦截
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterNoResponder(httpmock.NewJsonResponderOrPanic(http.StatusOK, response.Struct{"status": 2, "msg": "时间戳缺失或失效"}))

		// 测试 获取 MIS TOKEN 失败 并返回了错误信息
		pass, err = MisToken("")
		So(pass, ShouldBeFalse)
		So(err.Error(), ShouldContainSubstring, "时间戳缺失或失效")

		// 修改 httpmock 为不返回错误信息
		httpmock.RegisterNoResponder(httpmock.NewJsonResponderOrPanic(http.StatusOK, response.Struct{"status": 2}))

		// 测试 获取 MIS TOKEN 失败 并没有返回错误信息
		pass, err = MisToken("")
		So(pass, ShouldBeFalse)
		So(err.Error(), ShouldContainSubstring, "请求 MIS 口令码 失败")

		// 修改 httpmock 为返回成功
		httpmock.RegisterNoResponder(httpmock.NewJsonResponderOrPanic(http.StatusOK, response.Struct{"status": 1, "msg": "SIGN验签成功", "data": "fake-token"}))

		// 测试 MIS Token 错误
		pass, err = MisToken("wrong-fake-token")
		So(pass, ShouldBeFalse)
		So(err, ShouldBeNil)

		// 测试 MIS Token 正确
		pass, err = MisToken("fake-token")
		So(pass, ShouldBeTrue)
		So(err, ShouldBeNil)
	})
}
