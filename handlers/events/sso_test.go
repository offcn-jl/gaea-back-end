/*
   @Time : 2020/11/6 4:33 下午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : sso_test
   @Software: GoLand
*/

package events

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/jarcoal/httpmock"
	"github.com/jinzhu/gorm"
	"github.com/offcn-jl/gaea-back-end/commons"
	"github.com/offcn-jl/gaea-back-end/commons/config"
	"github.com/offcn-jl/gaea-back-end/commons/database/orm"
	"github.com/offcn-jl/gaea-back-end/commons/database/structs"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestSSOSendVerificationCode 测试 SSOSendVerificationCode 单点登模块发送验证码接口的处理函数
func TestSSOSendVerificationCode(t *testing.T) {
	// 初始化测试数据并获取测试所需的上下文
	unitTestTool, w, c := initTest()
	// 函数推出时重置数据库
	defer unitTestTool.CloseORM()
	Convey("测试 SSOSendVerificationCode 单点登模块发送验证码接口的处理函数", t, func() {
		// 测试 验证手机号码是否有效
		SSOSendVerificationCode(c)
		So(w.Body.String(), ShouldEqual, "{\"Msg\":\"手机号码不正确\"}")

		// 配置手机号码
		c.Params = gin.Params{gin.Param{Key: "Phone", Value: "17866886688"}}

		// 测试 获取登陆模块的配置
		w.Body.Reset() // 再次测试前重置 body
		SSOSendVerificationCode(c)
		So(w.Body.String(), ShouldEqual, "{\"Msg\":\"登陆模块配置有误\"}")

		// 配置登陆模块 ID
		c.Params = gin.Params{gin.Param{Key: "Phone", Value: "17866886688"}, gin.Param{Key: "MID", Value: "10001"}}
		w.Body.Reset() // 再次测试前重置 body
		SSOSendVerificationCode(c)
		So(w.Body.String(), ShouldEqual, "{\"Msg\":\"登陆模块 SMS 平台配置有误\"}")

		// 配置 SMS 平台为 中公短信平台
		orm.MySQL.Gaea.Model(structs.SingleSignOnLoginModule{}).Update(&structs.SingleSignOnLoginModule{Platform: 1})
		w.Body.Reset() // 再次测试前重置 body
		SSOSendVerificationCode(c)
		So(w.Body.String(), ShouldEqual, "{\"Msg\":\"短信模板配置有误\"}")

		// 配置 SMS 平台为 腾讯云
		orm.MySQL.Gaea.Model(structs.SingleSignOnLoginModule{}).Update(&structs.SingleSignOnLoginModule{Platform: 2})
		w.Body.Reset() // 再次测试前重置 body
		SSOSendVerificationCode(c)
		So(w.Body.String(), ShouldEqual, "{\"Msg\":\"未配置 腾讯云 令牌\"}")
	})
}

// TestSSOSignUp 测试 SSOSignUp 单点登模块注册接口的处理函数
func TestSSOSignUp(t *testing.T) {
	// 初始化测试数据并获取测试所需的上下文
	unitTestTool, w, c := initTest()
	// 函数推出时重置数据库
	defer unitTestTool.CloseORM()

	Convey("测试 SSOSignUp 单点登模块注册接口的处理函数", t, func() {
		// 测试 未绑定 Body 数据
		SSOSignUp(c)
		So(w.Body.String(), ShouldEqual, "{\"Msg\":\"提交的 Json 数据不正确\"}")

		// 增加 Body
		c.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"MID\":10000,\"Phone\":\"*\"}"))

		// 测试 验证手机号码是否有效
		w.Body.Reset() // 再次测试前重置 body
		SSOSignUp(c)
		So(w.Body.String(), ShouldEqual, "{\"Msg\":\"手机号码不正确\"}")

		// 修正手机号码
		c.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"MID\":10000,\"Phone\":\"17866668888\"}"))

		// 测试 校验登录模块配置
		w.Body.Reset() // 再次测试前重置 body
		SSOSignUp(c)
		So(w.Body.String(), ShouldEqual, "{\"Msg\":\"单点登陆模块配置有误\"}")

		// 修正为已存在的登陆模块 ID
		c.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"MID\":10001,\"Phone\":\"17866668888\"}"))

		// 测试 校验是否发送过验证码
		w.Body.Reset() // 再次测试前重置 body
		SSOSignUp(c)
		So(w.Body.String(), ShouldEqual, "{\"Msg\":\"请您先获取验证码后再进行注册\"}")

		// 模拟获取验证码
		createTestVerificationCode()

		// 修正为模拟获取过验证码的手机号码
		c.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"MID\":10001,\"Phone\":\"17888886666\"}"))

		// 测试 校验验证码是正确 ( 此时的上下文中未填写验证码 )
		w.Body.Reset() // 再次测试前重置 body
		SSOSignUp(c)
		So(w.Body.String(), ShouldEqual, "{\"Msg\":\"验证码有误\"}")

		// 向请求中添加正确, 但已经失效的验证码
		c.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"MID\":10001,\"Phone\":\"17888886666\",\"Code\":9999}"))

		// 测试 校验验证码是否有效
		w.Body.Reset() // 再次测试前重置 body
		SSOSignUp(c)
		So(w.Body.String(), ShouldEqual, "{\"Msg\":\"验证码失效\"}")

		// 更换上下文中的手机号码及验证码未为正确且未失效的内容
		c.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"MID\":10001,\"Phone\":\"17866886688\",\"Code\":9999}"))

		// 测试 注册成功
		w.Body.Reset() // 再次测试前重置 body
		SSOSignUp(c)
		So(w.Body.String(), ShouldEqual, "{\"Msg\":\"Success\"}")

		// 判断用户是否存在
		So(ssoIsSignUp("17866886688"), ShouldBeTrue)

		// 判断会话是否存在
		So(ssoIsSignIn("17866886688", 10001), ShouldBeTrue)
	})
}

// TestSSOSignIn 测试 SSOSignIn 单点登陆模块登陆接口的处理函数
func TestSSOSignIn(t *testing.T) {
	// 初始化测试数据并获取测试所需的上下文
	unitTestTool, w, c := initTest()
	// 函数推出时重置数据库
	defer unitTestTool.CloseORM()

	Convey("测试 SSOSignUp 单点登陆模块登陆接口的处理函数", t, func() {
		// 测试 未绑定 Body 数据
		SSOSignIn(c)
		So(w.Body.String(), ShouldEqual, "{\"Msg\":\"提交的 Json 数据不正确\"}")

		// 增加 Body
		c.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"MID\":10000,\"Phone\":\"*\"}"))

		// 测试 验证手机号码是否有效
		w.Body.Reset() // 再次测试前重置 body
		SSOSignIn(c)
		So(w.Body.String(), ShouldEqual, "{\"Msg\":\"手机号码不正确\"}")

		// 修正手机号码
		c.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"MID\":10000,\"Phone\":\"17866668888\"}"))

		// 测试 校验登录模块配置 ( 此时的上下文中没有登陆模块的配置 )
		w.Body.Reset() // 再次测试前重置 body
		SSOSignIn(c)
		So(w.Body.String(), ShouldEqual, "{\"Msg\":\"单点登陆模块配置有误\"}")

		// 修正为已存在的登陆模块 ID
		c.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"MID\":10001,\"Phone\":\"17866668888\"}"))

		// 测试 校验用户是否已经注册
		w.Body.Reset() // 再次测试前重置 body
		SSOSignIn(c)
		So(w.Body.String(), ShouldEqual, "{\"Msg\":\"请您先进行注册\"}")

		// 修正为已经注册的手机号码
		c.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"MID\":10001,\"Phone\":\"17888666688\"}"))

		// 测试 登陆成功
		w.Body.Reset() // 再次测试前重置 body
		SSOSignIn(c)
		So(w.Body.String(), ShouldEqual, "{\"Msg\":\"Success\"}")

		// 判断会话是否存在
		So(ssoIsSignIn("17888666688", 10001), ShouldBeTrue)
	})

}

// TestSSOGetSessionInfo 测试 SSOGetSessionInfo 函数是否可以获取会话信息
func TestSSOGetSessionInfo(t *testing.T) {
	// 初始化测试数据并获取测试所需的上下文
	unitTestTool, w, c := initTest()
	// 函数推出时重置数据库
	defer unitTestTool.CloseORM()

	Convey("测试 GetSessionInfo 函数是否可以获取会话信息", t, func() {
		Convey("测试 验证手机号是否有效 ( 需要是 0 或正确的手机号 )", func() {
			// 此时没有填写手机号
			SSOSessionInfo(c)
			So(w.Body.String(), ShouldContainSubstring, "手机号码不正确")
			// 此时填写了错误的手机号
			c.Params = gin.Params{gin.Param{Key: "Phone", Value: "1788710666"}}
			w.Body.Reset() // 再次测试前重置 body
			SSOSessionInfo(c)
			So(w.Body.String(), ShouldContainSubstring, "手机号码不正确")
		})

		// 修正为正确的手机号码, 供后续测试使用
		c.Params = gin.Params{gin.Param{Key: "Phone", Value: "17887106666"}}

		Convey("测试 校验登陆模块配置", func() {
			// 此时没有填写登陆模块配置
			w.Body.Reset() // 再次测试前重置 body
			SSOSessionInfo(c)
			So(w.Body.String(), ShouldContainSubstring, "单点登陆模块配置有误")
			// 此时填写了错误的登陆模块 ID
			c.Params = gin.Params{gin.Param{Key: "Phone", Value: "17887106666"}, gin.Param{Key: "MID", Value: "20001"}}
			w.Body.Reset() // 再次测试前重置 body
			SSOSessionInfo(c)
			So(w.Body.String(), ShouldContainSubstring, "单点登陆模块配置有误")
		})

		// 修正为正确的登陆模块 ID
		c.Params = gin.Params{gin.Param{Key: "Phone", Value: "17887106666"}, gin.Param{Key: "MID", Value: "10001"}}

		Convey("测试 后缀不存在或错误", func() {
			defaultInfo := "{\"Data\":{\"Sign\":\"中公教育\",\"CRMEID\":\"HD202010142576\",\"CRMEFID\":\"56975\",\"CRMEFSID\":\"f905e07b2bff94d564ac1fa41022a633\",\"CRMChannel\":7,\"CRMOCode\":22,\"CRMOName\":\"吉林分校\",\"CRMUID\":32431,\"CRMUser\":\"default\",\"Suffix\":\"default\",\"NTalkerGID\":\"NTalkerGID\",\"IsLogin\":false,\"NeedToRegister\":true},\"Msg\":\"Success\"}"
			// 此时未配置后缀, 即后缀不存在, 可以认为等同为后缀错误
			w.Body.Reset() // 再次测试前重置 body
			SSOSessionInfo(c)
			So(w.Body.String(), ShouldEqual, defaultInfo)
			// 设置设置默认后缀 CRM 组织 ID 为 0
			orm.MySQL.Gaea.Model(structs.SingleSignOnSuffix{}).Where("suffix = 'default'").Update("crm_oid", "0")
			w.Body.Reset() // 再次测试前重置 body
			SSOSessionInfo(c)
			So(w.Body.String(), ShouldEqual, defaultInfo)
			// 设置默认后缀 CRM 组织 ID 为不存在的 ID
			orm.MySQL.Gaea.Model(structs.SingleSignOnSuffix{}).Where("suffix = 'default'").Update("crm_oid", "10000")
			w.Body.Reset() // 再次测试前重置 body
			SSOSessionInfo(c)
			So(w.Body.String(), ShouldEqual, defaultInfo)
			// 还原默认后缀 CRM 组织 ID
			orm.MySQL.Gaea.Model(structs.SingleSignOnSuffix{}).Where("suffix = 'default'").Update("crm_oid", "1")
		})

		Convey("测试 配置了后缀 是否可以获取到后缀对应的信息", func() {
			testInfo := "{\"Data\":{\"Sign\":\"中公教育\",\"CRMEID\":\"HD202010142576\",\"CRMEFID\":\"56975\",\"CRMEFSID\":\"f905e07b2bff94d564ac1fa41022a633\",\"CRMChannel\":22,\"CRMOCode\":2290,\"CRMOName\":\"吉林长春分校\",\"CRMUID\":123,\"CRMUser\":\"test\",\"Suffix\":\"test\",\"NTalkerGID\":\"NTalkerGID\",\"IsLogin\":false,\"NeedToRegister\":true},\"Msg\":\"Success\"}"
			testInfoWithDefauleOrgnation := "{\"Data\":{\"Sign\":\"中公教育\",\"CRMEID\":\"HD202010142576\",\"CRMEFID\":\"56975\",\"CRMEFSID\":\"f905e07b2bff94d564ac1fa41022a633\",\"CRMChannel\":22,\"CRMOCode\":22,\"CRMOName\":\"吉林分校\",\"CRMUID\":123,\"CRMUser\":\"test\",\"Suffix\":\"test\",\"NTalkerGID\":\"NTalkerGID\",\"IsLogin\":false,\"NeedToRegister\":true},\"Msg\":\"Success\"}"
			// 配置后缀为测试后缀
			c.Params = gin.Params{gin.Param{Key: "Phone", Value: "17887106666"}, gin.Param{Key: "MID", Value: "10001"}, gin.Param{Key: "Suffix", Value: "test"}}
			w.Body.Reset() // 再次测试前重置 body
			SSOSessionInfo(c)
			So(w.Body.String(), ShouldEqual, testInfo)
			// 设置测试后缀 CRM 组织 ID 为 0
			orm.MySQL.Gaea.Unscoped().Model(structs.SingleSignOnSuffix{}).Where("suffix = 'test'").Update("crm_oid", "0")
			w.Body.Reset() // 再次测试前重置 body
			SSOSessionInfo(c)
			So(w.Body.String(), ShouldEqual, testInfoWithDefauleOrgnation)
			// 设置测试后缀 CRM 组织 ID 为不存在的 ID
			orm.MySQL.Gaea.Unscoped().Model(structs.SingleSignOnSuffix{}).Where("suffix = 'test'").Update("crm_oid", "10000")
			w.Body.Reset() // 再次测试前重置 body
			SSOSessionInfo(c)
			So(w.Body.String(), ShouldEqual, testInfoWithDefauleOrgnation)
			// 还原测试后缀 CRM 组织 ID
			orm.MySQL.Gaea.Unscoped().Model(structs.SingleSignOnSuffix{}).Where("suffix = 'test'").Update("crm_oid", "2")
		})

		Convey("测试 配置了已经过期的后缀 是否依旧可以获取到后缀对应的信息", func() {
			testInfo := "{\"Data\":{\"Sign\":\"中公教育\",\"CRMEID\":\"HD202010142576\",\"CRMEFID\":\"56975\",\"CRMEFSID\":\"f905e07b2bff94d564ac1fa41022a633\",\"CRMChannel\":22,\"CRMOCode\":2290,\"CRMOName\":\"吉林长春分校\",\"CRMUID\":123,\"CRMUser\":\"expired\",\"Suffix\":\"expired\",\"NTalkerGID\":\"NTalkerGID\",\"IsLogin\":false,\"NeedToRegister\":true},\"Msg\":\"Success\"}"
			testInfoWithDefauleOrgnation := "{\"Data\":{\"Sign\":\"中公教育\",\"CRMEID\":\"HD202010142576\",\"CRMEFID\":\"56975\",\"CRMEFSID\":\"f905e07b2bff94d564ac1fa41022a633\",\"CRMChannel\":22,\"CRMOCode\":22,\"CRMOName\":\"吉林分校\",\"CRMUID\":123,\"CRMUser\":\"expired\",\"Suffix\":\"expired\",\"NTalkerGID\":\"NTalkerGID\",\"IsLogin\":false,\"NeedToRegister\":true},\"Msg\":\"Success\"}"
			// 配置后缀为过期后缀
			c.Params = gin.Params{gin.Param{Key: "Phone", Value: "17887106666"}, gin.Param{Key: "MID", Value: "10001"}, gin.Param{Key: "Suffix", Value: "expired"}}
			w.Body.Reset() // 再次测试前重置 body
			SSOSessionInfo(c)
			So(w.Body.String(), ShouldEqual, testInfo)
			// 设置过期后缀 CRM 组织 ID 为 0
			orm.MySQL.Gaea.Unscoped().Model(structs.SingleSignOnSuffix{}).Where("suffix = 'expired'").Update("crm_oid", "0")
			w.Body.Reset() // 再次测试前重置 body
			SSOSessionInfo(c)
			So(w.Body.String(), ShouldEqual, testInfoWithDefauleOrgnation)
			// 设置过期后缀 CRM 组织 ID 为不存在的 ID
			orm.MySQL.Gaea.Unscoped().Model(structs.SingleSignOnSuffix{}).Where("suffix = 'expired'").Update("crm_oid", "10000")
			w.Body.Reset() // 再次测试前重置 body
			SSOSessionInfo(c)
			So(w.Body.String(), ShouldEqual, testInfoWithDefauleOrgnation)
			// 还原过期后缀 CRM 组织 ID
			orm.MySQL.Gaea.Unscoped().Model(structs.SingleSignOnSuffix{}).Where("suffix = 'expired'").Update("crm_oid", "2")
		})

		Convey("测试 校验是否需要注册", func() {
			// 模拟注册手机号
			userInfo := structs.SingleSignOnUser{}
			userInfo.Phone = "1788710" + time.Now().Format("0405") // 使用当前时间生成手机号尾号用于测试
			orm.MySQL.Gaea.Create(&userInfo)
			// 更换手机号为已经注册的手机号
			c.Params = gin.Params{gin.Param{Key: "Phone", Value: userInfo.Phone}, gin.Param{Key: "MID", Value: "10001"}, gin.Param{Key: "Suffix", Value: "test"}}
			w.Body.Reset() // 再次测试前重置 body
			SSOSessionInfo(c)
			So(w.Body.String(), ShouldContainSubstring, "\"NeedToRegister\":false")
		})

		Convey("测试 校验是否需要登陆", func() {
			// 模拟注册手机号
			userInfo := structs.SingleSignOnUser{
				Phone: "1788710" + time.Now().Format("0405"), // 使用当前时间生成手机号尾号用于测试
			}
			orm.MySQL.Gaea.Create(&userInfo)
			sessionInfo := structs.SingleSignOnSession{
				MID:   10001,
				Phone: userInfo.Phone,
			}
			orm.MySQL.Gaea.Create(&sessionInfo)
			// 更换手机号为已经注册的手机号
			c.Params = gin.Params{gin.Param{Key: "Phone", Value: userInfo.Phone}, gin.Param{Key: "MID", Value: "10001"}, gin.Param{Key: "Suffix", Value: "test"}}
			w.Body.Reset() // 再次测试前重置 body
			SSOSessionInfo(c)
			So(w.Body.String(), ShouldContainSubstring, "\"IsLogin\":true")
		})
	})
}

// Test_sendVerificationCodeByOFFCN 测试 sendVerificationCodeByOFFCN 是否可以使用中公短信平台发送验证码
func Test_sendVerificationCodeByOFFCN(t *testing.T) {
	// 初始化测试数据并获取测试所需的上下文
	unitTestTool, w, c := initTest()
	// 函数推出时重置数据库
	defer unitTestTool.CloseORM()

	Convey("测试 sendVerificationCodeByOFFCN 是否可以使用中公短信平台发送验证码", t, func() {
		// 测试验证配置
		// 此时没有任何配置
		sendVerificationCodeByOFFCN(c, 0, 0)
		So(w.Body.String(), ShouldEqual, "{\"Msg\":\"短信模板配置有误\"}")
		// 配置短信模板
		w.Body.Reset() // 再次测试前重置 body
		sendVerificationCodeByOFFCN(c, 392074, 0)
		So(w.Body.String(), ShouldEqual, "{\"Msg\":\"未配置 中公教育短信平台 接口地址\"}")
		// 配置 中公教育短信平台 接口地址
		config.Update(orm.MySQL.Gaea, structs.SystemConfig{OffcnSmsURL: "https://fake-sms-platform.offcn.com"})
		w.Body.Reset() // 再次测试前重置 body
		sendVerificationCodeByOFFCN(c, 392074, 0)
		So(w.Body.String(), ShouldEqual, "{\"Msg\":\"未配置 中公教育短信平台 用户名\"}")
		// 配置 中公教育短信平台 用户名
		config.Update(orm.MySQL.Gaea, structs.SystemConfig{OffcnSmsURL: "https://fake-sms-platform.offcn.com", OffcnSmsUserName: "fake-sms-user"})
		w.Body.Reset() // 再次测试前重置 body
		sendVerificationCodeByOFFCN(c, 392074, 0)
		So(w.Body.String(), ShouldEqual, "{\"Msg\":\"未配置 中公教育短信平台 密码\"}")
		// 配置 中公教育短信平台 密码
		config.Update(orm.MySQL.Gaea, structs.SystemConfig{OffcnSmsURL: "https://fake-sms-platform.offcn.com", OffcnSmsUserName: "fake-sms-user", OffcnSmsPassword: "fake-sms-password"})
		w.Body.Reset() // 再次测试前重置 body
		sendVerificationCodeByOFFCN(c, 392074, 0)
		So(w.Body.String(), ShouldEqual, "{\"Msg\":\"未配置 中公教育短信平台 发送方识别码\"}")
		// 配置 中公教育短信平台 发送方识别码
		config.Update(orm.MySQL.Gaea, structs.SystemConfig{OffcnSmsURL: "https://fake-sms-platform.offcn.com", OffcnSmsUserName: "fake-sms-user", OffcnSmsPassword: "fake-sms-password", OffcnSmsTjCode: "fake-sms-tj-code"})

		// 添加发送记录
		createTestVerificationCode()
		// 配置手机号码
		c.Params = gin.Params{gin.Param{Key: "Phone", Value: "17866886688"}}

		w.Body.Reset() // 再次测试前重置 body
		sendVerificationCodeByOFFCN(c, 392074, 0)
		So(w.Body.String(), ShouldEqual, "{\"Msg\":\"请勿重复发送验证码\"}")

		// 配置为未发送过短信的手机号码
		c.Params = gin.Params{gin.Param{Key: "Phone", Value: "17888668866"}}

		// 配置 httpmock 进行拦截
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		w.Body.Reset() // 再次测试前重置 body
		sendVerificationCodeByOFFCN(c, 392074, 0)
		So(w.Body.String(), ShouldEqual, "{\"Msg\":\"发送短信失败, 返回错误 : Post \\\"https://fake-sms-platform.offcn.com\\\": no responder found\"}")

		// 配置 httpmock 返回 非 200 的 http 状态码
		httpmock.RegisterResponder(http.MethodPost, "https://fake-sms-platform.offcn.com", httpmock.NewStringResponder(http.StatusInternalServerError, ""))
		w.Body.Reset() // 再次测试前重置 body
		sendVerificationCodeByOFFCN(c, 392074, 0)
		So(w.Body.String(), ShouldEqual, "{\"Msg\":\"发送短信失败, 返回状态码 : 500\"}")

		// 配置 httpmock 返回 无法读取的 body todo 暂时无法测试

		// 配置 httpmock 返回 错误的 Json
		httpmock.RegisterResponder(http.MethodPost, "https://fake-sms-platform.offcn.com", httpmock.NewBytesResponder(http.StatusOK, []byte("")))
		w.Body.Reset() // 再次测试前重置 body
		sendVerificationCodeByOFFCN(c, 392074, 0)
		So(w.Body.String(), ShouldEqual, "{\"Msg\":\"发送短信失败, 解码返回内容失败, 错误内容 : unexpected end of JSON input\"}")

		// 配置 httpmock 返回 发送失败的 Json
		httpmock.RegisterResponder(http.MethodPost, "https://fake-sms-platform.offcn.com", httpmock.NewJsonResponderOrPanic(http.StatusOK, gin.H{"status": 0, "msg": "fail"}))
		w.Body.Reset() // 再次测试前重置 body
		sendVerificationCodeByOFFCN(c, 392074, 0)
		So(w.Body.String(), ShouldEqual, "{\"Msg\":\"发送短信失败, [ 0 ] fail\"}")

		// 配置 httpmock 返回 发送成功 Json
		httpmock.RegisterResponder(http.MethodPost, "https://fake-sms-platform.offcn.com", httpmock.NewJsonResponderOrPanic(http.StatusOK, gin.H{"status": 1}))
		w.Body.Reset()                                            // 再次测试前重置 body
		c.Request, _ = http.NewRequest(http.MethodPost, "/", nil) // 初始化 Request 避免 c.ClientIP() 出现空指针错误
		sendVerificationCodeByOFFCN(c, 392074, 10)
		So(w.Body.String(), ShouldEqual, "{\"Msg\":\"Success\"}")

		// 测试其他短信模板
		w.Body.Reset() // 再次测试前重置 body
		sendVerificationCodeByOFFCN(c, 392030, 10)
		So(w.Body.String(), ShouldEqual, "{\"Msg\":\"请勿重复发送验证码\"}")
		w.Body.Reset() // 再次测试前重置 body
		sendVerificationCodeByOFFCN(c, 391863, 10)
		So(w.Body.String(), ShouldEqual, "{\"Msg\":\"请勿重复发送验证码\"}")

		// 检查发送记录
		singleSignOnVerificationCode := structs.SingleSignOnVerificationCode{}
		orm.MySQL.Gaea.Find(&singleSignOnVerificationCode)
		So(singleSignOnVerificationCode.Phone, ShouldEqual, "17888668866")
		So(singleSignOnVerificationCode.Term, ShouldEqual, 10)
		So(singleSignOnVerificationCode.SourceIP, ShouldEqual, "")
	})
}

// Test_sendVerificationCodeByTencentCloudSMSV2 测试 sendVerificationCodeByTencentCloudSMSV2 使用腾讯云短信平台发送验证码
func Test_sendVerificationCodeByTencentCloudSMSV2(t *testing.T) {
	// 初始化测试数据并获取测试所需的上下文
	unitTestTool, w, c := initTest()
	// 函数推出时重置数据库
	defer unitTestTool.CloseORM()

	Convey("测试 sendVerificationCodeByTencentCloudSMSV2 使用腾讯云短信平台发送验证码", t, func() {
		// 测试验证配置
		// 此时没有任何配置
		sendVerificationCodeByTencentCloudSMSV2(c, "", 0, 0)
		So(w.Body.String(), ShouldEqual, "{\"Msg\":\"未配置 腾讯云 令牌\"}")

		// 配置 腾讯云 令牌
		config.Update(orm.MySQL.Gaea, structs.SystemConfig{TencentCloudAPISecretID: "fake-tencent-cloud-api-secret-id"})
		w.Body.Reset() // 再次测试前重置 body
		sendVerificationCodeByTencentCloudSMSV2(c, "", 0, 0)
		So(w.Body.String(), ShouldEqual, "{\"Msg\":\"未配置 腾讯云 密钥\"}")
		// 配置 腾讯云 密钥
		config.Update(orm.MySQL.Gaea, structs.SystemConfig{TencentCloudAPISecretID: "fake-tencent-cloud-api-secret-id", TencentCloudAPISecretKey: "fake-tencent-cloud-api-secret-key"})
		w.Body.Reset() // 再次测试前重置 body
		sendVerificationCodeByTencentCloudSMSV2(c, "", 0, 0)
		So(w.Body.String(), ShouldEqual, "{\"Msg\":\"未配置 腾讯云 短信应用 ID\"}")
		// 配置 腾讯云 短信应用 ID
		config.Update(orm.MySQL.Gaea, structs.SystemConfig{TencentCloudAPISecretID: "fake-tencent-cloud-api-secret-id", TencentCloudAPISecretKey: "fake-tencent-cloud-api-secret-key", TencentCloudSmsSdkAppId: "fake-tencent-cloud-sms-sdk-app-id"})

		// 添加发送记录
		createTestVerificationCode()
		// 配置手机号码
		c.Params = gin.Params{gin.Param{Key: "Phone", Value: "17866886688"}}

		// 测试重复发送验证码
		w.Body.Reset() // 再次测试前重置 body
		sendVerificationCodeByTencentCloudSMSV2(c, "", 0, 0)
		So(w.Body.String(), ShouldEqual, "{\"Msg\":\"请勿重复发送验证码\"}")

		// 配置为未发送过短信的手机号码
		c.Params = gin.Params{gin.Param{Key: "Phone", Value: "17888668866"}}

		// 配置 httpmock 进行拦截
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		// 测试发送短信失败
		w.Body.Reset() // 再次测试前重置 body
		sendVerificationCodeByTencentCloudSMSV2(c, "", 0, 0)
		So(w.Body.String(), ShouldEqual, "{\"Msg\":\"发送短信失败, [ ClientError.NetworkError ] Fail to get response because Post \\\"https://sms.internal.tencentcloudapi.com/\\\": no responder found\"}")

		// 测试 非 SDK 异常 todo 暂时无法测试

		// 配置 httpmock 返回 发送失败的 Json
		// https://cloud.tencent.com/document/product/382/38770
		httpmock.RegisterNoResponder(httpmock.NewJsonResponderOrPanic(http.StatusOK, gin.H{"Response": gin.H{"Error": gin.H{"Code": "AuthFailure.SignatureFailure", "Message": "The provided credentials could not be validated. Please check your signature is correct."}, "RequestId": "ed93f3cb-f35e-473f-b9f3-0d451b8b79c6"}}))
		w.Body.Reset() // 再次测试前重置 body
		sendVerificationCodeByTencentCloudSMSV2(c, "", 0, 0)
		So(w.Body.String(), ShouldEqual, "{\"Msg\":\"发送短信失败, [ AuthFailure.SignatureFailure ] The provided credentials could not be validated. Please check your signature is correct.\"}")

		// 配置 httpmock 返回 发送失败的 Json
		// https://cloud.tencent.com/document/product/382/38770
		httpmock.RegisterNoResponder(httpmock.NewJsonResponderOrPanic(http.StatusOK, gin.H{"Response": gin.H{"SendStatusSet": []gin.H{{"Code": "FailCode", "Message": "FailMessage"}}}}))
		w.Body.Reset() // 再次测试前重置 body
		sendVerificationCodeByTencentCloudSMSV2(c, "", 0, 10)
		So(w.Body.String(), ShouldEqual, "{\"Msg\":\"发送短信失败, 错误内容 : FailMessage\"}")

		// 配置 httpmock 返回 发送成功的 Json
		// https://cloud.tencent.com/document/product/382/38770
		httpmock.RegisterNoResponder(httpmock.NewJsonResponderOrPanic(http.StatusOK, gin.H{"Response": gin.H{"SendStatusSet": []gin.H{{"Code": "Ok"}}}}))
		w.Body.Reset()                                            // 再次测试前重置 body
		c.Request, _ = http.NewRequest(http.MethodPost, "/", nil) // 初始化 Request 避免 c.ClientIP() 出现空指针错误
		sendVerificationCodeByTencentCloudSMSV2(c, "", 0, 10)
		So(w.Body.String(), ShouldEqual, "{\"Msg\":\"Success\"}")

		// 检查发送记录
		singleSignOnVerificationCode := structs.SingleSignOnVerificationCode{}
		orm.MySQL.Gaea.Find(&singleSignOnVerificationCode)
		So(singleSignOnVerificationCode.Phone, ShouldEqual, "17888668866")
		So(singleSignOnVerificationCode.Term, ShouldEqual, 10)
		So(singleSignOnVerificationCode.SourceIP, ShouldEqual, "")
	})
}

// Test_ssoIsSignUp 测试 ssoIsSignUp 函数是否可以检查用户是否已经注册且未失效
func Test_ssoIsSignUp(t *testing.T) {
	// 初始化测试数据并获取测试所需的上下文
	unitTestTool, _, _ := initTest()
	// 函数推出时重置数据库
	defer unitTestTool.CloseORM()

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
	// 初始化测试数据并获取测试所需的上下文
	unitTestTool, _, _ := initTest()
	// 函数推出时重置数据库
	defer unitTestTool.CloseORM()

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
	// 初始化测试数据并获取测试所需的上下文
	unitTestTool, _, _ := initTest()
	// 函数推出时重置数据库
	defer unitTestTool.CloseORM()

	Convey("测试 ssoCreateSession 函数是否可以按照预期矫正信息后创建会话", t, func() {
		// 创建测试会话
		session := structs.SingleSignOnSession{}
		session.Phone = "1788710" + time.Now().Format("0405") // 使用当前时间生成手机号尾号用于测试
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
		session.Phone = "1868648" + time.Now().Format("0405") // 使用当前时间生成手机号尾号用于测试
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
		session.Phone = "1868648" + time.Now().Format("0405") // 使用当前时间生成手机号尾号用于测试
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
		session.Phone = "1868648" + time.Now().Format("0405") // 使用当前时间生成手机号尾号用于测试
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
		session.Phone = "1868648" + time.Now().Format("0405") // 使用当前时间生成手机号尾号用于测试
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
		//httpmock.RegisterResponder(http.MethodGet, "https://dc.offcn.com:8443/a.gif", httpmock.ConnectionFailure)
		session.ID = 0 // 将 ID 恢复为 0, 令 ORM 认为这条 Session 是新记录
		session.ActualSuffix = "default"
		session.Phone = "1868648" + time.Now().Format("0405") // 使用当前时间生成手机号尾号用于测试
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
		session.Phone = "1868648" + time.Now().Format("0405") // 使用当前时间生成手机号尾号用于测试
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

// Test_ssoGetDefaultSuffix 测试 ssoGetDefaultSuffix 函数是否可以获取默认后缀配置
func Test_ssoGetDefaultSuffix(t *testing.T) {
	// 初始化测试数据
	unitTestTool, _, _ := initTest()
	// 函数推出时重置数据库
	defer unitTestTool.CloseORM()

	Convey("测试 ssoGetDefaultSuffix 函数是否可以获取默认后缀配置", t, func() {
		// 创建测试会话
		session := structs.SingleSignOnSession{}
		session.Phone = "17887106666"
		So(session.CRMChannel, ShouldEqual, 0)
		So(session.CRMOCode, ShouldEqual, 0)
		ssoGetDefaultSuffix(&session)
		So(session.CRMChannel, ShouldEqual, 7)
		So(session.CRMOCode, ShouldEqual, 2290)
		So(session.CurrentSuffix, ShouldEqual, "default")
		So(session.CRMUID, ShouldEqual, 32431)
	})
}

// Test_ssoDistributionByPhoneNumber 测试 ssoDistributionByPhoneNumber 函数是否可以按照手机号码归属地进行归属分部分配
// 号段数据来自 http://www.bixinshui.com
func Test_ssoDistributionByPhoneNumber(t *testing.T) {
	// 初始化测试数据
	unitTestTool, _, _ := initTest()
	// 函数推出时重置数据库
	defer unitTestTool.CloseORM()

	Convey("测试 ssoDistributionByPhoneNumber 函数是否可以按照手机号码归属地进行归属分部分配", t, func() {
		// 创建测试会话
		session := structs.SingleSignOnSession{}

		// 解析出错
		session.Phone = "9999"
		session.CRMOCode = 0
		So(session.CRMOCode, ShouldEqual, 0)
		ssoDistributionByPhoneNumber(&session)
		So(session.CRMOCode, ShouldNotEqual, 0)

		// 长春
		session.Phone = "17887106666"
		session.CRMOCode = 0
		So(session.CRMOCode, ShouldEqual, 0)
		ssoDistributionByPhoneNumber(&session)
		So(session.CRMOCode, ShouldEqual, 2290)

		// 吉林
		session.Phone = "13009156666"
		session.CRMOCode = 0
		So(session.CRMOCode, ShouldEqual, 0)
		ssoDistributionByPhoneNumber(&session)
		So(session.CRMOCode, ShouldEqual, 2305)

		// 延边
		session.Phone = "18943306666"
		session.CRMOCode = 0
		So(session.CRMOCode, ShouldEqual, 0)
		ssoDistributionByPhoneNumber(&session)
		So(session.CRMOCode, ShouldEqual, 2277)

		// 通化
		session.Phone = "13009196666"
		session.CRMOCode = 0
		So(session.CRMOCode, ShouldEqual, 0)
		ssoDistributionByPhoneNumber(&session)
		So(session.CRMOCode, ShouldEqual, 2271)

		// 白山
		session.Phone = "13009076666"
		session.CRMOCode = 0
		So(session.CRMOCode, ShouldEqual, 0)
		ssoDistributionByPhoneNumber(&session)
		So(session.CRMOCode, ShouldEqual, 2310)

		// 四平
		session.Phone = "13009026666"
		session.CRMOCode = 0
		So(session.CRMOCode, ShouldEqual, 0)
		ssoDistributionByPhoneNumber(&session)
		So(session.CRMOCode, ShouldEqual, 2263)

		// 松原
		session.Phone = "13009056666"
		session.CRMOCode = 0
		So(session.CRMOCode, ShouldEqual, 0)
		ssoDistributionByPhoneNumber(&session)
		So(session.CRMOCode, ShouldEqual, 2284)

		// 白城
		session.Phone = "13009066666"
		session.CRMOCode = 0
		So(session.CRMOCode, ShouldEqual, 0)
		ssoDistributionByPhoneNumber(&session)
		So(session.CRMOCode, ShouldEqual, 2315)

		// 辽源
		session.Phone = "13009046666"
		session.CRMOCode = 0
		So(session.CRMOCode, ShouldEqual, 0)
		ssoDistributionByPhoneNumber(&session)
		So(session.CRMOCode, ShouldEqual, 2268)

		// 外省
		session.Phone = "13384736666"
		session.CRMOCode = 0
		So(session.CRMOCode, ShouldEqual, 0)
		ssoDistributionByPhoneNumber(&session)
		So(session.CRMOCode, ShouldNotEqual, 0)
	})
}

// Test_ssoRoundCrmList 测试 ssoRoundCrmList 函数是否可以循环分配手机号给九个地市分部
func Test_ssoRoundCrmList(t *testing.T) {
	// 初始化测试数据
	unitTestTool, _, _ := initTest()
	// 函数推出时重置数据库
	defer unitTestTool.CloseORM()

	Convey("测试 SSOSignUp 单点登模块注册接口的处理函数", t, func() {
		// 创建测试会话
		session := structs.SingleSignOnSession{}
		So(session.CRMOCode, ShouldEqual, 0)
		ssoRoundCrmList(&session)
		So(session.CRMOCode, ShouldNotEqual, 0)
	})
}

// initTest 初始化测试数据并返回测试所需的上下文
func initTest() (commons.UnitTestTool, *httptest.ResponseRecorder, *gin.Context) {
	// 初始化数据库
	unitTestTool := commons.UnitTestTool{}
	unitTestTool.CreatORM()
	unitTestTool.InitORM()
	orm.MySQL.Gaea = unitTestTool.ORM

	// 创建 测试用组织信息
	// 省级分校
	orm.MySQL.Gaea.Create(&structs.SingleSignOnOrganization{Model: gorm.Model{ID: 1}, Code: 22, Name: "吉林分校"})
	// 地市分校 1
	orm.MySQL.Gaea.Create(&structs.SingleSignOnOrganization{Model: gorm.Model{ID: 2}, FID: 1, Code: 2290, Name: "吉林长春分校"})
	// 地市分校 2
	orm.MySQL.Gaea.Create(&structs.SingleSignOnOrganization{Model: gorm.Model{ID: 3}, FID: 1, Code: 2305, Name: "吉林市分校"})

	// 创建 测试用后缀信息
	// 默认后缀 ( ID = 1 )
	orm.MySQL.Gaea.Create(&structs.SingleSignOnSuffix{Model: gorm.Model{ID: 1}, Suffix: "default", CRMUser: "default", CRMUID: 32431 /* 齐* */, CRMOID: 1 /* 吉林分校 */, CRMChannel: 7 /* 19 课堂 ( 网推 ) */, NTalkerGID: "NTalkerGID"})
	// 已删除, 但是依旧有效 ( 未到达配置的删除时间 ) 的后缀
	tmpTime := time.Now().Add(8760 * time.Hour) // 一年后
	orm.MySQL.Gaea.Create(&structs.SingleSignOnSuffix{Model: gorm.Model{ID: 2, DeletedAt: &tmpTime}, Suffix: "test", CRMUser: "test", CRMUID: 123 /* 高** */, CRMOID: 2 /* 吉林长春分校 */, CRMChannel: 22 /* 户外推广 ( 市场 ) */, NTalkerGID: "NTalkerGID"})
	// 已删除, 并且已经失效 ( 到达删除时间 ) 的后缀
	tmpTime = time.Now().Add(-8760 * time.Hour) // 一年前
	orm.MySQL.Gaea.Create(&structs.SingleSignOnSuffix{Model: gorm.Model{ID: 3, DeletedAt: &tmpTime}, Suffix: "expired", CRMUser: "expired", CRMUID: 123 /* 高** */, CRMOID: 2 /* 吉林长春分校 */, CRMChannel: 22 /* 户外推广 ( 市场 ) */, NTalkerGID: "NTalkerGID"})

	// 创建 测试用登陆模块信息
	orm.MySQL.Gaea.Create(&structs.SingleSignOnLoginModule{Model: gorm.Model{ID: 10001}, CRMEID: "HD202010142576", CRMEFID: "56975", CRMEFSID: "f905e07b2bff94d564ac1fa41022a633", Sign: "中公教育"})

	// 创建 测试用用户
	orm.MySQL.Gaea.Create(&structs.SingleSignOnUser{Phone: "17888666688"})

	// 创建测试使用的上下文
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// 返回上下文
	return unitTestTool, w, c
}

// createTestVerificationCode 模拟获取验证码
func createTestVerificationCode() {
	// 验证码正确, 但是已经失效
	orm.MySQL.Gaea.Create(&structs.SingleSignOnVerificationCode{
		Model: gorm.Model{CreatedAt: time.Now().Add(-1 * time.Hour)},
		Phone: "17888886666",
		Term:  5,
		Code:  9999,
	})
	// 验证码正确, 并且有效
	orm.MySQL.Gaea.Create(&structs.SingleSignOnVerificationCode{
		Phone: "17866886688",
		Term:  5,
		Code:  9999,
	})
}
