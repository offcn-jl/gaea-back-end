/*
   @Time : 2020/11/6 4:33 下午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : sso_test
   @Software: GoLand
   @Description: 单点登陆模块的接口及其辅助函数 单元测试
*/

package events

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/jarcoal/httpmock"
	"github.com/jinzhu/gorm"
	"github.com/offcn-jl/gaea-back-end/commons/config"
	"github.com/offcn-jl/gaea-back-end/commons/database/orm"
	"github.com/offcn-jl/gaea-back-end/commons/database/structs"
	"github.com/offcn-jl/gaea-back-end/commons/utt"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"testing"
	"time"
)

// 覆盖 orm 库中的 ORM 对象
func init() {
	utt.InitTest() // 初始化测试数据并获取测试所需的上下文
	orm.MySQL.Gaea = utt.ORM
}

// TestSSOGetWechatMiniProgramQrCode 测试 SSOGetWechatMiniProgramQrCode 是否可以获取微信小程序个人后缀二维码
func TestSSOGetWechatMiniProgramQrCode(t *testing.T) {
	Convey("测试 SSOGetWechatMiniProgramQrCode 是否可以获取微信小程序个人后缀二维码", t, func() {
		// 初始化 Request
		utt.GinTestContext.Request, _ = http.NewRequest(http.MethodGet, "/", nil)

		// 测试 未绑定 Query 数据
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		SSOGetWechatMiniProgramQrCode(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Error\":\"Key: 'AppID' Error:Field validation for 'AppID' failed on the 'required' tag\\nKey: 'Page' Error:Field validation for 'Page' failed on the 'required' tag\",\"Message\":\"提交的 Query 查询不正确\"}")

		// 重置 Request 添加 Query 参数
		utt.GinTestContext.Request, _ = http.NewRequest(http.MethodGet, "/?app-id=fake-app-id&page=fake-page", nil)

		// 测试后缀配置错误 ( 此时未配置后缀 )
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		SSOGetWechatMiniProgramQrCode(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"个人后缀不正确\"}")

		// 配置个人后缀参数
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "Suffix", Value: "default"}}

		// 配置 httpmock 进行拦截
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		// 测试创建小程序码失败
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		SSOGetWechatMiniProgramQrCode(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"创建失败, Get \\\"https://api.gaea.jilinoffcn.com/release/services/authentication/mini-program/get/access-token?access-token=\\u0026app-id=fake-app-id\\\": no responder found\"}")

		// 配置 httpmock 返回, 返回小程序码的图片
		httpmock.RegisterResponder(http.MethodGet, "https://api.gaea.jilinoffcn.com/release/services/authentication/mini-program/get/access-token", httpmock.NewStringResponder(http.StatusOK, "{\"Message\":\"Success\",\"Data\":\"fake-access-token\"}"))
		fakeImage := make([]byte, 10)
		httpmock.RegisterResponder(http.MethodPost, "https://api.weixin.qq.com/wxa/getwxacodeunlimit", httpmock.NewBytesResponder(http.StatusOK, fakeImage))

		// 测试创建小程序码成功
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		SSOGetWechatMiniProgramQrCode(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Header().Get("Content-Disposition"), ShouldEqual, "attachment;filename=Wechat-MiniProgram-QrCode-default.jpg")
		So(utt.HttpTestResponseRecorder.Header().Get("Cache-Control"), ShouldEqual, "must-revalidate,post-check=0,pre-check=0")
		So(utt.HttpTestResponseRecorder.Header().Get("Expires"), ShouldEqual, "0")
		So(utt.HttpTestResponseRecorder.Header().Get("Pragma"), ShouldEqual, "public")
		So(string(utt.HttpTestResponseRecorder.Body.Bytes()), ShouldEqual, string(fakeImage))
	})
}

// TestSSOSendVerificationCode 测试 SSOSendVerificationCode 单点登模块发送验证码接口的处理函数
func TestSSOSendVerificationCode(t *testing.T) {
	// 函数推出时重置数据库
	Convey("测试 SSOSendVerificationCode 单点登模块发送验证码接口的处理函数", t, func() {
		// 测试 验证手机号码是否有效
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		SSOSendVerificationCode(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"手机号码不正确\"}")

		// 配置手机号码
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "Phone", Value: "17866886688"}}

		// 测试 获取登陆模块的配置
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		SSOSendVerificationCode(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"登陆模块配置有误\"}")

		// 配置登陆模块 ID
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "Phone", Value: "17866886688"}, gin.Param{Key: "MID", Value: "10001"}}
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		SSOSendVerificationCode(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"登陆模块 SMS 平台配置有误\"}")

		// 配置 SMS 平台为 中公短信平台
		orm.MySQL.Gaea.Model(structs.SingleSignOnLoginModule{}).Update(&structs.SingleSignOnLoginModule{Platform: 1})
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		SSOSendVerificationCode(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"短信模板配置有误\"}")

		// 配置 SMS 平台为 腾讯云
		orm.MySQL.Gaea.Model(structs.SingleSignOnLoginModule{}).Update(&structs.SingleSignOnLoginModule{Platform: 2})
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		SSOSendVerificationCode(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"未配置 腾讯云 令牌\"}")
	})
}

// TestSSOSignUp 测试 SSOSignUp 单点登模块注册接口的处理函数
func TestSSOSignUp(t *testing.T) {
	Convey("测试 SSOSignUp 单点登模块注册接口的处理函数", t, func() {
		// 测试 未绑定 Body 数据
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		SSOSignUp(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Error\":\"invalid request\",\"Message\":\"提交的 Json 数据不正确\"}")

		// 增加 Body
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"MID\":10000,\"Phone\":\"*\"}"))

		// 测试 验证手机号码是否有效
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		SSOSignUp(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"手机号码不正确\"}")

		// 修正手机号码
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"MID\":10000,\"Phone\":\"17866668888\"}"))

		// 测试 校验登录模块配置
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		SSOSignUp(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"单点登陆模块配置有误\"}")

		// 修正为已存在的登陆模块 ID, 未发送过验证码的手机号码
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"MID\":10001,\"Phone\":\"17866660001\"}"))

		// 测试 校验是否发送过验证码
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		SSOSignUp(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"请您先获取验证码后再进行注册\"}")

		// 修正为模拟获取过验证码的手机号码
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"MID\":10001,\"Phone\":\"17888886666\"}"))

		// 测试 校验验证码是正确 ( 此时的上下文中未填写验证码 )
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		SSOSignUp(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"验证码有误\"}")

		// 向请求中添加正确, 但已经失效的验证码
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"MID\":10001,\"Phone\":\"17888886666\",\"Code\":9999}"))

		// 测试 校验验证码是否有效
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		SSOSignUp(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"验证码失效\"}")

		// 更换上下文中的手机号码及验证码未为正确且未失效的内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"MID\":10001,\"Phone\":\"17866886688\",\"Code\":9999}"))

		// 测试 注册成功
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		SSOSignUp(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"Success\"}")

		// 判断用户是否存在
		So(ssoIsSignUp("17866886688"), ShouldBeTrue)

		// 判断会话是否存在
		So(ssoIsSignIn("17866886688", 10001), ShouldBeTrue)
	})
}

// TestSSOSignIn 测试 SSOSignIn 单点登陆模块登陆接口的处理函数
func TestSSOSignIn(t *testing.T) {
	utt.ResetContext() // 重置测试上下文

	Convey("测试 SSOSignUp 单点登陆模块登陆接口的处理函数", t, func() {
		// 测试 未绑定 Body 数据
		SSOSignIn(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Error\":\"invalid request\",\"Message\":\"提交的 Json 数据不正确\"}")

		// 增加 Body
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"MID\":10000,\"Phone\":\"*\"}"))

		// 测试 验证手机号码是否有效
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		SSOSignIn(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"手机号码不正确\"}")

		// 修正手机号码
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"MID\":10000,\"Phone\":\"17866668888\"}"))

		// 测试 校验登录模块配置 ( 此时的上下文中没有登陆模块的配置 )
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		SSOSignIn(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"单点登陆模块配置有误\"}")

		// 修正为已存在的登陆模块 ID
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"MID\":10001,\"Phone\":\"17866668888\"}"))

		// 测试 校验用户是否已经注册
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		SSOSignIn(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"请您先进行注册\"}")

		// 修正为已经注册的手机号码
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"MID\":10001,\"Phone\":\"17888666688\"}"))

		// 测试 登陆成功
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		SSOSignIn(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"Success\"}")

		// 判断会话是否存在
		So(ssoIsSignIn("17888666688", 10001), ShouldBeTrue)
	})
}

// TestSSOGetSessionInfo 测试 SSOGetSessionInfo 函数是否可以获取会话信息
func TestSSOGetSessionInfo(t *testing.T) {
	utt.ResetContext() // 重置测试上下文

	Convey("测试 GetSessionInfo 函数是否可以获取会话信息", t, func() {
		Convey("测试 验证手机号是否有效 ( 需要是 0 或正确的手机号 )", func() {
			// 此时没有填写手机号
			SSOSessionInfo(utt.GinTestContext)
			So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "手机号码不正确")
			// 此时填写了错误的手机号
			utt.GinTestContext.Params = gin.Params{gin.Param{Key: "Phone", Value: "1788710666"}}
			utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
			SSOSessionInfo(utt.GinTestContext)
			So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "手机号码不正确")
		})

		// 修正为正确的手机号码, 供后续测试使用
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "Phone", Value: "17887106666"}}

		Convey("测试 校验登陆模块配置", func() {
			// 此时没有填写登陆模块配置
			utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
			SSOSessionInfo(utt.GinTestContext)
			So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "单点登陆模块配置有误")
			// 此时填写了错误的登陆模块 ID
			utt.GinTestContext.Params = gin.Params{gin.Param{Key: "Phone", Value: "17887106666"}, gin.Param{Key: "MID", Value: "20001"}}
			utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
			SSOSessionInfo(utt.GinTestContext)
			So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "单点登陆模块配置有误")
		})

		// 修正为正确的登陆模块 ID
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "Phone", Value: "17887106666"}, gin.Param{Key: "MID", Value: "10001"}}

		Convey("测试 后缀不存在或错误", func() {
			defaultInfo := "{\"Data\":{\"Sign\":\"中公教育\",\"CRMEID\":\"HD202010142576\",\"CRMEFID\":56975,\"CRMEFSID\":\"f905e07b2bff94d564ac1fa41022a633\",\"CRMChannel\":7,\"CRMOCode\":22,\"CRMOName\":\"吉林分校\",\"CRMUID\":32431,\"CRMUser\":\"default\",\"Suffix\":\"default\",\"NTalkerGID\":\"NTalkerGID\",\"IsLogin\":false,\"NeedToRegister\":true},\"Message\":\"Success\"}"
			// 此时未配置后缀, 即后缀不存在, 可以认为等同为后缀错误
			utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
			SSOSessionInfo(utt.GinTestContext)
			So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, defaultInfo)
			// 设置设置默认后缀 CRM 组织 ID 为 0
			orm.MySQL.Gaea.Model(structs.SingleSignOnSuffix{}).Where("suffix = 'default'").Update("crm_oid", "0")
			utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
			SSOSessionInfo(utt.GinTestContext)
			So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, defaultInfo)
			// 设置默认后缀 CRM 组织 ID 为不存在的 ID
			orm.MySQL.Gaea.Model(structs.SingleSignOnSuffix{}).Where("suffix = 'default'").Update("crm_oid", "10000")
			utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
			SSOSessionInfo(utt.GinTestContext)
			So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, defaultInfo)
			// 还原默认后缀 CRM 组织 ID
			orm.MySQL.Gaea.Model(structs.SingleSignOnSuffix{}).Where("suffix = 'default'").Update("crm_oid", "1")
		})

		Convey("测试 配置了后缀 是否可以获取到后缀对应的信息", func() {
			testInfo := "{\"Data\":{\"Sign\":\"中公教育\",\"CRMEID\":\"HD202010142576\",\"CRMEFID\":56975,\"CRMEFSID\":\"f905e07b2bff94d564ac1fa41022a633\",\"CRMChannel\":22,\"CRMOCode\":2290,\"CRMOName\":\"吉林长春分校\",\"CRMUID\":123,\"CRMUser\":\"test\",\"Suffix\":\"test\",\"NTalkerGID\":\"NTalkerGID\",\"IsLogin\":false,\"NeedToRegister\":true},\"Message\":\"Success\"}"
			testInfoWithDefauleOrgnation := "{\"Data\":{\"Sign\":\"中公教育\",\"CRMEID\":\"HD202010142576\",\"CRMEFID\":56975,\"CRMEFSID\":\"f905e07b2bff94d564ac1fa41022a633\",\"CRMChannel\":22,\"CRMOCode\":22,\"CRMOName\":\"吉林分校\",\"CRMUID\":123,\"CRMUser\":\"test\",\"Suffix\":\"test\",\"NTalkerGID\":\"NTalkerGID\",\"IsLogin\":false,\"NeedToRegister\":true},\"Message\":\"Success\"}"
			// 配置后缀为测试后缀
			utt.GinTestContext.Params = gin.Params{gin.Param{Key: "Phone", Value: "17887106666"}, gin.Param{Key: "MID", Value: "10001"}, gin.Param{Key: "Suffix", Value: "test"}}
			utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
			SSOSessionInfo(utt.GinTestContext)
			So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, testInfo)
			// 设置测试后缀 CRM 组织 ID 为 0
			orm.MySQL.Gaea.Unscoped().Model(structs.SingleSignOnSuffix{}).Where("suffix = 'test'").Update("crm_oid", "0")
			utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
			SSOSessionInfo(utt.GinTestContext)
			So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, testInfoWithDefauleOrgnation)
			// 设置测试后缀 CRM 组织 ID 为不存在的 ID
			orm.MySQL.Gaea.Unscoped().Model(structs.SingleSignOnSuffix{}).Where("suffix = 'test'").Update("crm_oid", "10000")
			utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
			SSOSessionInfo(utt.GinTestContext)
			So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, testInfoWithDefauleOrgnation)
			// 还原测试后缀 CRM 组织 ID
			orm.MySQL.Gaea.Unscoped().Model(structs.SingleSignOnSuffix{}).Where("suffix = 'test'").Update("crm_oid", "2")
		})

		Convey("测试 配置了已经过期的后缀 是否依旧可以获取到后缀对应的信息", func() {
			testInfo := "{\"Data\":{\"Sign\":\"中公教育\",\"CRMEID\":\"HD202010142576\",\"CRMEFID\":56975,\"CRMEFSID\":\"f905e07b2bff94d564ac1fa41022a633\",\"CRMChannel\":22,\"CRMOCode\":2290,\"CRMOName\":\"吉林长春分校\",\"CRMUID\":123,\"CRMUser\":\"expired\",\"Suffix\":\"expired\",\"NTalkerGID\":\"NTalkerGID\",\"IsLogin\":false,\"NeedToRegister\":true},\"Message\":\"Success\"}"
			testInfoWithDefauleOrgnation := "{\"Data\":{\"Sign\":\"中公教育\",\"CRMEID\":\"HD202010142576\",\"CRMEFID\":56975,\"CRMEFSID\":\"f905e07b2bff94d564ac1fa41022a633\",\"CRMChannel\":22,\"CRMOCode\":22,\"CRMOName\":\"吉林分校\",\"CRMUID\":123,\"CRMUser\":\"expired\",\"Suffix\":\"expired\",\"NTalkerGID\":\"NTalkerGID\",\"IsLogin\":false,\"NeedToRegister\":true},\"Message\":\"Success\"}"
			// 配置后缀为过期后缀
			utt.GinTestContext.Params = gin.Params{gin.Param{Key: "Phone", Value: "17887106666"}, gin.Param{Key: "MID", Value: "10001"}, gin.Param{Key: "Suffix", Value: "expired"}}
			utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
			SSOSessionInfo(utt.GinTestContext)
			So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, testInfo)
			// 设置过期后缀 CRM 组织 ID 为 0
			orm.MySQL.Gaea.Unscoped().Model(structs.SingleSignOnSuffix{}).Where("suffix = 'expired'").Update("crm_oid", "0")
			utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
			SSOSessionInfo(utt.GinTestContext)
			So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, testInfoWithDefauleOrgnation)
			// 设置过期后缀 CRM 组织 ID 为不存在的 ID
			orm.MySQL.Gaea.Unscoped().Model(structs.SingleSignOnSuffix{}).Where("suffix = 'expired'").Update("crm_oid", "10000")
			utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
			SSOSessionInfo(utt.GinTestContext)
			So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, testInfoWithDefauleOrgnation)
			// 还原过期后缀 CRM 组织 ID
			orm.MySQL.Gaea.Unscoped().Model(structs.SingleSignOnSuffix{}).Where("suffix = 'expired'").Update("crm_oid", "2")
		})

		Convey("测试 校验是否需要注册", func() {
			// 模拟注册手机号
			userInfo := structs.SingleSignOnUser{}
			userInfo.Phone = utt.GetFakePhone()
			orm.MySQL.Gaea.Create(&userInfo)
			// 更换手机号为已经注册的手机号
			utt.GinTestContext.Params = gin.Params{gin.Param{Key: "Phone", Value: userInfo.Phone}, gin.Param{Key: "MID", Value: "10001"}, gin.Param{Key: "Suffix", Value: "test"}}
			utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
			SSOSessionInfo(utt.GinTestContext)
			So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\"NeedToRegister\":false")
		})

		Convey("测试 校验是否需要登陆", func() {
			// 模拟注册手机号
			userInfo := structs.SingleSignOnUser{
				Phone: utt.GetFakePhone(),
			}
			orm.MySQL.Gaea.Create(&userInfo)
			sessionInfo := structs.SingleSignOnSession{
				MID:   10001,
				Phone: userInfo.Phone,
			}
			orm.MySQL.Gaea.Create(&sessionInfo)
			// 更换手机号为已经注册的手机号
			utt.GinTestContext.Params = gin.Params{gin.Param{Key: "Phone", Value: userInfo.Phone}, gin.Param{Key: "MID", Value: "10001"}, gin.Param{Key: "Suffix", Value: "test"}}
			utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
			SSOSessionInfo(utt.GinTestContext)
			So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\"IsLogin\":true")
		})
	})
}

// Test_sendVerificationCodeByOFFCN 测试 sendVerificationCodeByOFFCN 是否可以使用中公短信平台发送验证码
func Test_sendVerificationCodeByOFFCN(t *testing.T) {
	Convey("测试 sendVerificationCodeByOFFCN 是否可以使用中公短信平台发送验证码", t, func() {
		// 测试验证配置
		// 此时没有任何配置
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		sendVerificationCodeByOFFCN(utt.GinTestContext, 0, 0)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"短信模板配置有误\"}")
		// 配置短信模板
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		sendVerificationCodeByOFFCN(utt.GinTestContext, 392074, 0)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"未配置 中公教育短信平台 接口地址\"}")
		// 配置 中公教育短信平台 接口地址
		config.Update(orm.MySQL.Gaea, structs.SystemConfig{OffcnSmsURL: "https://fake-sms-platform.offcn.com"})
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		sendVerificationCodeByOFFCN(utt.GinTestContext, 392074, 0)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"未配置 中公教育短信平台 用户名\"}")
		// 配置 中公教育短信平台 用户名
		config.Update(orm.MySQL.Gaea, structs.SystemConfig{OffcnSmsURL: "https://fake-sms-platform.offcn.com", OffcnSmsUserName: "fake-sms-user"})
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		sendVerificationCodeByOFFCN(utt.GinTestContext, 392074, 0)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"未配置 中公教育短信平台 密码\"}")
		// 配置 中公教育短信平台 密码
		config.Update(orm.MySQL.Gaea, structs.SystemConfig{OffcnSmsURL: "https://fake-sms-platform.offcn.com", OffcnSmsUserName: "fake-sms-user", OffcnSmsPassword: "fake-sms-password"})
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		sendVerificationCodeByOFFCN(utt.GinTestContext, 392074, 0)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"未配置 中公教育短信平台 发送方识别码\"}")
		// 配置 中公教育短信平台 发送方识别码
		config.Update(orm.MySQL.Gaea, structs.SystemConfig{OffcnSmsURL: "https://fake-sms-platform.offcn.com", OffcnSmsUserName: "fake-sms-user", OffcnSmsPassword: "fake-sms-password", OffcnSmsTjCode: "fake-sms-tj-code"})

		// 配置手机号码
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "Phone", Value: "17866886688"}}

		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		sendVerificationCodeByOFFCN(utt.GinTestContext, 392074, 0)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"请勿重复发送验证码\"}")

		// 配置为未发送过短信的手机号码
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "Phone", Value: utt.GetFakePhone()}}

		// 配置 httpmock 进行拦截
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		sendVerificationCodeByOFFCN(utt.GinTestContext, 392074, 0)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"发送短信失败, 返回错误 : Post \\\"https://fake-sms-platform.offcn.com\\\": no responder found\"}")

		// 配置 httpmock 返回 非 200 的 http 状态码
		httpmock.RegisterResponder(http.MethodPost, "https://fake-sms-platform.offcn.com", httpmock.NewStringResponder(http.StatusInternalServerError, ""))
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		sendVerificationCodeByOFFCN(utt.GinTestContext, 392074, 0)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"发送短信失败, 返回状态码 : 500\"}")

		// 配置 httpmock 返回 无法读取的 body todo 暂时无法测试

		// 配置 httpmock 返回 错误的 Json
		httpmock.RegisterResponder(http.MethodPost, "https://fake-sms-platform.offcn.com", httpmock.NewBytesResponder(http.StatusOK, []byte("")))
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		sendVerificationCodeByOFFCN(utt.GinTestContext, 392074, 0)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"发送短信失败, 解码返回内容失败, 错误内容 : unexpected end of JSON input\"}")

		// 配置 httpmock 返回 发送失败的 Json
		httpmock.RegisterResponder(http.MethodPost, "https://fake-sms-platform.offcn.com", httpmock.NewJsonResponderOrPanic(http.StatusOK, gin.H{"status": 0, "msg": "fail"}))
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		sendVerificationCodeByOFFCN(utt.GinTestContext, 392074, 0)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"发送短信失败, [ 0 ] fail\"}")

		// 配置 httpmock 返回 发送成功 Json
		httpmock.RegisterResponder(http.MethodPost, "https://fake-sms-platform.offcn.com", httpmock.NewJsonResponderOrPanic(http.StatusOK, gin.H{"status": 1}))
		utt.HttpTestResponseRecorder.Body.Reset()                                  // 再次测试前重置 body
		utt.GinTestContext.Request, _ = http.NewRequest(http.MethodPost, "/", nil) // 初始化 Request 避免 utt.GinTestContext.ClientIP() 出现空指针错误
		sendVerificationCodeByOFFCN(utt.GinTestContext, 392074, 10)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"Success\"}")

		// 测试其他短信模板
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		sendVerificationCodeByOFFCN(utt.GinTestContext, 392030, 10)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"请勿重复发送验证码\"}")
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		sendVerificationCodeByOFFCN(utt.GinTestContext, 391863, 10)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"请勿重复发送验证码\"}")

		// 检查发送记录
		singleSignOnVerificationCode := structs.SingleSignOnVerificationCode{}
		orm.MySQL.Gaea.Find(&singleSignOnVerificationCode)
		So(singleSignOnVerificationCode.Phone, ShouldEqual, utt.GinTestContext.Param("Phone"))
		So(singleSignOnVerificationCode.Term, ShouldEqual, 10)
		So(singleSignOnVerificationCode.SourceIP, ShouldEqual, "")
	})
}

// Test_sendVerificationCodeByTencentCloudSMSV2 测试 sendVerificationCodeByTencentCloudSMSV2 使用腾讯云短信平台发送验证码
func Test_sendVerificationCodeByTencentCloudSMSV2(t *testing.T) {
	Convey("测试 sendVerificationCodeByTencentCloudSMSV2 使用腾讯云短信平台发送验证码", t, func() {
		// 测试验证配置
		// 此时没有任何配置
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		sendVerificationCodeByTencentCloudSMSV2(utt.GinTestContext, "", 0, 0)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"未配置 腾讯云 令牌\"}")

		// 配置 腾讯云 令牌
		config.Update(orm.MySQL.Gaea, structs.SystemConfig{TencentCloudAPISecretID: "fake-tencent-cloud-api-secret-id"})
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		sendVerificationCodeByTencentCloudSMSV2(utt.GinTestContext, "", 0, 0)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"未配置 腾讯云 密钥\"}")
		// 配置 腾讯云 密钥
		config.Update(orm.MySQL.Gaea, structs.SystemConfig{TencentCloudAPISecretID: "fake-tencent-cloud-api-secret-id", TencentCloudAPISecretKey: "fake-tencent-cloud-api-secret-key"})
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		sendVerificationCodeByTencentCloudSMSV2(utt.GinTestContext, "", 0, 0)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"未配置 腾讯云 短信应用 ID\"}")
		// 配置 腾讯云 短信应用 ID
		config.Update(orm.MySQL.Gaea, structs.SystemConfig{TencentCloudAPISecretID: "fake-tencent-cloud-api-secret-id", TencentCloudAPISecretKey: "fake-tencent-cloud-api-secret-key", TencentCloudSmsSdkAppId: "fake-tencent-cloud-sms-sdk-app-id"})

		// 配置手机号码
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "Phone", Value: "17866886688"}}

		// 测试重复发送验证码
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		sendVerificationCodeByTencentCloudSMSV2(utt.GinTestContext, "", 0, 0)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"请勿重复发送验证码\"}")

		// 配置为未发送过短信的手机号码
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "Phone", Value: utt.GetFakePhone()}}

		// 配置 httpmock 进行拦截
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		// 测试发送短信失败
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		sendVerificationCodeByTencentCloudSMSV2(utt.GinTestContext, "", 0, 0)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"发送短信失败, [ ClientError.NetworkError ] Fail to get response because Post \\\"https://sms.internal.tencentcloudapi.com/\\\": no responder found\"}")

		// 测试 非 SDK 异常 todo 暂时无法测试

		// 配置 httpmock 返回 发送失败的 Json
		// https://cloud.tencent.com/document/product/382/38770
		httpmock.RegisterNoResponder(httpmock.NewJsonResponderOrPanic(http.StatusOK, gin.H{"Response": gin.H{"Error": gin.H{"Code": "AuthFailure.SignatureFailure", "Message": "The provided credentials could not be validated. Please check your signature is correct."}, "RequestId": "ed93f3cb-f35e-473f-b9f3-0d451b8b79c6"}}))
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		sendVerificationCodeByTencentCloudSMSV2(utt.GinTestContext, "", 0, 0)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"发送短信失败, [ AuthFailure.SignatureFailure ] The provided credentials could not be validated. Please check your signature is correct.\"}")

		// 配置 httpmock 返回 发送失败的 Json
		// https://cloud.tencent.com/document/product/382/38770
		httpmock.RegisterNoResponder(httpmock.NewJsonResponderOrPanic(http.StatusOK, gin.H{"Response": gin.H{"SendStatusSet": []gin.H{{"Code": "FailCode", "Message": "FailMessage"}}}}))
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		sendVerificationCodeByTencentCloudSMSV2(utt.GinTestContext, "", 0, 10)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"发送短信失败, 错误内容 : FailMessage\"}")

		// 配置 httpmock 返回 发送成功的 Json
		// https://cloud.tencent.com/document/product/382/38770
		httpmock.RegisterNoResponder(httpmock.NewJsonResponderOrPanic(http.StatusOK, gin.H{"Response": gin.H{"SendStatusSet": []gin.H{{"Code": "Ok"}}}}))
		utt.HttpTestResponseRecorder.Body.Reset()                                  // 再次测试前重置 body
		utt.GinTestContext.Request, _ = http.NewRequest(http.MethodPost, "/", nil) // 初始化 Request 避免 utt.GinTestContext.ClientIP() 出现空指针错误
		sendVerificationCodeByTencentCloudSMSV2(utt.GinTestContext, "", 0, 10)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"Success\"}")

		// 检查发送记录
		singleSignOnVerificationCode := structs.SingleSignOnVerificationCode{}
		orm.MySQL.Gaea.Find(&singleSignOnVerificationCode)
		So(singleSignOnVerificationCode.Phone, ShouldEqual, utt.GinTestContext.Param("Phone"))
		So(singleSignOnVerificationCode.Term, ShouldEqual, 10)
		So(singleSignOnVerificationCode.SourceIP, ShouldEqual, "")
	})
}

// Test_ssoIsSignUp 测试 ssoIsSignUp 函数是否可以检查用户是否已经注册且未失效
func Test_ssoIsSignUp(t *testing.T) {
	Convey("测试 ssoIsSignUp 函数是否可以检查用户是否已经注册且未失效", t, func() {
		// 测试用户不存在的情况
		So(ssoIsSignUp("17887106666"), ShouldBeFalse)
		// 创建会话, 会话时间为 30 天前
		orm.MySQL.Gaea.Create(&structs.SingleSignOnUser{Model: gorm.Model{CreatedAt: time.Now().AddDate(0, 0, -31)}, Phone: "17887106666"})
		// 测试注册已经过期的情况
		So(ssoIsSignUp("17887106666"), ShouldBeFalse)
		// 创建会话, 会话时间为昨天
		orm.MySQL.Gaea.Create(&structs.SingleSignOnUser{Model: gorm.Model{CreatedAt: time.Now().AddDate(0, 0, -1)}, Phone: "17887108888"})
		// 测试注册有效的情况
		So(ssoIsSignUp("17887108888"), ShouldBeTrue)
	})
}

// Test_ssoIsSignIn 测试 ssoIsSignIn 函数是否可以检查用户是否已经登陆
func Test_ssoIsSignIn(t *testing.T) {
	Convey("测试 ssoIsSignIn 函数是否可以检查用户是否已经登陆", t, func() {
		// 测试会话不存在时的情况
		So(ssoIsSignIn("17887106666", 1), ShouldBeFalse)
		// 创建会话
		orm.MySQL.Gaea.Create(&structs.SingleSignOnSession{MID: 1, Phone: "17887106666"})
		// 测试会话存在时的情况
		So(ssoIsSignIn("17887106666", 1), ShouldBeTrue)
	})
}

// Test_ssoCreateSession 测试 ssoCreateSession 函数是否可以按照预期矫正信息后创建会话
func Test_ssoCreateSession(t *testing.T) {
	Convey("测试 ssoCreateSession 函数是否可以按照预期矫正信息后创建会话", t, func() {
		// 创建测试会话
		session := structs.SingleSignOnSession{}
		session.Phone = utt.GetFakePhone()
		session.CRMEFSID = "f905e07b2bff94d564ac1fa41022a633" // 测试 CRM 活动表单

		session.CustomerName = "测试姓名"
		session.CustomerIdentityID = 1 // 在校生-大一
		session.CustomerColleage = "测试学校"
		session.CustomerMayor = "测试专业"
		session.Remark = "测试备注"

		// 校验校正前的数据
		So(session.ActualSuffix, ShouldEqual, "")  // 后缀
		So(session.CurrentSuffix, ShouldEqual, "") // 校正后的后缀
		So(session.CRMChannel, ShouldEqual, 0)     // CRM 所属渠道
		So(session.CRMUID, ShouldEqual, 0)         // CRM 用户 ID
		So(session.CRMOCode, ShouldEqual, 0)       // CRM 组织代码

		// 验证未配置后缀
		ssoCreateSession(&session)
		// 校验校正后的数据
		So(session.ActualSuffix, ShouldEqual, "")         // 后缀
		So(session.CurrentSuffix, ShouldEqual, "default") // 校正后的后缀
		So(session.CRMChannel, ShouldEqual, 7)            // CRM 所属渠道
		So(session.CRMUID, ShouldEqual, 32431)            // CRM 用户 ID
		So(session.CRMOCode, ShouldEqual, 2290)           // CRM 组织代码

		// 验证无效后缀
		session.ID = 0 // 将 ID 恢复为 0, 令 ORM 认为这条 Session 是新记录
		session.ActualSuffix = "invalid_suffix"
		session.Phone = utt.GetFakePhone()
		ssoCreateSession(&session)
		// 校验测试后缀信息
		So(session.ActualSuffix, ShouldEqual, "invalid_suffix") // 后缀
		So(session.CurrentSuffix, ShouldEqual, "default")       // 校正后的后缀
		So(session.CRMChannel, ShouldEqual, 7)                  // CRM 所属渠道
		So(session.CRMUID, ShouldEqual, 32431)                  // CRM 用户 ID
		So(session.CRMOCode, ShouldEqual, 2290)                 // CRM 组织代码

		// 验证默认后缀
		session.ID = 0 // 将 ID 恢复为 0, 令 ORM 认为这条 Session 是新记录
		session.ActualSuffix = "default"
		session.Phone = utt.GetFakePhone()
		ssoCreateSession(&session)
		// 校验测试后缀信息
		So(session.ActualSuffix, ShouldEqual, "default")  // 后缀
		So(session.CurrentSuffix, ShouldEqual, "default") // 校正后的后缀
		So(session.CRMChannel, ShouldEqual, 7)            // CRM 所属渠道
		So(session.CRMUID, ShouldEqual, 32431)            // CRM 用户 ID
		So(session.CRMOCode, ShouldEqual, 2290)           // CRM 组织代码

		// 验证测试后缀
		session.ID = 0 // 将 ID 恢复为 0, 令 ORM 认为这条 Session 是新记录
		session.ActualSuffix = "test"
		session.Phone = utt.GetFakePhone()
		ssoCreateSession(&session)
		// 校验测试后缀信息
		So(session.ActualSuffix, ShouldEqual, "test")  // 后缀
		So(session.CurrentSuffix, ShouldEqual, "test") // 校正后的后缀
		So(session.CRMChannel, ShouldEqual, 22)        // CRM 所属渠道
		So(session.CRMUID, ShouldEqual, 123)           // CRM 用户 ID
		So(session.CRMOCode, ShouldEqual, 2290)        // CRM 组织代码

		// 验证已经过期的后缀是否依旧有效
		session.ID = 0 // 将 ID 恢复为 0, 令 ORM 认为这条 Session 是新记录
		session.ActualSuffix = "expired"
		session.Phone = utt.GetFakePhone()
		ssoCreateSession(&session)
		// 校验测试后缀信息
		So(session.ActualSuffix, ShouldEqual, "expired")  // 后缀
		So(session.CurrentSuffix, ShouldEqual, "expired") // 校正后的后缀
		So(session.CRMChannel, ShouldEqual, 22)           // CRM 所属渠道
		So(session.CRMUID, ShouldEqual, 123)              // CRM 用户 ID
		So(session.CRMOCode, ShouldEqual, 2290)           // CRM 组织代码

		// 配置 httpmock 进行拦截
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		// 测试 发送 GET 请求出错 ( 开启 httpmock 后，未配置监听器时默认返回连接失败错误 )
		//httpmock.RegisterResponder(http.MethodGet, "https://dc.com:8443/a.gif", httpmock.ConnectionFailure)
		session.ID = 0 // 将 ID 恢复为 0, 令 ORM 认为这条 Session 是新记录
		session.ActualSuffix = "default"
		session.Phone = utt.GetFakePhone()
		ssoCreateSession(&session)
		// 校验测试后缀信息
		So(session.ActualSuffix, ShouldEqual, "default")  // 后缀
		So(session.CurrentSuffix, ShouldEqual, "default") // 校正后的后缀
		So(session.CRMChannel, ShouldEqual, 7)            // CRM 所属渠道
		So(session.CRMUID, ShouldEqual, 32431)            // CRM 用户 ID
		So(session.CRMOCode, ShouldEqual, 2290)           // CRM 组织代码
		// 查询推送错误日志
		singleSignOnErrorLog := structs.SingleSignOnErrorLog{}
		orm.MySQL.Gaea.Find(&singleSignOnErrorLog)
		So(singleSignOnErrorLog.Phone, ShouldEqual, session.Phone)
		So(singleSignOnErrorLog.MID, ShouldEqual, session.MID)
		So(singleSignOnErrorLog.CRMChannel, ShouldEqual, session.CRMChannel)
		So(singleSignOnErrorLog.CRMUID, ShouldEqual, session.CRMUID)
		So(singleSignOnErrorLog.CRMOCode, ShouldEqual, session.CRMOCode)
		So(singleSignOnErrorLog.Error, ShouldContainSubstring, "Get \"https://dc.offcn.com:8443/a.gif?")
		So(singleSignOnErrorLog.Error, ShouldContainSubstring, ": no responder found")

		// 验证 推送失败 时, 返回会话信息并保存日志
		// 匹配 URL
		httpmock.RegisterResponder(http.MethodGet, "https://dc.offcn.com:8443/a.gif", httpmock.NewStringResponder(http.StatusInternalServerError, ""))
		session.ID = 0 // 将 ID 恢复为 0, 令 ORM 认为这条 Session 是新记录
		session.ActualSuffix = "default"
		session.Phone = utt.GetFakePhone()
		ssoCreateSession(&session)
		// 校验测试后缀信息
		So(session.ActualSuffix, ShouldEqual, "default")  // 后缀
		So(session.CurrentSuffix, ShouldEqual, "default") // 校正后的后缀
		So(session.CRMChannel, ShouldEqual, 7)            // CRM 所属渠道
		So(session.CRMUID, ShouldEqual, 32431)            // CRM 用户 ID
		So(session.CRMOCode, ShouldEqual, 2290)           // CRM 组织代码
		// 查询推送错误日志
		singleSignOnErrorLog = structs.SingleSignOnErrorLog{}
		orm.MySQL.Gaea.Find(&singleSignOnErrorLog)
		So(singleSignOnErrorLog.Phone, ShouldEqual, session.Phone)
		So(singleSignOnErrorLog.MID, ShouldEqual, session.MID)
		So(singleSignOnErrorLog.CRMChannel, ShouldEqual, session.CRMChannel)
		So(singleSignOnErrorLog.CRMUID, ShouldEqual, session.CRMUID)
		So(singleSignOnErrorLog.CRMOCode, ShouldEqual, session.CRMOCode)
		So(singleSignOnErrorLog.Error, ShouldEqual, "CRM 响应状态码 : 500")
	})
}

// TestSSOGetDefaultSuffix 测试 SSOGetDefaultSuffix 函数是否可以获取默认后缀配置
func TestSSOGetDefaultSuffix(t *testing.T) {
	Convey("测试 ssoGetDefaultSuffix 函数是否可以获取默认后缀配置", t, func() {
		// 创建测试会话
		session := structs.SingleSignOnSession{}
		session.Phone = "17887106666"
		So(session.CRMChannel, ShouldEqual, 0)
		So(session.CRMOCode, ShouldEqual, 0)
		SSOGetDefaultSuffix(&session)
		So(session.CRMChannel, ShouldEqual, 7)
		So(session.CRMOCode, ShouldEqual, 2290)
		So(session.CurrentSuffix, ShouldEqual, "default")
		So(session.CRMUID, ShouldEqual, 32431)
	})
}

// TestSSODistributionByPhoneNumber 测试 SSODistributionByPhoneNumber 函数是否可以按照手机号码归属地进行归属分部分配
// 号段数据来自 http://wwutt.HttpTestResponseRecorder.bixinshui.com
func TestSSODistributionByPhoneNumber(t *testing.T) {
	Convey("测试 ssoDistributionByPhoneNumber 函数是否可以按照手机号码归属地进行归属分部分配", t, func() {
		// 创建测试会话
		session := structs.SingleSignOnSession{}

		// 解析出错
		session.Phone = "9999"
		session.CRMOCode = 0
		So(session.CRMOCode, ShouldEqual, 0)
		SSODistributionByPhoneNumber(&session)
		So(session.CRMOCode, ShouldNotEqual, 0)

		// 长春
		session.Phone = "17887106666"
		session.CRMOCode = 0
		So(session.CRMOCode, ShouldEqual, 0)
		SSODistributionByPhoneNumber(&session)
		So(session.CRMOCode, ShouldEqual, 2290)

		// 吉林
		session.Phone = "13009156666"
		session.CRMOCode = 0
		So(session.CRMOCode, ShouldEqual, 0)
		SSODistributionByPhoneNumber(&session)
		So(session.CRMOCode, ShouldEqual, 2305)

		// 延边
		session.Phone = "18943306666"
		session.CRMOCode = 0
		So(session.CRMOCode, ShouldEqual, 0)
		SSODistributionByPhoneNumber(&session)
		So(session.CRMOCode, ShouldEqual, 2277)

		// 通化
		session.Phone = "13009196666"
		session.CRMOCode = 0
		So(session.CRMOCode, ShouldEqual, 0)
		SSODistributionByPhoneNumber(&session)
		So(session.CRMOCode, ShouldEqual, 2271)

		// 白山
		session.Phone = "13009076666"
		session.CRMOCode = 0
		So(session.CRMOCode, ShouldEqual, 0)
		SSODistributionByPhoneNumber(&session)
		So(session.CRMOCode, ShouldEqual, 2310)

		// 四平
		session.Phone = "13009026666"
		session.CRMOCode = 0
		So(session.CRMOCode, ShouldEqual, 0)
		SSODistributionByPhoneNumber(&session)
		So(session.CRMOCode, ShouldEqual, 2263)

		// 松原
		session.Phone = "13009056666"
		session.CRMOCode = 0
		So(session.CRMOCode, ShouldEqual, 0)
		SSODistributionByPhoneNumber(&session)
		So(session.CRMOCode, ShouldEqual, 2284)

		// 白城
		session.Phone = "13009066666"
		session.CRMOCode = 0
		So(session.CRMOCode, ShouldEqual, 0)
		SSODistributionByPhoneNumber(&session)
		So(session.CRMOCode, ShouldEqual, 2315)

		// 辽源
		session.Phone = "13009046666"
		session.CRMOCode = 0
		So(session.CRMOCode, ShouldEqual, 0)
		SSODistributionByPhoneNumber(&session)
		So(session.CRMOCode, ShouldEqual, 2268)

		// 外省
		session.Phone = "13384736666"
		session.CRMOCode = 0
		So(session.CRMOCode, ShouldEqual, 0)
		SSODistributionByPhoneNumber(&session)
		So(session.CRMOCode, ShouldNotEqual, 0)
	})
}

// Test_ssoRoundCrmList 测试 ssoRoundCrmList 函数是否可以循环分配手机号给九个地市分部
func Test_ssoRoundCrmList(t *testing.T) {
	Convey("测试 SSOSignUp 单点登模块注册接口的处理函数", t, func() {
		// 创建测试会话
		session := structs.SingleSignOnSession{}
		So(session.CRMOCode, ShouldEqual, 0)
		ssoRoundCrmList(&session)
		So(session.CRMOCode, ShouldNotEqual, 0)
	})
}
