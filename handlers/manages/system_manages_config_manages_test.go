/*
   @Time : 2021/1/8 8:34 上午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : system_manages_config_manages_test.go
   @Description: [ 单元测试 ] 系统管理 - 配置管理
*/

package manages

import (
	"bytes"
	"encoding/json"
	"github.com/offcn-jl/gaea-back-end/commons/config"
	"github.com/offcn-jl/gaea-back-end/commons/database/orm"
	"github.com/offcn-jl/gaea-back-end/commons/database/structs"
	"github.com/offcn-jl/gaea-back-end/commons/utt"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"testing"
	"time"
)

// 初始化测试数据并获取测试所需的上下文
func init() {
	utt.InitTest()
	orm.MySQL.Gaea = utt.ORM // 覆盖 orm 库中的 ORM 对象
}

// TestSystemManagesConfigManagesPaginationGetConfig 测试 SystemManagesConfigManagesPaginationGetConfig 是否可以分页获取配置列表
func TestSystemManagesConfigManagesPaginationGetConfig(t *testing.T) {
	Convey("测试 SystemManagesConfigManagesPaginationGetConfig 是否可以分页获取配置列表", t, func() {
		// 测试 校验参数
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesConfigManagesPaginationGetConfig(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "Data")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "Message")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "Total")
	})
}

// TestSystemManagesConfigManagesUpdateConfig 测试 SystemManagesConfigManagesUpdateConfig 是否可以修改配置
func TestSystemManagesConfigManagesUpdateConfig(t *testing.T) {
	Convey("测试 SystemManagesConfigManagesUpdateConfig 是否可以修改配置", t, func() {
		// 重置请求上下文
		utt.ResetContext()

		// 测试绑定数据错误
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesConfigManagesUpdateConfig(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Error\":\"invalid request\",\"Message\":\"提交的 Json 数据不正确\"}")

		// 随机生成一个新配置
		newConfig := structs.SystemConfig{DisableDebug: false, CORSRuleServices: time.Now().Format("20060102150405"), CORSRuleManages: "-", CORSRuleEvents: "-", OffcnSmsURL: "-", OffcnSmsUserName: "-", OffcnSmsPassword: "-", OffcnSmsTjCode: "-", OffcnMisURL: "-", OffcnMisAppID: "-", OffcnMisToken: "-", OffcnMisCode: "-", OffcnOCCKey: "-", TencentCloudAPISecretID: "-", TencentCloudAPISecretKey: "-", TencentCloudSmsSdkAppId: "-", ServicesAccessToken: "-", RSAPublicKey: "-", RSAPrivateKey: "-"}

		// 增加 Body
		jsonBytes, _ := json.Marshal(newConfig)
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewReader(jsonBytes))

		// 测试修改配置成功
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		SystemManagesConfigManagesUpdateConfig(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"Success\"}")

		// 取出数据库中的新配置
		newConfigInDB := structs.SystemConfig{}
		orm.MySQL.Gaea.Last(&newConfigInDB)

		// 检查配置是否修改成功
		So(newConfig.ID, ShouldNotEqual, newConfigInDB.ID)
		So(newConfigInDB.ID, ShouldEqual, config.Get().ID)
		So(newConfig.CORSRuleServices, ShouldEqual, newConfigInDB.CORSRuleServices)

		// 清空 Request
		utt.GinTestContext.Request, _ = http.NewRequest("GET", "/", nil)
	})
}
