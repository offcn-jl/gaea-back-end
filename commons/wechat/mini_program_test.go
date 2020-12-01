/*
   @Time : 2020/12/1 3:59 下午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : mini_program_test
   @Software: GoLand
   @Description: 微信小程序的单元测试
*/

package wechat

import (
	"github.com/jarcoal/httpmock"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"testing"
)

// TestMiniProgramCreateQrCode 测试 MiniProgramCreateQrCode 是否可以创建微信小程序码
func TestMiniProgramCreateQrCode(t *testing.T) {
	Convey("测试 MiniProgramCreateQrCode 是否可以创建小程序码", t, func() {
		// 配置 httpmock 进行拦截
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder(http.MethodGet, "https://api.gaea.jilinoffcn.com/release/services/authentication/mini-program/get/access-token", httpmock.NewBytesResponder(http.StatusOK, nil))

		// 测试获取访问令牌失败
		miniProgramQrCodeImage, err := MiniProgramCreateQrCode("fake-app-id", "", "", 0, false, nil, false)
		So(miniProgramQrCodeImage, ShouldBeEmpty)
		So(err.Error(), ShouldEqual, "unexpected end of JSON input")

		// 配置 httpmock 返回正确的访问令牌
		httpmock.RegisterResponder(http.MethodGet, "https://api.gaea.jilinoffcn.com/release/services/authentication/mini-program/get/access-token", httpmock.NewStringResponder(http.StatusOK, "{\"Message\":\"Success\",\"Data\":\"fake-access-token\"}"))

		// 测试调用接口失败
		miniProgramQrCodeImage, err = MiniProgramCreateQrCode("fake-app-id", "", "", 0, false, nil, false)
		So(miniProgramQrCodeImage, ShouldBeEmpty)
		So(err.Error(), ShouldEqual, "Post \"https://api.weixin.qq.com/wxa/getwxacodeunlimit?access_token=fake-access-token\": no responder found")

		// 配置 httpmock 返回微信小程序码
		fakeImage := make([]byte, 10)
		httpmock.RegisterResponder(http.MethodPost, "https://api.weixin.qq.com/wxa/getwxacodeunlimit", httpmock.NewBytesResponder(http.StatusOK, fakeImage))

		// 测试创建小程序码成功
		miniProgramQrCodeImage, err = MiniProgramCreateQrCode("fake-app-id", "", "", 0, false, nil, false)
		So(string(miniProgramQrCodeImage), ShouldEqual, string(fakeImage))
		So(err, ShouldBeEmpty)

		// 配置 httpmock 返回可以格式化为 Json 的数据
		httpmock.RegisterResponder(http.MethodPost, "https://api.weixin.qq.com/wxa/getwxacodeunlimit", httpmock.NewStringResponder(http.StatusOK, "{\"errcode\":0,\"errmsg\":\"none\"}"))

		// 测试创建小程序码失败
		miniProgramQrCodeImage, err = MiniProgramCreateQrCode("fake-app-id", "", "", 0, false, nil, false)
		So(miniProgramQrCodeImage, ShouldBeEmpty)
		So(err.Error(), ShouldEqual, "创建小程序码失败 [ 0 ] none")
	})
}

// TestMiniProgramGetAccessToken 测试 MiniProgramGetAccessToken 是否可以获取微信小程序访问令牌
func TestMiniProgramGetAccessToken(t *testing.T) {
	Convey("测试 MiniProgramGetAccessToken 是否可以获取微信小程序访问令牌", t, func() {
		// 配置 httpmock 进行拦截
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder(http.MethodGet, "https://api.gaea.jilinoffcn.com/release/services/authentication/mini-program/get/access-token", httpmock.NewBytesResponder(http.StatusOK, nil))

		// 测试请求失败
		accessToken, err := MiniProgramGetAccessToken("fake-app-id")
		So(accessToken, ShouldBeEmpty)
		So(err.Error(), ShouldEqual, "unexpected end of JSON input")

		// 配置 httpmock 返回错误
		httpmock.RegisterResponder(http.MethodGet, "https://api.gaea.jilinoffcn.com/release/services/authentication/mini-program/get/access-token", httpmock.NewStringResponder(http.StatusOK, "{\"Message\":\"发生错误\"}"))

		// 测试请求成功, 但是接口返回了错误
		accessToken, err = MiniProgramGetAccessToken("fake-app-id")
		So(accessToken, ShouldBeEmpty)
		So(err.Error(), ShouldEqual, "发生错误")

		// 配置 httpmock 返回成功
		httpmock.RegisterResponder(http.MethodGet, "https://api.gaea.jilinoffcn.com/release/services/authentication/mini-program/get/access-token", httpmock.NewStringResponder(http.StatusOK, "{\"Message\":\"Success\",\"Data\":\"fake-access-token\"}"))

		// 测试请求成功
		accessToken, err = MiniProgramGetAccessToken("fake-app-id")
		So(accessToken, ShouldEqual, "fake-access-token")
		So(err, ShouldBeEmpty)
	})
}
