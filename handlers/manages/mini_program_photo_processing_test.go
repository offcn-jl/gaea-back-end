/*
   @Time : 2021/7/8 3:36 下午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : mini_program_photo_processing_test.go
   @Package : manages
   @Description: 单元测试 小程序 证件照
*/

package manages

import (
	"bytes"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/offcn-jl/gaea-back-end/commons/database/orm"
	"github.com/offcn-jl/gaea-back-end/commons/database/structs"
	"github.com/offcn-jl/gaea-back-end/commons/utt"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"testing"
)

// 初始化测试数据并获取测试所需的上下文
func init() {
	utt.InitTest()
	orm.MySQL.Gaea = utt.ORM // 覆盖 orm 库中的 ORM 对象
}

// createMiniProgramPhotoProcessingTestUserAndRoleData 创建测试用的用户与角色信息
func createMiniProgramPhotoProcessingTestUserAndRoleData() {
	orm.MySQL.Gaea.DropTableIfExists(structs.SystemRole{})
	orm.MySQL.Gaea.AutoMigrate(structs.SystemRole{})
	orm.MySQL.Gaea.DropTableIfExists(structs.SystemUser{})
	orm.MySQL.Gaea.AutoMigrate(structs.SystemUser{})
	orm.MySQL.Gaea.Create(&structs.SystemRole{Model: gorm.Model{ID: 1101}, Name: "1101"})
	orm.MySQL.Gaea.Create(&structs.SystemRole{Model: gorm.Model{ID: 1102}, Name: "1102", SuperiorID: 1101})
	orm.MySQL.Gaea.Create(&structs.SystemUser{Model: gorm.Model{ID: 111}, Name: "111", Username: "111", RoleID: 1101, CreatedUserID: 1, UpdatedUserID: 1})
	orm.MySQL.Gaea.Create(&structs.SystemUser{Model: gorm.Model{ID: 112}, Name: "112", Username: "112", RoleID: 1102, CreatedUserID: 1, UpdatedUserID: 1})
}

// TestMiniProgramPhotoProcessingCreate 测试 MiniProgramPhotoProcessingCreate 是否能按照预期完成新建
func TestMiniProgramPhotoProcessingCreate(t *testing.T) {
	Convey("测试 MiniProgramPhotoProcessingCreate 是否能按照预期完成新建", t, func() {
		// 重置请求上下文
		utt.ResetContext()

		// 测试 绑定参数
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		MiniProgramPhotoProcessingCreate(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Error\":\"invalid request\",\"Message\":\"提交的 Json 数据不正确\"}")

		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"Name\": \"Name\",\"Project\":\"Project\",\"CRMEventFormID\":1,\"CRMEventFormSID\":\"CRMEventFormSID\",\"MillimeterWidth\":1,\"MillimeterHeight\":1,\"PixelWidth\":1,\"PixelHeight\":1,\"BackgroundColors\":\"BackgroundColors\",\"Description\":\"Description\",\"Hot\":true}"))

		// 测试 校验参数是否合法
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		MiniProgramPhotoProcessingCreate(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"CRM 活动表单 SID 格式不正确\"}")

		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"Name\": \"Name\",\"Project\":\"Project\",\"CRMEventFormID\":1,\"CRMEventFormSID\":\"6936c8475486621c61ad6a9d0865ae7a\",\"MillimeterWidth\":1,\"MillimeterHeight\":1,\"PixelWidth\":1,\"PixelHeight\":1,\"BackgroundColors\":\"BackgroundColors\",\"Description\":\"Description\",\"Hot\":true}"))

		// 测试 创建成功
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		MiniProgramPhotoProcessingCreate(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"Success\"}")
	})
}

// TestMiniProgramPhotoProcessingUpdate 测试 MiniProgramPhotoProcessingUpdate 是否能按照预期完成修改
func TestMiniProgramPhotoProcessingUpdate(t *testing.T) {
	Convey("测试 MiniProgramPhotoProcessingUpdate 是否能按照预期完成修改", t, func() {
		// 重置请求上下文
		utt.ResetContext()

		// 测试 绑定参数
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		MiniProgramPhotoProcessingUpdate(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Error\":\"invalid request\",\"Message\":\"提交的 Json 数据不正确\"}")

		// 创建测试用的用户与角色信息
		createMiniProgramPhotoProcessingTestUserAndRoleData()

		// 创建用于测试的数据
		testInfo := structs.MiniProgramPhotoProcessingConfig{CreatedUserID: 111, UpdatedUserID: 111, Name: "Name", Project: "Project", CRMEventFormID: 1, CRMEventFormSID: "CRMEventFormSID", MillimeterWidth: 1, MillimeterHeight: 1, PixelWidth: 1, PixelHeight: 1, BackgroundColors: "BackgroundColors", Description: "Description", Hot: false}
		orm.MySQL.Gaea.Create(&testInfo)

		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"ID\": "+fmt.Sprint(testInfo.ID)+",\"Name\": \"Name\",\"Project\":\"Project\",\"CRMEventFormID\":1,\"CRMEventFormSID\":\"CRMEventFormSID\",\"MillimeterWidth\":1,\"MillimeterHeight\":1,\"PixelWidth\":1,\"PixelHeight\":1,\"BackgroundColors\":\"BackgroundColors\",\"Description\":\"Description\",\"Hot\":true}"))

		// 测试 校验参数是否合法
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		MiniProgramPhotoProcessingUpdate(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"CRM 活动表单 SID 格式不正确\"}")

		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"ID\": "+fmt.Sprint(testInfo.ID)+",\"Name\": \"new-name\",\"Project\":\"new-project\",\"CRMEventFormID\":1,\"CRMEventFormSID\":\"6936c8475486621c61ad6a9d0865ae7a\",\"MillimeterWidth\":2,\"MillimeterHeight\":2,\"PixelWidth\":2,\"PixelHeight\":2,\"BackgroundColors\":\"new-background-colors\",\"Description\":\"new-description\",\"Hot\":false}"))

		// 测试 检查当前用户所属角色是否有操作目标角色的权限
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		MiniProgramPhotoProcessingUpdate(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"该证件照配置的创建角色不是当前用户所属角色, 并且不是当前用户所属角色的下属角色\"}")

		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"ID\": "+fmt.Sprint(testInfo.ID)+",\"Name\": \"new-name\",\"Project\":\"new-project\",\"CRMEventFormID\":2,\"CRMEventFormSID\":\"6936c8475486621c61ad6a9d0865ae7a\",\"MillimeterWidth\":2,\"MillimeterHeight\":2,\"PixelWidth\":2,\"PixelHeight\":2,\"BackgroundColors\":\"new-background-colors\",\"Description\":\"new-description\",\"Hot\":false}"))

		// 保存 当前角色信息到上下文
		utt.GinTestContext.Set("RoleInfo", structs.SystemRole{Model: gorm.Model{ID: 1102}, Name: "1102", SuperiorID: 1101})

		// 测试 检查当前用户所属角色是否有操作目标角色的权限
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		MiniProgramPhotoProcessingUpdate(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"该证件照配置的创建角色不是当前用户所属角色, 并且不是当前用户所属角色的下属角色\"}")

		// 添加请求内容
		utt.GinTestContext.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"ID\": "+fmt.Sprint(testInfo.ID)+",\"DeletedAt\":\"2021-07-09T09:35:01+08:00\",\"Name\": \"new-name\",\"Project\":\"new-project\",\"CRMEventFormID\":2,\"CRMEventFormSID\":\"6936c8475486621c61ad6a9d0865ae7a\",\"MillimeterWidth\":2,\"MillimeterHeight\":2,\"PixelWidth\":2,\"PixelHeight\":2,\"BackgroundColors\":\"new-background-colors\",\"Description\":\"new-description\",\"Hot\":false}"))

		// 保存 当前角色信息到上下文
		utt.GinTestContext.Set("RoleInfo", structs.SystemRole{Model: gorm.Model{ID: 1101}, Name: "1101", SuperiorID: 0})

		// 测试 操作成功
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		MiniProgramPhotoProcessingUpdate(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"Success\"}")
		checkInfo := structs.MiniProgramPhotoProcessingConfig{}
		orm.MySQL.Gaea.Unscoped().Where("id = ?", testInfo.ID).Find(&checkInfo)
		So(checkInfo.DeletedAt, ShouldNotBeNil)
		So(checkInfo.Name, ShouldEqual, "new-name")
		So(checkInfo.Project, ShouldEqual, "new-project")
		So(checkInfo.CRMEventFormID, ShouldEqual, 2)
		So(checkInfo.MillimeterWidth, ShouldEqual, 2)
		So(checkInfo.MillimeterHeight, ShouldEqual, 2)
		So(checkInfo.PixelWidth, ShouldEqual, 2)
		So(checkInfo.PixelHeight, ShouldEqual, 2)
		So(checkInfo.BackgroundColors, ShouldEqual, "new-background-colors")
		So(checkInfo.Description, ShouldEqual, "new-description")
		So(checkInfo.Hot, ShouldEqual, false)
	})
}

// TestMiniProgramPhotoProcessingGetList 测试 MiniProgramPhotoProcessingGetList 是否能够按照预期获取列表 带搜索
func TestMiniProgramPhotoProcessingGetList(t *testing.T) {
	Convey("测试 MiniProgramPhotoProcessingGetList 是否能够按照预期获取列表 带搜索", t, func() {
		// 创建测试用的用户与角色信息
		createMiniProgramPhotoProcessingTestUserAndRoleData()

		// 创建测试数据
		orm.MySQL.Gaea.DropTableIfExists(structs.MiniProgramPhotoProcessingConfig{})
		orm.MySQL.Gaea.AutoMigrate(structs.MiniProgramPhotoProcessingConfig{})
		orm.MySQL.Gaea.Create(&structs.MiniProgramPhotoProcessingConfig{CreatedUserID: 111, UpdatedUserID: 111, Name: "名称1", Project: "项目1", CRMEventFormID: 1, CRMEventFormSID: "6936c8475486621c61ad6a9d0865ae7a", MillimeterWidth: 1, MillimeterHeight: 1, PixelWidth: 1, PixelHeight: 1, BackgroundColors: "BackgroundColors", Description: "Description", Hot: false})
		orm.MySQL.Gaea.Create(&structs.MiniProgramPhotoProcessingConfig{CreatedUserID: 112, UpdatedUserID: 112, Name: "名称2", Project: "项目1", CRMEventFormID: 2, CRMEventFormSID: "6936c8475486621c61ad6a9d0865ae7b", MillimeterWidth: 1, MillimeterHeight: 1, PixelWidth: 1, PixelHeight: 1, BackgroundColors: "BackgroundColors", Description: "Description", Hot: true})
		orm.MySQL.Gaea.Create(&structs.MiniProgramPhotoProcessingConfig{CreatedUserID: 112, UpdatedUserID: 112, Name: "测试名1", Project: "项目2", CRMEventFormID: 3, CRMEventFormSID: "6936c8475486621c61ad6a9d0865ae7c", MillimeterWidth: 1, MillimeterHeight: 1, PixelWidth: 1, PixelHeight: 1, BackgroundColors: "BackgroundColors", Description: "Description", Hot: true})

		// 重置测试上下文
		utt.ResetContext()
		// 配置 QueryString
		utt.GinTestContext.Request, _ = http.NewRequest("GET", "/?id=2", nil)

		// 测试 配置了搜索条件 ID [ 强匹配 ]
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		MiniProgramPhotoProcessingGetList(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "{\"Data\":[{\"ID\":2,\"CreatedAt\":\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\",\"CreatedUser\":\"112\",\"CreatedRole\":\"1102\",\"LastUpdatedAt\":\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\",\"LastUpdatedUser\":\"112\",\"LastUpdatedRole\":\"1102\",\"DeletedAt\":null,\"Operational\":false,\"Name\":\"名称2\",\"Project\":\"项目1\",\"CRMEventFormID\":2,\"CRMEventFormSID\":\"6936c8475486621c61ad6a9d0865ae7b\",\"MillimeterWidth\":1,\"MillimeterHeight\":1,\"PixelWidth\":1,\"PixelHeight\":1,\"BackgroundColors\":\"BackgroundColors\",\"Description\":\"Description\",\"Hot\":true}],\"Message\":\"Success\",\"Total\":1}")

		// 重置测试上下文
		utt.ResetContext()
		// 配置 QueryString
		utt.GinTestContext.Request, _ = http.NewRequest("GET", "/?name=名称", nil)

		// 测试 配置了搜索条件 名称 [ 模糊搜索 ]
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		MiniProgramPhotoProcessingGetList(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "{\"Data\":[{\"ID\":2,\"CreatedAt\":\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\",\"CreatedUser\":\"112\",\"CreatedRole\":\"1102\",\"LastUpdatedAt\":\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\",\"LastUpdatedUser\":\"112\",\"LastUpdatedRole\":\"1102\",\"DeletedAt\":null,\"Operational\":false,\"Name\":\"名称2\",\"Project\":\"项目1\",\"CRMEventFormID\":2,\"CRMEventFormSID\":\"6936c8475486621c61ad6a9d0865ae7b\",\"MillimeterWidth\":1,\"MillimeterHeight\":1,\"PixelWidth\":1,\"PixelHeight\":1,\"BackgroundColors\":\"BackgroundColors\",\"Description\":\"Description\",\"Hot\":true},{\"ID\":1,\"CreatedAt\":\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\",\"CreatedUser\":\"111\",\"CreatedRole\":\"1101\",\"LastUpdatedAt\":\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\",\"LastUpdatedUser\":\"111\",\"LastUpdatedRole\":\"1101\",\"DeletedAt\":null,\"Operational\":false,\"Name\":\"名称1\",\"Project\":\"项目1\",\"CRMEventFormID\":1,\"CRMEventFormSID\":\"6936c8475486621c61ad6a9d0865ae7a\",\"MillimeterWidth\":1,\"MillimeterHeight\":1,\"PixelWidth\":1,\"PixelHeight\":1,\"BackgroundColors\":\"BackgroundColors\",\"Description\":\"Description\",\"Hot\":false}],\"Message\":\"Success\",\"Total\":2}")

		// 重置测试上下文
		utt.ResetContext()
		// 配置 QueryString
		utt.GinTestContext.Request, _ = http.NewRequest("GET", "/?project=项目1", nil)

		// 测试 配置了搜索条件 项目 [ 强匹配 ]
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		MiniProgramPhotoProcessingGetList(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "{\"Data\":[{\"ID\":2,\"CreatedAt\":\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\",\"CreatedUser\":\"112\",\"CreatedRole\":\"1102\",\"LastUpdatedAt\":\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\",\"LastUpdatedUser\":\"112\",\"LastUpdatedRole\":\"1102\",\"DeletedAt\":null,\"Operational\":false,\"Name\":\"名称2\",\"Project\":\"项目1\",\"CRMEventFormID\":2,\"CRMEventFormSID\":\"6936c8475486621c61ad6a9d0865ae7b\",\"MillimeterWidth\":1,\"MillimeterHeight\":1,\"PixelWidth\":1,\"PixelHeight\":1,\"BackgroundColors\":\"BackgroundColors\",\"Description\":\"Description\",\"Hot\":true},{\"ID\":1,\"CreatedAt\":\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\",\"CreatedUser\":\"111\",\"CreatedRole\":\"1101\",\"LastUpdatedAt\":\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\",\"LastUpdatedUser\":\"111\",\"LastUpdatedRole\":\"1101\",\"DeletedAt\":null,\"Operational\":false,\"Name\":\"名称1\",\"Project\":\"项目1\",\"CRMEventFormID\":1,\"CRMEventFormSID\":\"6936c8475486621c61ad6a9d0865ae7a\",\"MillimeterWidth\":1,\"MillimeterHeight\":1,\"PixelWidth\":1,\"PixelHeight\":1,\"BackgroundColors\":\"BackgroundColors\",\"Description\":\"Description\",\"Hot\":false}],\"Message\":\"Success\",\"Total\":2}")

		// 重置测试上下文
		utt.ResetContext()
		// 配置 QueryString
		utt.GinTestContext.Request, _ = http.NewRequest("GET", "/?crm-event-form-id=2", nil)

		// 测试 配置了搜索条件 CRM 活动表单 ID [ 强匹配 ]
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		MiniProgramPhotoProcessingGetList(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "{\"Data\":[{\"ID\":2,\"CreatedAt\":\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\",\"CreatedUser\":\"112\",\"CreatedRole\":\"1102\",\"LastUpdatedAt\":\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\",\"LastUpdatedUser\":\"112\",\"LastUpdatedRole\":\"1102\",\"DeletedAt\":null,\"Operational\":false,\"Name\":\"名称2\",\"Project\":\"项目1\",\"CRMEventFormID\":2,\"CRMEventFormSID\":\"6936c8475486621c61ad6a9d0865ae7b\",\"MillimeterWidth\":1,\"MillimeterHeight\":1,\"PixelWidth\":1,\"PixelHeight\":1,\"BackgroundColors\":\"BackgroundColors\",\"Description\":\"Description\",\"Hot\":true}],\"Message\":\"Success\",\"Total\":1}")

		// 重置测试上下文
		utt.ResetContext()
		// 配置 QueryString
		utt.GinTestContext.Request, _ = http.NewRequest("GET", "/?crm-event-form-sid=6936c8475486621c61ad6a9d0865ae7", nil)

		// 测试 配置了搜索条件 CRM 活动表单 SID [ 强匹配 ]
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		MiniProgramPhotoProcessingGetList(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"CRM 活动表单 SID 格式不正确\"}")

		// 重置测试上下文
		utt.ResetContext()
		// 配置 QueryString
		utt.GinTestContext.Request, _ = http.NewRequest("GET", "/?crm-event-form-sid=6936c8475486621c61ad6a9d0865ae7b", nil)

		// 测试 配置了搜索条件 CRM 活动表单 SID [ 强匹配 ]
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		MiniProgramPhotoProcessingGetList(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "{\"Data\":[{\"ID\":2,\"CreatedAt\":\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\",\"CreatedUser\":\"112\",\"CreatedRole\":\"1102\",\"LastUpdatedAt\":\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\",\"LastUpdatedUser\":\"112\",\"LastUpdatedRole\":\"1102\",\"DeletedAt\":null,\"Operational\":false,\"Name\":\"名称2\",\"Project\":\"项目1\",\"CRMEventFormID\":2,\"CRMEventFormSID\":\"6936c8475486621c61ad6a9d0865ae7b\",\"MillimeterWidth\":1,\"MillimeterHeight\":1,\"PixelWidth\":1,\"PixelHeight\":1,\"BackgroundColors\":\"BackgroundColors\",\"Description\":\"Description\",\"Hot\":true}],\"Message\":\"Success\",\"Total\":1}")

		// 重置测试上下文
		utt.ResetContext()
		// 配置 QueryString
		utt.GinTestContext.Request, _ = http.NewRequest("GET", "/?hot=true", nil)

		// 测试 配置了搜索条件 是否热门 [ 强匹配 ]
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		MiniProgramPhotoProcessingGetList(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "{\"Data\":[{\"ID\":3,\"CreatedAt\":\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\",\"CreatedUser\":\"112\",\"CreatedRole\":\"1102\",\"LastUpdatedAt\":\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\",\"LastUpdatedUser\":\"112\",\"LastUpdatedRole\":\"1102\",\"DeletedAt\":null,\"Operational\":false,\"Name\":\"测试名1\",\"Project\":\"项目2\",\"CRMEventFormID\":3,\"CRMEventFormSID\":\"6936c8475486621c61ad6a9d0865ae7c\",\"MillimeterWidth\":1,\"MillimeterHeight\":1,\"PixelWidth\":1,\"PixelHeight\":1,\"BackgroundColors\":\"BackgroundColors\",\"Description\":\"Description\",\"Hot\":true},{\"ID\":2,\"CreatedAt\":\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\",\"CreatedUser\":\"112\",\"CreatedRole\":\"1102\",\"LastUpdatedAt\":\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\",\"LastUpdatedUser\":\"112\",\"LastUpdatedRole\":\"1102\",\"DeletedAt\":null,\"Operational\":false,\"Name\":\"名称2\",\"Project\":\"项目1\",\"CRMEventFormID\":2,\"CRMEventFormSID\":\"6936c8475486621c61ad6a9d0865ae7b\",\"MillimeterWidth\":1,\"MillimeterHeight\":1,\"PixelWidth\":1,\"PixelHeight\":1,\"BackgroundColors\":\"BackgroundColors\",\"Description\":\"Description\",\"Hot\":true}],\"Message\":\"Success\",\"Total\":2}")

		// 重置测试上下文
		utt.ResetContext()
		// 配置 QueryString
		utt.GinTestContext.Request, _ = http.NewRequest("GET", "/?created-user=1", nil)

		// 测试 配置了搜索条件 创建用户 条件三合一 ( ID [ 强匹配 ]; 工号 [ 模糊搜索 ]; 姓名 [ 模糊搜索 ] )
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		MiniProgramPhotoProcessingGetList(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "{\"Data\":[{\"ID\":3,\"CreatedAt\":\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\",\"CreatedUser\":\"112\",\"CreatedRole\":\"1102\",\"LastUpdatedAt\":\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\",\"LastUpdatedUser\":\"112\",\"LastUpdatedRole\":\"1102\",\"DeletedAt\":null,\"Operational\":false,\"Name\":\"测试名1\",\"Project\":\"项目2\",\"CRMEventFormID\":3,\"CRMEventFormSID\":\"6936c8475486621c61ad6a9d0865ae7c\",\"MillimeterWidth\":1,\"MillimeterHeight\":1,\"PixelWidth\":1,\"PixelHeight\":1,\"BackgroundColors\":\"BackgroundColors\",\"Description\":\"Description\",\"Hot\":true},{\"ID\":2,\"CreatedAt\":\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\",\"CreatedUser\":\"112\",\"CreatedRole\":\"1102\",\"LastUpdatedAt\":\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\",\"LastUpdatedUser\":\"112\",\"LastUpdatedRole\":\"1102\",\"DeletedAt\":null,\"Operational\":false,\"Name\":\"名称2\",\"Project\":\"项目1\",\"CRMEventFormID\":2,\"CRMEventFormSID\":\"6936c8475486621c61ad6a9d0865ae7b\",\"MillimeterWidth\":1,\"MillimeterHeight\":1,\"PixelWidth\":1,\"PixelHeight\":1,\"BackgroundColors\":\"BackgroundColors\",\"Description\":\"Description\",\"Hot\":true},{\"ID\":1,\"CreatedAt\":\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\",\"CreatedUser\":\"111\",\"CreatedRole\":\"1101\",\"LastUpdatedAt\":\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\",\"LastUpdatedUser\":\"111\",\"LastUpdatedRole\":\"1101\",\"DeletedAt\":null,\"Operational\":false,\"Name\":\"名称1\",\"Project\":\"项目1\",\"CRMEventFormID\":1,\"CRMEventFormSID\":\"6936c8475486621c61ad6a9d0865ae7a\",\"MillimeterWidth\":1,\"MillimeterHeight\":1,\"PixelWidth\":1,\"PixelHeight\":1,\"BackgroundColors\":\"BackgroundColors\",\"Description\":\"Description\",\"Hot\":false}],\"Message\":\"Success\",\"Total\":3}")

		// 重置测试上下文
		utt.ResetContext()
		// 配置 QueryString
		utt.GinTestContext.Request, _ = http.NewRequest("GET", "/?wrong-query-string=wrong-query-string", nil)

		// 测试 配置了搜索条件 搜索条件配置有误
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		MiniProgramPhotoProcessingGetList(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"搜索条件配置有误\"}")

		// 重置测试上下文
		utt.ResetContext()
		// 配置 QueryString
		utt.GinTestContext.Request, _ = http.NewRequest("GET", "/", nil)

		// 测试 没有配置搜索条件
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		MiniProgramPhotoProcessingGetList(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "{\"Data\":[{\"ID\":3,\"CreatedAt\":\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\",\"CreatedUser\":\"112\",\"CreatedRole\":\"1102\",\"LastUpdatedAt\":\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\",\"LastUpdatedUser\":\"112\",\"LastUpdatedRole\":\"1102\",\"DeletedAt\":null,\"Operational\":false,\"Name\":\"测试名1\",\"Project\":\"项目2\",\"CRMEventFormID\":3,\"CRMEventFormSID\":\"6936c8475486621c61ad6a9d0865ae7c\",\"MillimeterWidth\":1,\"MillimeterHeight\":1,\"PixelWidth\":1,\"PixelHeight\":1,\"BackgroundColors\":\"BackgroundColors\",\"Description\":\"Description\",\"Hot\":true},{\"ID\":2,\"CreatedAt\":\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\",\"CreatedUser\":\"112\",\"CreatedRole\":\"1102\",\"LastUpdatedAt\":\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\",\"LastUpdatedUser\":\"112\",\"LastUpdatedRole\":\"1102\",\"DeletedAt\":null,\"Operational\":false,\"Name\":\"名称2\",\"Project\":\"项目1\",\"CRMEventFormID\":2,\"CRMEventFormSID\":\"6936c8475486621c61ad6a9d0865ae7b\",\"MillimeterWidth\":1,\"MillimeterHeight\":1,\"PixelWidth\":1,\"PixelHeight\":1,\"BackgroundColors\":\"BackgroundColors\",\"Description\":\"Description\",\"Hot\":true},{\"ID\":1,\"CreatedAt\":\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\",\"CreatedUser\":\"111\",\"CreatedRole\":\"1101\",\"LastUpdatedAt\":\"")
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldContainSubstring, "\",\"LastUpdatedUser\":\"111\",\"LastUpdatedRole\":\"1101\",\"DeletedAt\":null,\"Operational\":false,\"Name\":\"名称1\",\"Project\":\"项目1\",\"CRMEventFormID\":1,\"CRMEventFormSID\":\"6936c8475486621c61ad6a9d0865ae7a\",\"MillimeterWidth\":1,\"MillimeterHeight\":1,\"PixelWidth\":1,\"PixelHeight\":1,\"BackgroundColors\":\"BackgroundColors\",\"Description\":\"Description\",\"Hot\":false}],\"Message\":\"Success\",\"Total\":3}")
	})
}
