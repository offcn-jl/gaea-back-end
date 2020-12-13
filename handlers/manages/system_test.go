/*
   @Time : 2020/12/3 3:17 下午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : system_test
   @Software: GoLand
   @Description: 系统服务的单元测试
*/

package manages

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/gin-gonic/gin"
	"github.com/jarcoal/httpmock"
	"github.com/offcn-jl/gaea-back-end/commons/config"
	"github.com/offcn-jl/gaea-back-end/commons/database/orm"
	"github.com/offcn-jl/gaea-back-end/commons/database/structs"
	"github.com/offcn-jl/gaea-back-end/commons/encrypt"
	"github.com/offcn-jl/gaea-back-end/commons/response"
	"github.com/offcn-jl/gaea-back-end/commons/utt"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"testing"
)

// 覆盖 orm 库中的 ORM 对象
func init() {
	utt.InitTest() // 初始化测试数据并获取测试所需的上下文
	orm.MySQL.Gaea = utt.ORM
}

// TestSystemGetRSAPublicKey 测试 SystemGetRSAPublicKey 函数是否可以获取 RSA 公钥
func TestSystemGetRSAPublicKey(t *testing.T) {
	Convey("测试 SystemGetRSAPublicKey 函数是否可以获取 RSA 公钥", t, func() {
		// 测试未配置公钥
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemGetRSAPublicKey(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldBeEmpty)

		// 配置公钥
		config.Update(orm.MySQL.Gaea, structs.SystemConfig{RSAPublicKey: "fake-rsa-public-key"})

		// 测试已配置公钥
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemGetRSAPublicKey(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "fake-rsa-public-key")
	})
}

// TestSystemLogin 测试 SystemLogin 是否可以进行用户登陆操作
func TestSystemLogin(t *testing.T) {
	Convey("测试 SystemLogin 是否可以进行用户登陆操作", t, func() {
		// 测试 绑定参数错误
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemLogin(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Error\":\"invalid request\",\"Message\":\"提交的 Json 数据不正确\"}")

		// 增加 Body
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{}"))

		// 测试 校验参数
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemLogin(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Error\":\"Key: 'Username' Error:Field validation for 'Username' failed on the 'required' tag\\nKey: 'Password' Error:Field validation for 'Password' failed on the 'required' tag\\nKey: 'MisToken' Error:Field validation for 'MisToken' failed on the 'required' tag\",\"Message\":\"提交的 Json 数据不正确\"}")

		// 修正请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"UserName\":\"fake-username\",\"Password\":\"fake-password\",\"MisToken\":\"fake-token\"}"))

		// 测试 对比 Mis 口令码失败
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemLogin(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\"Message\":\"校验 Mis 口令码失败\"")

		// 配置 httpmock 进行拦截
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterNoResponder(httpmock.NewJsonResponderOrPanic(http.StatusOK, response.Struct{"status": 1, "msg": "SIGN验签成功", "data": "fake-token"}))

		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"UserName\":\"fake-username\",\"Password\":\"fake-password\",\"MisToken\":\"wrong-fake-token\"}"))

		// 测试 Mis 口令码不正确
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemLogin(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"Mis 口令码不正确\"}")

		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"UserName\":\"fake-username\",\"Password\":\"fake-password\",\"MisToken\":\"fake-token\"}"))

		// 测试 用户不存在
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemLogin(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"用户不存在或已经被禁用\"}")

		// 创建用户
		utt.ORM.Create(&structs.SystemUser{Username: "fake-username"})

		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"UserName\":\"fake-username\",\"Password\":\"fake-password\",\"MisToken\":\"fake-token\"}"))

		// 测试 请求中的用户密码 RSA 解密失败
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemLogin(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Error\":\"illegal base64 data at input byte 4\",\"Message\":\"请求中的用户密码 RSA 解密失败\"}")

		// 生成私钥
		// # https://www.cnblogs.com/PeterXu1997/p/12218553.html
		// # https://blog.csdn.net/chenxing1230/article/details/83757638
		// 生成 RSA 密钥对
		// GenerateKey 函数使用随机数据生成器random生成一对具有指定字位数的RSA密钥
		// Reader 是一个全局、共享的密码用强随机数生成器
		privateKey, _ := rsa.GenerateKey(rand.Reader, 4096)
		// 通过 x509 标准将得到的 ras 私钥序列化为 ASN.1 的 DER 编码字符串
		X509PrivateKey := x509.MarshalPKCS1PrivateKey(privateKey)
		// 构建一个 pem.Block 结构体对象
		privateBlock := pem.Block{Type: "RSA Private Key", Bytes: X509PrivateKey}
		// 初始化用于接收 pem 的 buffer
		bufferPrivate := new(bytes.Buffer)
		// 使用 pem 格式对 x509 输出的内容进行编码
		pem.Encode(bufferPrivate, &privateBlock)

		// 生成公钥
		// X509 对公钥编码
		X509PublicKey, _ := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
		//创建一个pem.Block结构体对象
		publicBlock := pem.Block{Type: "RSA Public Key", Bytes: X509PublicKey}
		// 初始化用于接收 pem 的 buffer
		bufferPublic := new(bytes.Buffer)
		// pem格式编码
		pem.Encode(bufferPublic, &publicBlock)

		// 添加 RSA 配置
		config.Update(utt.ORM, structs.SystemConfig{RSAPublicKey: bufferPublic.String(), RSAPrivateKey: bufferPrivate.String()})

		// 使用前面的步骤中生成的 RSA 公钥对密码进行加密
		encryptedRequestPassword, _ := encrypt.RSAEncrypt([]byte("fake-password"))
		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"UserName\":\"fake-username\",\"Password\":\""+encryptedRequestPassword+"\",\"MisToken\":\"fake-token\"}"))
		// 测试 数据库中的用户密码 RSA 解密失败
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemLogin(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Error\":\"crypto/rsa: decryption error\",\"Message\":\"数据库中的用户密码 RSA 解密失败\"}")

		// 使用前面的步骤中生成的 RSA 公钥对密码进行加密
		encryptedRequestPassword, _ = encrypt.RSAEncrypt([]byte("wrong-fake-password"))
		encryptedDatabasePassword, _ := encrypt.RSAEncrypt([]byte("fake-password"))
		// 修改数据库中的用户密码
		utt.ORM.Model(structs.SystemUser{}).Update(&structs.SystemUser{Username: "fake-username", Password: encryptedDatabasePassword})
		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"UserName\":\"fake-username\",\"Password\":\""+encryptedRequestPassword+"\",\"MisToken\":\"fake-token\"}"))
		// 测试密码不正确
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemLogin(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"密码不正确, 已经登陆失败 1 次\"}")

		// 修改请求的密码为正确的密码
		encryptedRequestPassword, _ = encrypt.RSAEncrypt([]byte("fake-password"))
		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"UserName\":\"fake-username\",\"Password\":\""+encryptedRequestPassword+"\",\"MisToken\":\"fake-token\"}"))
		// 测试登陆成功
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemLogin(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\"Message\":\"Success\"")
		// 检查是否存在会话记录
		sessionInfo := structs.SystemSession{}
		utt.ORM.Find(&sessionInfo)
		So(sessionInfo.MisToken, ShouldEqual, "fake-token")
		So(sessionInfo.UserID, ShouldEqual, 1)

		// 添加登陆失败记录
		for i := 0; i < 5; i++ {
			utt.ORM.Create(&structs.SystemUserLoginFailLog{UserID: 1})
		}

		// 测试连续登陆失败超过 5 次
		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"UserName\":\"fake-username\",\"Password\":\""+encryptedRequestPassword+"\",\"MisToken\":\"fake-token\"}"))
		// 测试密码不正确
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemLogin(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"用户 24 小时内连续登陆失败 5 次, 已经被暂时冻结, 24 小时后自动解冻\"}")
	})
}

// TestSystemLogout 测试 SystemLogout 是否可以进行退出 ( 销毁会话 ) 操作
func TestSystemLogout(t *testing.T) {
	Convey("测试 SystemLogout 是否可以进行退出 ( 销毁会话 ) 操作", t, func() {
		// 模拟登陆
		// 清除前一测试中添加的登陆失败记录
		utt.ORM.Delete(&structs.SystemUserLoginFailLog{})
		// 配置 httpmock 进行拦截
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterNoResponder(httpmock.NewJsonResponderOrPanic(http.StatusOK, response.Struct{"status": 1, "msg": "SIGN验签成功", "data": "fake-token"}))
		// 加密登陆密码
		encryptedRequestPassword, _ := encrypt.RSAEncrypt([]byte("fake-password"))
		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"UserName\":\"fake-username\",\"Password\":\""+encryptedRequestPassword+"\",\"MisToken\":\"fake-token\"}"))
		// 进行登陆操作
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemLogin(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\"Message\":\"Success\"")
		// 检查是否存在会话记录
		sessionInfo := structs.SystemSession{}
		utt.ORM.Find(&sessionInfo)
		So(sessionInfo.MisToken, ShouldEqual, "fake-token")
		So(sessionInfo.UserID, ShouldEqual, 1)
		So(sessionInfo.DeletedAt, ShouldBeNil)

		// 配置 UUID 参数
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "UUID", Value: sessionInfo.UUID}}

		// 测试退出操作
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemLogout(utt.GinTestContext)

		// 检查会话是否已经被销毁
		utt.ORM.Unscoped().Find(&sessionInfo)
		So(sessionInfo.MisToken, ShouldEqual, "fake-token")
		So(sessionInfo.UserID, ShouldEqual, 1)
		So(sessionInfo.DeletedAt, ShouldNotBeNil)
	})
}

// TestSystemUpdateMisToken 测试 SystemUpdateMisToken 是否可以进行更新 Mis 口令码操作
func TestSystemUpdateMisToken(t *testing.T) {
	Convey("测试 SystemUpdateMisToken 是否可以进行更新 Mis 口令码操作", t, func() {
		// 测试会话无效
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemUpdateMisToken(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"会话无效\"}")

		// 模拟登陆
		// 配置 httpmock 进行拦截
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterNoResponder(httpmock.NewJsonResponderOrPanic(http.StatusOK, response.Struct{"status": 1, "msg": "SIGN验签成功", "data": "fake-token"}))
		// 加密登陆密码
		encryptedRequestPassword, _ := encrypt.RSAEncrypt([]byte("fake-password"))
		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"UserName\":\"fake-username\",\"Password\":\""+encryptedRequestPassword+"\",\"MisToken\":\"fake-token\"}"))
		// 进行登陆操作
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemLogin(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\"Message\":\"Success\"")
		// 检查是否存在会话记录
		sessionInfo := structs.SystemSession{}
		utt.ORM.Find(&sessionInfo)
		So(sessionInfo.MisToken, ShouldEqual, "fake-token")
		So(sessionInfo.UserID, ShouldEqual, 1)
		So(sessionInfo.DeletedAt, ShouldBeNil)
		// 将会话 UUID 配置到请求上下文中
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "UUID", Value: sessionInfo.UUID}}

		// 修改 httpmock 为获取口令码失败
		httpmock.RegisterNoResponder(httpmock.NewJsonResponderOrPanic(http.StatusOK, response.Struct{"status": 2}))

		// 测试校验口令码失败
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemUpdateMisToken(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Error\":\"请求 MIS 口令码 失败\",\"Message\":\"校验 Mis 口令码失败\"}")

		// 修改 httpmock 为获取口令码成功
		httpmock.RegisterNoResponder(httpmock.NewJsonResponderOrPanic(http.StatusOK, response.Struct{"status": 1, "msg": "SIGN验签成功", "data": "new-fake-token"}))

		// 测试口令码不正确
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemUpdateMisToken(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"Mis 口令码不正确\"}")

		// 将正确的 Mis 口令码 配置到请求上下文中
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "UUID", Value: sessionInfo.UUID}, gin.Param{Key: "MisToken", Value: "new-fake-token"}}

		// 测试更新成功
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemUpdateMisToken(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"Success\"}")
		// 检查记录中的信息是否更新
		utt.ORM.Where("id = ?", sessionInfo.ID).Find(&sessionInfo)
		So(sessionInfo.MisToken, ShouldEqual, "new-fake-token")
	})
}
