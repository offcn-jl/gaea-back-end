/*
   @Time : 2020/11/8 9:14 上午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : suffix_test
   @Software: GoLand
   @Description: 个人后缀业务的服务接口 单元测试
*/

package services

import (
	"bytes"
	"github.com/jarcoal/httpmock"
	"github.com/offcn-jl/gaea-back-end/commons/database/orm"
	"github.com/offcn-jl/gaea-back-end/commons/database/structs"
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

// TestSuffixGetActive 测试 SuffixGetActive 函数是否可以按照预期获取有效的个人后缀
func TestSuffixGetActive(t *testing.T) {
	Convey("测试 SuffixGetActive 函数是否可以按照预期获取有效的个人后缀", t, func() {
		// 重命名表, 创造查询失败的条件
		orm.MySQL.Gaea.DropTable("single_sign_on_suffixes_backup")
		orm.MySQL.Gaea.Exec("RENAME TABLE single_sign_on_suffixes TO single_sign_on_suffixes_backup")

		// 测试执行查询失败
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SuffixGetActive(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Error\":\"Error 1146: Table 'gaea_unit_test.single_sign_on_suffixes' doesn't exist\",\"Message\":\"执行 SQL 查询出错\"}")

		// 恢复表名
		orm.MySQL.Gaea.Exec("RENAME TABLE single_sign_on_suffixes_backup TO single_sign_on_suffixes")

		// 测试未配置查询条件
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		SuffixGetActive(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Data\":[{\"ID\":1,\"Suffix\":\"default\",\"Name\":\"\",\"CRMUser\":\"default\",\"CRMUID\":32431,\"CRMChannel\":7,\"NTalkerGID\":\"NTalkerGID\",\"CRMOID\":1,\"CRMOFID\":0,\"CRMOCode\":22,\"CRMOName\":\"吉林分校\"}],\"Message\":\"Success\"}")
	})
}

// TestSuffixGetDeleting 测试 SuffixGetDeleting 函数是否可以按照预期获取即将过期的后缀
func TestSuffixGetDeleting(t *testing.T) {
	Convey("测试 SuffixGetDeleting 函数是否可以按照预期获取即将过期的后缀", t, func() {
		// 重命名表, 创造查询失败的条件
		orm.MySQL.Gaea.DropTable("single_sign_on_suffixes_backup")
		orm.MySQL.Gaea.Exec("RENAME TABLE single_sign_on_suffixes TO single_sign_on_suffixes_backup")

		// 测试执行查询失败
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		SuffixGetDeleting(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Error\":\"Error 1146: Table 'gaea_unit_test.single_sign_on_suffixes' doesn't exist\",\"Message\":\"执行 SQL 查询出错\"}")

		// 恢复表名
		orm.MySQL.Gaea.Exec("RENAME TABLE single_sign_on_suffixes_backup TO single_sign_on_suffixes")

		// 测试未配置查询条件
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		SuffixGetDeleting(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "{\"Data\":[{\"ID\":2,\"DeletedAt\":\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\",\"Suffix\":\"test\",\"Name\":\"\",\"CRMUser\":\"test\",\"CRMUID\":123,\"CRMChannel\":22,\"NTalkerGID\":\"NTalkerGID\",\"CRMOID\":2,\"CRMOFID\":1,\"CRMOCode\":2290,\"CRMOName\":\"吉林长春分校\"}],\"Message\":\"Success\"}")
	})
}

// TestSuffixGetAvailable 测试 SuffixGetAvailable 函数是否可以按照预期获取所有后缀
func TestSuffixGetAvailable(t *testing.T) {
	Convey("测试 SuffixGetAvailable 函数是否可以按照预期获取所有后缀", t, func() {
		// 重命名表, 创造查询失败的条件
		orm.MySQL.Gaea.DropTable("single_sign_on_suffixes_backup")
		orm.MySQL.Gaea.Exec("RENAME TABLE single_sign_on_suffixes TO single_sign_on_suffixes_backup")

		// 测试执行查询失败
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		SuffixGetAvailable(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Error\":\"Error 1146: Table 'gaea_unit_test.single_sign_on_suffixes' doesn't exist\",\"Message\":\"执行 SQL 查询出错\"}")

		// 恢复表名
		orm.MySQL.Gaea.Exec("RENAME TABLE single_sign_on_suffixes_backup TO single_sign_on_suffixes")

		// 测试未配置查询条件
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		SuffixGetAvailable(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "{\"Data\":[{\"ID\":1,\"Suffix\":\"default\",\"Name\":\"\",\"CRMUser\":\"default\",\"CRMUID\":32431,\"CRMChannel\":7,\"NTalkerGID\":\"NTalkerGID\",\"CRMOID\":1,\"CRMOFID\":0,\"CRMOCode\":22,\"CRMOName\":\"吉林分校\"},{\"ID\":2,\"Suffix\":\"test\",\"Name\":\"\",\"CRMUser\":\"test\",\"CRMUID\":123,\"CRMChannel\":22,\"NTalkerGID\":\"NTalkerGID\",\"CRMOID\":2,\"CRMOFID\":1,\"CRMOCode\":2290,\"CRMOName\":\"吉林长春分校\"},{\"ID\":3,\"Suffix\":\"expired\",\"Name\":\"\",\"CRMUser\":\"expired\",\"CRMUID\":123,\"CRMChannel\":22,\"NTalkerGID\":\"NTalkerGID\",\"CRMOID\":2,\"CRMOFID\":1,\"CRMOCode\":2290,\"CRMOName\":\"吉林长春分校\"}],\"Message\":\"Success\"}")
	})
}

// TestSuffixPushCRM 测试 SuffixPushCRM 函数是否可以按照预期推送带有个人后缀的信息到 CRM
func TestSuffixPushCRM(t *testing.T) {
	Convey("测试 SuffixGetAvailable 函数是否可以按照预期获取所有后缀", t, func() {
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		SuffixPushCRM(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Error\":\"invalid request\",\"Message\":\"提交的 Json 数据不正确\"}")

		// 配置请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"Phone\":\"0\"}"))

		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		SuffixPushCRM(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Error\":\"Key: 'SingleSignOnPushLog.CRMEFSID' Error:Field validation for 'CRMEFSID' failed on the 'required' tag\",\"Message\":\"提交的 Json 数据不正确\"}")

		// 修正请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"Phone\":\"0\",\"CRMEFSID\":\"CRMEFSID\"}"))

		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		SuffixPushCRM(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"手机号码不正确\"}")

		// 修正请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"Phone\":\"17887106666\",\"CRMEFSID\":\"CRMEFSID\"}"))

		// 配置 httpmock 进行拦截
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()

		// 测试 向 CRM 发起请求失败
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		SuffixPushCRM(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Error\":\"Get \\\"https://dc.offcn.com:8443/a.gif?channel=7\\u0026mobile=17887106666\\u0026orgn=2290\\u0026owner=32431\\u0026sid=CRMEFSID\\\": no responder found\",\"Message\":\"向 CRM 发起请求失败\"}")
		// 查询推送失败日志
		singleSignOnErrorLog := structs.SingleSignOnErrorLog{}
		orm.MySQL.Gaea.Find(&singleSignOnErrorLog)
		So(singleSignOnErrorLog.Phone, ShouldEqual, "17887106666")
		So(singleSignOnErrorLog.Error, ShouldEqual, "推送接口 > 请求失败 : Get \"https://dc.offcn.com:8443/a.gif?channel=7&mobile=17887106666&orgn=2290&owner=32431&sid=CRMEFSID\": no responder found")

		// 匹配 URL
		httpmock.RegisterResponder(http.MethodGet, "https://dc.offcn.com:8443/a.gif", httpmock.NewStringResponder(http.StatusInternalServerError, ""))

		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"Phone\":\"17887106666\",\"CRMEFSID\":\"CRMEFSID\"}"))
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		SuffixPushCRM(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"CRM 返回了错误的状态码 : 500\"}")
		// 查询推送失败日志
		singleSignOnErrorLog = structs.SingleSignOnErrorLog{}
		orm.MySQL.Gaea.Find(&singleSignOnErrorLog)
		So(singleSignOnErrorLog.Phone, ShouldEqual, "17887106666")
		So(singleSignOnErrorLog.Error, ShouldEqual, "推送接口 > CRM 返回了错误的状态码 : 500")

		// 测试匹配后缀 无效后缀
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"Phone\":\"17887106666\",\"CRMEFSID\":\"CRMEFSID\",\"Suffix\":\"Suffix\"}"))
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		SuffixPushCRM(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"CRM 返回了错误的状态码 : 500\"}")

		// 测试匹配后缀 非省级
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"Phone\":\"17887106666\",\"CRMEFSID\":\"CRMEFSID\",\"Suffix\":\"test\"}"))
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		SuffixPushCRM(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"CRM 返回了错误的状态码 : 500\"}")

		// 测试匹配后缀 省级
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"Phone\":\"17887106666\",\"CRMEFSID\":\"CRMEFSID\",\"Suffix\":\"default\"}"))
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		SuffixPushCRM(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"CRM 返回了错误的状态码 : 500\"}")

		// 匹配 URL
		httpmock.RegisterResponder(http.MethodGet, "https://dc.offcn.com:8443/a.gif", func(req *http.Request) (*http.Response, error) {
			// 测试请求的参数是否符合预期
			So(req.URL.String(), ShouldEqual, "https://dc.offcn.com:8443/a.gif?channel=7&colleage=CustomerColleage&khsf=1&mayor=CustomerMayor&mobile=17887106666&name=CustomerName&orgn=2290&owner=32431&remark=Remark&sid=CRMEFSID")
			return httpmock.NewStringResponse(http.StatusOK, ""), nil
		})

		// 测试推送成功
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"Phone\":\"17887106666\",\"CRMEFSID\":\"CRMEFSID\",\"CustomerName\":\"CustomerName\",\"CustomerIdentityID\":1,\"CustomerColleage\":\"CustomerColleage\",\"CustomerMayor\":\"CustomerMayor\",\"Remark\":\"Remark\"}"))
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		SuffixPushCRM(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"Success\"}")

		// 测试重复推送
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"Phone\":\"17887106666\",\"CRMEFSID\":\"CRMEFSID\"}"))
		utt.HttpTestResponseRecorder.Body.Reset() // 再次测试前重置 body
		SuffixPushCRM(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"Success\"}")
	})
}
