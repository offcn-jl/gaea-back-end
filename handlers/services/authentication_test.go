/*
   @Time : 2020/11/29 1:44 下午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : authentication_test
   @Software: GoLand
   @Description: 认证服务的单元测试
*/

package services

import (
	"fmt"
	"github.com/jarcoal/httpmock"
	"github.com/offcn-jl/gaea-back-end/commons/config"
	"github.com/offcn-jl/gaea-back-end/commons/database/orm"
	"github.com/offcn-jl/gaea-back-end/commons/database/structs"
	"github.com/offcn-jl/gaea-back-end/commons/response"
	"github.com/offcn-jl/gaea-back-end/commons/utt"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"testing"
	"time"
)

// TestGetMiniProgramAccessToken 测试 GetMiniProgramAccessToken 是否可以获取微信小程序 AccessToken
func TestGetMiniProgramAccessToken(t *testing.T) {
	Convey("测试 GetMiniProgramAccessToken 是否可以获取微信小程序 AccessToken", t, func() {
		// 初始化 Request
		utt.GinTestContext.Request, _ = http.NewRequest("GET", "", nil)

		// 修改请求上下文, 修改为非生产环境的请求
		utt.GinTestContext.Request.Host = "fake.request"
		utt.GinTestContext.Request.RequestURI = "/test/services/authentication/mini-program/get/access-token?access-token=fake-access-token&app-id=fake-app-id"

		// 配置 httpmock 拦截发送到生产环境的请求
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		// 测试调用生产环境，但是发送请求失败的情况
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		GetMiniProgramAccessToken(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Error\":\"Get \\\"https://fake.request/release/services/authentication/mini-program/get/access-token?access-token=fake-access-token\\u0026app-id=fake-app-id\\\": no responder found\",\"Message\":\"发送请求失败\"}")

		// 配置 httpmock 返回无法读取的 body  todo 暂时无法测试

		// 配置 httpmock 返回响应
		httpmock.RegisterResponder(http.MethodGet, "https://fake.request/release/services/authentication/mini-program/get/access-token?access-token=fake-access-token&app-id=fake-app-id", httpmock.NewStringResponder(http.StatusOK, "生产环境的响应"))

		// 测试调用生产环境
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		GetMiniProgramAccessToken(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "生产环境的响应")

		// 修改请求上下文, 修改为错误的生产环境的请求
		utt.GinTestContext.Request.RequestURI = "/release/services/authentication/mini-program/get/access-token"

		// 测试绑定数据错误
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		GetMiniProgramAccessToken(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Error\":\"Key: 'AccessToken' Error:Field validation for 'AccessToken' failed on the 'required' tag\\nKey: 'AppID' Error:Field validation for 'AppID' failed on the 'required' tag\",\"Message\":\"提交的 Query 查询不正确\"}")

		// 修改请求上下文, 修改为正确的生产环境的请求
		utt.GinTestContext.Request, _ = http.NewRequest("GET", "https://fake.request/release/services/authentication/mini-program/get/access-token?access-token=fake-access-token&app-id=fake-app-id", nil)
		utt.GinTestContext.Request.RequestURI = "/release/services/authentication/mini-program/get/access-token"

		// 测试校验 AccessToken 不合法
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		GetMiniProgramAccessToken(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"AccessToken 不正确\"}")

		// 修改 AccessToken
		currentConfig := config.Get()
		currentConfig.ServicesAccessToken = "fake-access-token"
		config.Update(orm.MySQL.Gaea, currentConfig)

		// 测试对应的小程序不存在
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		GetMiniProgramAccessToken(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"AppID fake-app-id 对应的小程序不存在\"}")

		// 添加小程序设置
		orm.MySQL.Gaea.Create(&structs.MiniProgram{AppID: "fake-app-id", Secret: "fake-secret"})

		// 测试不存在 AccessToken 记录, 调用获取函数获取一条 AccessToken 并返回
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		GetMiniProgramAccessToken(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Error\":\"Get \\\"https://api.weixin.qq.com/cgi-bin/token?appid=fake-app-id\\u0026grant_type=client_credential\\u0026secret=fake-secret\\\": no responder found\",\"Message\":\"发送请求失败\"}")

		// 添加一条已经超过有效期的记录
		orm.MySQL.Gaea.Create(&structs.MiniProgramAccessToken{AppID: "fake-app-id", AccessToken: "ExpiredAccessToken", ExpiresIn: -1000})

		// 测试记录已经超过有效期，调用获取函数获取一条 AccessToken 并返回
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		GetMiniProgramAccessToken(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Error\":\"Get \\\"https://api.weixin.qq.com/cgi-bin/token?appid=fake-app-id\\u0026grant_type=client_credential\\u0026secret=fake-secret\\\": no responder found\",\"Message\":\"发送请求失败\"}")

		// 添加一条有效的记录
		orm.MySQL.Gaea.Create(&structs.MiniProgramAccessToken{AppID: "fake-app-id", AccessToken: "AccessToken", ExpiresIn: 1000})

		// 测试记录没有超过有效期，直接返回
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		GetMiniProgramAccessToken(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Data\":\"AccessToken\",\"Message\":\"Success\"}")
	})
}

// Test_requestMiniProgramAccessToken 测试 requestMiniProgramAccessToken 是否可以 获取一条新的 微信小程序 Access Token 保存到数据库后返回
func Test_requestMiniProgramAccessToken(t *testing.T) {
	Convey("测试 requestMiniProgramAccessToken 是否可以 获取一条新的 微信小程序 Access Token 保存到数据库后返回", t, func() {
		// 配置 httpmock 进行拦截
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder(http.MethodGet, "https://api.weixin.qq.com/cgi-bin/token?appid=fakeAppID&grant_type=client_credential&secret=fakeSecret", httpmock.NewBytesResponder(http.StatusInternalServerError, nil))

		// 测试发送请求失败
		statusCode, responseData := requestMiniProgramAccessToken("fakeAppID", "fakeSecret")
		So(statusCode, ShouldEqual, http.StatusInternalServerError)
		So(fmt.Sprint(responseData), ShouldEqual, "map[Error:unexpected end of JSON input Message:发送请求失败]")

		// 调整 httpmock 返回错误的响应
		httpmock.RegisterResponder(http.MethodGet, "https://api.weixin.qq.com/cgi-bin/token?appid=fakeAppID&grant_type=client_credential&secret=fakeSecret", httpmock.NewJsonResponderOrPanic(http.StatusOK, response.Struct{"errcode": 40013, "errmsg": "invalid appid"}))

		// 测试返回了错误信息
		statusCode, responseData = requestMiniProgramAccessToken("fakeAppID", "fakeSecret")
		So(statusCode, ShouldEqual, http.StatusForbidden)
		So(fmt.Sprint(responseData), ShouldEqual, "map[Message:[ 40013 ] invalid appid]")

		// 定义用于测试的假 AccessToken
		fakeAccessToken := "ACCESS_TOKEN " + time.Now().Format("2006-01-02 15:04:05")

		// 调整 httpmock 返回正确的响应
		httpmock.RegisterResponder(http.MethodGet, "https://api.weixin.qq.com/cgi-bin/token?appid=fakeAppID&grant_type=client_credential&secret=fakeSecret", httpmock.NewJsonResponderOrPanic(http.StatusOK, response.Struct{"access_token": fakeAccessToken, "expires_in": 7200}))

		// 测试返回了正确的信息
		statusCode, responseData = requestMiniProgramAccessToken("fakeAppID", "fakeSecret")
		So(statusCode, ShouldEqual, http.StatusOK)
		So(fmt.Sprint(responseData), ShouldEqual, "map[Data:"+fakeAccessToken+" Message:Success]")

		// 读取数据库中是否保存了信息
		accessTokenInfo := structs.MiniProgramAccessToken{}
		orm.MySQL.Gaea.Last(&accessTokenInfo)
		So(accessTokenInfo.AppID, ShouldEqual, "fakeAppID")
		So(accessTokenInfo.AccessToken, ShouldEqual, fakeAccessToken)
		So(accessTokenInfo.ExpiresIn, ShouldEqual, 7200)
	})
}
