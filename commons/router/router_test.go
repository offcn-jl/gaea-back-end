/*
   @Time : 2020/12/13 4:04 下午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : router_test
   @Software: GoLand
   @Description: 路由及路由相关的业务的单元测试
*/

package router

import (
	"fmt"
	"github.com/jarcoal/httpmock"
	"github.com/offcn-jl/gaea-back-end/commons/database/orm"
	"github.com/offcn-jl/gaea-back-end/commons/database/structs"
	"github.com/offcn-jl/gaea-back-end/commons/response"
	"github.com/offcn-jl/gaea-back-end/commons/utt"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// 初始化测试数据并获取测试所需的上下文
func init() {
	utt.InitTest()
	orm.MySQL.Gaea = utt.ORM // 覆盖 orm 库中的 ORM 对象
}

// TestInitRouter 测试 InitRouter 是否可以初始化路由
func TestInitRouter(t *testing.T) {
	Convey("测试 initRouter 是否可以初始化路由", t, func() {
		ginEngine := InitRouter("/test")
		So(ginEngine.AppEngine, ShouldBeFalse)

		testServer := httptest.NewServer(ginEngine)

		// 测试实际响应
		res, err := http.Get(fmt.Sprintf("%s/fake-path", testServer.URL))
		So(err, ShouldBeNil)
		resp, err := ioutil.ReadAll(res.Body)
		So(err, ShouldBeNil)
		So(string(resp), ShouldEqual, "{\"Message\":\"路径有误\"}")
	})
}

// Test_corsCheck 测试 corsCheck CORS 是否可以进行跨域检查
func Test_corsCheck(t *testing.T) {
	Convey("测试 corsCheck CORS 是否可以进行跨域检查", t, func() {
		// 初始化 Gin 上下文中的 Request
		utt.GinTestContext.Request, _ = http.NewRequest(http.MethodGet, "/favicon.ico", nil)

		So(utt.GinTestContext.IsAborted(), ShouldBeFalse)

		// 测试跳过 /favicon.ico
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		corsCheck(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Code, ShouldEqual, http.StatusNotFound)
		So(utt.GinTestContext.IsAborted(), ShouldBeTrue) // 测试 Abort 是否成功

		// 修改 Request 为请求非标准路径
		utt.GinTestContext.Request, _ = http.NewRequest(http.MethodGet, "/fake-path", nil)

		// 测试不是标准路径
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		corsCheck(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Code, ShouldEqual, http.StatusNotFound)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"路径有误\"}")
		So(utt.GinTestContext.IsAborted(), ShouldBeTrue) // 测试 Abort 是否成功

		// 修改 Request 为请求非标准路径
		utt.GinTestContext.Request, _ = http.NewRequest(http.MethodGet, "/fake-path/fake-path", nil)

		// 测试不是标准路径
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		corsCheck(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Code, ShouldEqual, http.StatusNotFound)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"路径有误\"}")
		So(utt.GinTestContext.IsAborted(), ShouldBeTrue) // 测试 Abort 是否成功

		// 修改 Request 为请求标准路径
		utt.GinTestContext.Request, _ = http.NewRequest(http.MethodGet, "/test/services/fake-path", nil)
		utt.GinTestContext.Request.Header.Set("origin", "origin")

		// 测试未通过检查
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		corsCheck(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Code, ShouldEqual, http.StatusNotFound)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"请求未通过跨域检查\"}")
		So(utt.GinTestContext.IsAborted(), ShouldBeTrue) // 测试 Abort 是否成功

		// 修改 Request 为请求标准路径
		utt.GinTestContext.Request, _ = http.NewRequest(http.MethodOptions, "/test/services/fake-path", nil)
		utt.GinTestContext.Request.Header.Set("Access-Control-Request-Headers", "Access-Control-Request-Headers")
		utt.GinTestContext.Request.Header.Set("Access-Control-Request-Method", "Access-Control-Request-Method")
		utt.GinTestContext.Request.Method = "OPTIONS"

		// 测试通过检查后放行 OPTIONS 方法
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		corsCheck(utt.GinTestContext)

		So(utt.GinTestContext.GetHeader("Access-Control-Request-Headers"), ShouldEqual, "Access-Control-Request-Headers")
		So(utt.GinTestContext.GetHeader("Access-Control-Request-Method"), ShouldEqual, "Access-Control-Request-Method")
		So(utt.GinTestContext.IsAborted(), ShouldBeTrue) // 测试 Abort 是否成功

		// 测试匹配检查规则

		// 修改 Request 为请求标准路径
		utt.GinTestContext.Request, _ = http.NewRequest(http.MethodGet, "/test/manages/fake-path", nil)
		utt.GinTestContext.Request.Header.Set("origin", "origin")

		// 测试未通过检查
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		corsCheck(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Code, ShouldEqual, http.StatusNotFound)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"请求未通过跨域检查\"}")
		So(utt.GinTestContext.IsAborted(), ShouldBeTrue) // 测试 Abort 是否成功

		// 修改 Request 为请求标准路径
		utt.GinTestContext.Request, _ = http.NewRequest(http.MethodGet, "/test/events/fake-path", nil)
		utt.GinTestContext.Request.Header.Set("origin", "origin")

		// 测试未通过检查
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		corsCheck(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Code, ShouldEqual, http.StatusNotFound)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"请求未通过跨域检查\"}")
		So(utt.GinTestContext.IsAborted(), ShouldBeTrue) // 测试 Abort 是否成功

		// 测试匹配检查规则 结束

		// 修改 Request 为请求标准路径
		utt.GinTestContext.Request, _ = http.NewRequest(http.MethodGet, "/test/events/fake-path", nil)

		// 测试通过检查
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		corsCheck(utt.GinTestContext)
		So(utt.GinTestContext.IsAborted(), ShouldBeTrue) // 测试 Abort 是否成功
	})
}

// Test_checkSessionAndPermission 测试 checkSessionAndPermission 是否可以检查会话有效性与接口访问权限
func Test_checkSessionAndPermission(t *testing.T) {
	Convey("测试 checkSessionAndPermission 是否可以检查会话有效性与接口访问权限", t, func() {
		// 测试未配置鉴权信息导致会话无效
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		handler := checkSessionAndPermission("")
		handler(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"会话无效\"}")
		So(utt.GinTestContext.IsAborted(), ShouldBeTrue) // 测试 Abort 是否成功

		// 配置鉴权信息
		utt.GinTestContext.Request.Header.Set("Authorization", "Gaea fake-uuid")

		// 测试 UUID 无效导致会话无效
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		handler = checkSessionAndPermission("")
		handler(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"会话无效\"}")
		So(utt.GinTestContext.IsAborted(), ShouldBeTrue) // 测试 Abort 是否成功

		// 创建测试用会话记录
		lastRequestAt := time.Now()
		utt.ORM.Create(&structs.SystemSession{UserID: 1, UUID: "wrong-fake-uuid", LastRequestAt: lastRequestAt})
		utt.ORM.Create(&structs.SystemSession{UserID: 1, UUID: "fake-uuid", LastRequestAt: lastRequestAt})

		// 配置鉴权信息
		utt.GinTestContext.Request.Header.Set("Authorization", "Gaea wrong-fake-uuid")

		// 测试 UUID 与最后一条不匹配导致的对话无效
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		handler = checkSessionAndPermission("")
		handler(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"会话无效\"}")
		So(utt.GinTestContext.IsAborted(), ShouldBeTrue) // 测试 Abort 是否成功

		// 配置鉴权信息
		utt.GinTestContext.Request.Header.Set("Authorization", "Gaea fake-uuid")

		// 测试没有绑定 MIS Token 导致的 Mis 口令码无效
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		handler = checkSessionAndPermission("")
		handler(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"Mis 口令码无效\"}")
		So(utt.GinTestContext.IsAborted(), ShouldBeTrue) // 测试 Abort 是否成功

		// 配置 httpmock 进行拦截
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterNoResponder(httpmock.NewJsonResponderOrPanic(http.StatusOK, response.Struct{"status": 2, "msg": "时间戳缺失或失效"}))

		// 添加 Mis 口令码配置到会话记录中
		utt.ORM.Model(structs.SystemSession{}).Update(&structs.SystemSession{MisToken: "fake-mis-token"})

		// 配置鉴权信息
		utt.GinTestContext.Request.Header.Set("Authorization", "Gaea fake-uuid")

		// 测试系统故障导致到校验口令码失败
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		handler = checkSessionAndPermission("")
		handler(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Error\":\"时间戳缺失或失效\",\"Message\":\"校验 Mis 口令码失败\"}")
		So(utt.GinTestContext.IsAborted(), ShouldBeTrue) // 测试 Abort 是否成功

		// 修改 httpmock 为返回不一致的口令码
		httpmock.RegisterNoResponder(httpmock.NewJsonResponderOrPanic(http.StatusOK, response.Struct{"status": 1, "msg": "SIGN验签成功", "data": "wrong-fake-mis-token"}))

		// 测试口令码无效
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		handler = checkSessionAndPermission("")
		handler(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"Mis 口令码无效\"}")
		So(utt.GinTestContext.IsAborted(), ShouldBeTrue) // 测试 Abort 是否成功

		// 修改 httpmock 为返回一致的口令码
		httpmock.RegisterNoResponder(httpmock.NewJsonResponderOrPanic(http.StatusOK, response.Struct{"status": 1, "msg": "SIGN验签成功", "data": "fake-mis-token"}))

		// 测试口令码有效, 且无需验证权限
		handler = checkSessionAndPermission("")
		handler(utt.GinTestContext)
		userInfo, exists := utt.GinTestContext.Get("UserInfo")
		So(exists, ShouldBeTrue)
		So(userInfo.(structs.SystemUser).ID, ShouldEqual, 0)
		roleInfo, exists := utt.GinTestContext.Get("RoleInfo")
		So(exists, ShouldBeTrue)
		So(roleInfo.(structs.SystemRole).ID, ShouldEqual, 0)

		// 测试校验权限是反序列化角色权限失败
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		handler = checkSessionAndPermission("fake-permission")
		handler(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Error\":\"unexpected end of JSON input\",\"Message\":\"反序列化角色权限配置失败\"}")
		So(utt.GinTestContext.IsAborted(), ShouldBeTrue) // 测试 Abort 是否成功

		// 添加用于测试的用户
		orm.MySQL.Gaea.Create(&structs.SystemUser{RoleID: 1})
		// 添加用于测试的角色
		orm.MySQL.Gaea.Create(&structs.SystemRole{Permissions: "[\"fake-permission-1\",\"fake-permission-2\"]"})

		// 测试角色没有目标权限
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		handler = checkSessionAndPermission("fake-permission")
		handler(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"没有接口访问权限\"}")
		So(utt.GinTestContext.IsAborted(), ShouldBeTrue) // 测试 Abort 是否成功

		// 测试通过全部检查
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		handler = checkSessionAndPermission("fake-permission-2")
		handler(utt.GinTestContext)
		userInfo, exists = utt.GinTestContext.Get("UserInfo")
		So(exists, ShouldBeTrue)
		So(userInfo.(structs.SystemUser).ID, ShouldEqual, 1)
		roleInfo, exists = utt.GinTestContext.Get("RoleInfo")
		So(exists, ShouldBeTrue)
		So(roleInfo.(structs.SystemRole).ID, ShouldEqual, 1)
	})
}
