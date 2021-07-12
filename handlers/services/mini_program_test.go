/*
   @Time : 2021/7/1 10:51 上午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : mini_program_test.go
   @Package : services
   @Description: 单元测试 小程序内部接口
*/

package services

import (
	"github.com/gin-gonic/gin"
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

// 创建测试数据
func createTestData() {
	orm.MySQL.Gaea.DropTableIfExists(&structs.MiniProgramPhotoProcessingConfig{})
	orm.MySQL.Gaea.AutoMigrate(&structs.MiniProgramPhotoProcessingConfig{})

	orm.MySQL.Gaea.Create(&structs.MiniProgramPhotoProcessingConfig{Name: "测试11", Project: "项目1", CRMEventFormID: 72270, CRMEventFormSID: "6936c8475486621c61ad6a9d0865ae7a", MillimeterWidth: 50, MillimeterHeight: 50, PixelWidth: 50, PixelHeight: 50, BackgroundColors: "[]", Description: "备注", Hot: false})
	orm.MySQL.Gaea.Create(&structs.MiniProgramPhotoProcessingConfig{Name: "测试12", Project: "项目1", CRMEventFormID: 72270, CRMEventFormSID: "6936c8475486621c61ad6a9d0865ae7a", MillimeterWidth: 60, MillimeterHeight: 60, PixelWidth: 60, PixelHeight: 60, BackgroundColors: "[]", Description: "备注", Hot: false})
	orm.MySQL.Gaea.Create(&structs.MiniProgramPhotoProcessingConfig{Name: "测试13", Project: "项目1", CRMEventFormID: 72270, CRMEventFormSID: "6936c8475486621c61ad6a9d0865ae7a", MillimeterWidth: 20, MillimeterHeight: 70, PixelWidth: 70, PixelHeight: 70, BackgroundColors: "[]", Description: "备注", Hot: true})
	orm.MySQL.Gaea.Create(&structs.MiniProgramPhotoProcessingConfig{Name: "测试21", Project: "项目2", CRMEventFormID: 72270, CRMEventFormSID: "6936c8475486621c61ad6a9d0865ae7a", MillimeterWidth: 50, MillimeterHeight: 50, PixelWidth: 50, PixelHeight: 50, BackgroundColors: "[]", Description: "备注", Hot: false})
	orm.MySQL.Gaea.Create(&structs.MiniProgramPhotoProcessingConfig{Name: "测试22", Project: "项目2", CRMEventFormID: 72270, CRMEventFormSID: "6936c8475486621c61ad6a9d0865ae7a", MillimeterWidth: 60, MillimeterHeight: 60, PixelWidth: 60, PixelHeight: 60, BackgroundColors: "[]", Description: "备注", Hot: true})
}

// TestMiniProgramPhotoProcessingConfigList 测试 MiniProgramPhotoProcessingConfigList 是否可以按预期按照查询参数获取照片处理配置列表
func TestMiniProgramPhotoProcessingConfigList(t *testing.T) {
	Convey("测试 MiniProgramPhotoProcessingConfigList 是否可以按预期按照查询参数获取照片处理配置列表", t, func() {
		// 创建测试数据
		createTestData()

		// 测试 未配置条件 强匹配 热门
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		MiniProgramPhotoProcessingConfigList(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Data\":[{\"ID\":5,\"Name\":\"测试22\",\"MillimeterWidth\":60,\"MillimeterHeight\":60,\"PixelWidth\":60,\"PixelHeight\":60},{\"ID\":3,\"Name\":\"测试13\",\"MillimeterWidth\":20,\"MillimeterHeight\":70,\"PixelWidth\":70,\"PixelHeight\":70}],\"Message\":\"Success\",\"Total\":2}")

		// 重置测试上下文
		utt.ResetContext()
		// 配置 QueryString
		utt.GinTestContext.Request, _ = http.NewRequest("GET", "/?search=2", nil)

		// 测试 search 模糊搜索 名称 尺寸
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		MiniProgramPhotoProcessingConfigList(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Data\":[{\"ID\":5,\"Name\":\"测试22\",\"MillimeterWidth\":60,\"MillimeterHeight\":60,\"PixelWidth\":60,\"PixelHeight\":60},{\"ID\":4,\"Name\":\"测试21\",\"MillimeterWidth\":50,\"MillimeterHeight\":50,\"PixelWidth\":50,\"PixelHeight\":50},{\"ID\":3,\"Name\":\"测试13\",\"MillimeterWidth\":20,\"MillimeterHeight\":70,\"PixelWidth\":70,\"PixelHeight\":70},{\"ID\":2,\"Name\":\"测试12\",\"MillimeterWidth\":60,\"MillimeterHeight\":60,\"PixelWidth\":60,\"PixelHeight\":60}],\"Message\":\"Success\",\"Total\":4}")

		// 重置测试上下文
		utt.ResetContext()
		// 配置 QueryString
		utt.GinTestContext.Request, _ = http.NewRequest("GET", "/?project=项目2", nil)

		// 测试 project 强匹配 项目
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		MiniProgramPhotoProcessingConfigList(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Data\":[{\"ID\":5,\"Name\":\"测试22\",\"MillimeterWidth\":60,\"MillimeterHeight\":60,\"PixelWidth\":60,\"PixelHeight\":60},{\"ID\":4,\"Name\":\"测试21\",\"MillimeterWidth\":50,\"MillimeterHeight\":50,\"PixelWidth\":50,\"PixelHeight\":50}],\"Message\":\"Success\",\"Total\":2}")

		// 重置测试上下文
		utt.ResetContext()
		// 配置 QueryString
		utt.GinTestContext.Request, _ = http.NewRequest("GET", "/?search=2&page=2&limit=1", nil)

		// 测试 分页
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		MiniProgramPhotoProcessingConfigList(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Data\":[{\"ID\":4,\"Name\":\"测试21\",\"MillimeterWidth\":50,\"MillimeterHeight\":50,\"PixelWidth\":50,\"PixelHeight\":50}],\"Message\":\"Success\",\"Total\":4}")

		// 重置测试上下文
		utt.ResetContext()
		// 配置 QueryString
		utt.GinTestContext.Request, _ = http.NewRequest("GET", "/?search=2&page=3&limit=1", nil)

		// 测试 分页
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		MiniProgramPhotoProcessingConfigList(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Data\":[{\"ID\":3,\"Name\":\"测试13\",\"MillimeterWidth\":20,\"MillimeterHeight\":70,\"PixelWidth\":70,\"PixelHeight\":70}],\"Message\":\"Success\",\"Total\":4}")
	})
}

// TestMiniProgramPhotoProcessingConfig 测试 MiniProgramPhotoProcessingConfig 是否可以按照预期获取照片处理配置
func TestMiniProgramPhotoProcessingConfig(t *testing.T) {
	Convey("测试 MiniProgramPhotoProcessingConfig 是否可以按照预期获取照片处理配置", t, func() {
		// 创建测试数据
		createTestData()

		// 测试 未配置 ID
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		MiniProgramPhotoProcessingConfig(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"ID 为  的照片处理不存在\"}")

		// 配置 Path 中的请求参数
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "ID", Value: "99"}}

		// 测试 配置的 ID 不存在
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		MiniProgramPhotoProcessingConfig(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Message\":\"ID 为 99 的照片处理不存在\"}")

		// 配置 Path 中的请求参数
		utt.GinTestContext.Params = gin.Params{gin.Param{Key: "ID", Value: "1"}}

		// 测试 获取成功
		utt.HttpTestResponseRecorder.Body.Reset() // 测试前重置 body
		MiniProgramPhotoProcessingConfig(utt.GinTestContext)
		So(utt.HttpTestResponseRecorder.Body.String(), ShouldEqual, "{\"Data\":{\"Name\":\"测试11\",\"Project\":\"项目1\",\"CRMEventFormID\":72270,\"CRMEventFormSID\":\"6936c8475486621c61ad6a9d0865ae7a\",\"MillimeterWidth\":50,\"MillimeterHeight\":50,\"PixelWidth\":50,\"PixelHeight\":50,\"BackgroundColors\":\"[]\",\"Description\":\"备注\"},\"Message\":\"Success\"}")
	})
}
