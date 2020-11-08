/*
   @Time : 2020/11/6 4:18 下午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : unit_test_tools_test
   @Software: GoLand
   @Description: 单元测试工具的单元测试
*/

package commons

import (
	"github.com/gin-gonic/gin"
	"github.com/offcn-jl/gaea-back-end/commons/database/structs"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"testing"
)

var unitTestTool UnitTestTool

// TestCreatORM 测试 CreatORM 函数是否可以创建 ORM
func TestCreatORM(t *testing.T) {
	Convey("测试 CreatORM 函数是否可以创建 ORM", t, func() {
		So(func() { unitTestTool.CreatORM() }, ShouldNotPanic)
	})
}

// TestInitORM 测试 InitORM 函数是否可以初始化 ORM
func TestInitORM(t *testing.T) {
	Convey("测试 InitORM 函数是否可以初始化 ORM", t, func() {
		So(func() { unitTestTool.InitORM() }, ShouldNotPanic)
	})
}

// TestCloseORM 测试 CloseORM 函数是否可以关闭 ORM
func TestCloseORM(t *testing.T) {
	Convey("测试 CloseORM 函数是否可以关闭 ORM", t, func() {
		So(func() { unitTestTool.CloseORM() }, ShouldNotPanic)
	})
}

// TestInitTest 测试 InitTest 函数是否可以初始化测试数据并创建测试上下文
func TestInitTest(t *testing.T) {
	Convey("测试 InitTest 函数是否可以初始化测试数据并创建测试上下文", t, func() {
		// 检查上下文是否为空
		So(unitTestTool.HttpTestResponseRecorder, ShouldBeNil)
		So(unitTestTool.GinTestContext, ShouldBeNil)

		// 初始化测试数据并创建测试上下文
		unitTestTool = InitTest()

		// 检查上下文是否不为空
		So(unitTestTool.HttpTestResponseRecorder, ShouldNotBeNil)
		So(unitTestTool.GinTestContext, ShouldNotBeNil)

		// 检查是否创建了数据
		singleSignOnOrganization := structs.SingleSignOnOrganization{}
		unitTestTool.ORM.First(&singleSignOnOrganization)
		So(singleSignOnOrganization.ID, ShouldEqual, 1)
		So(singleSignOnOrganization.Code, ShouldEqual, 22)
		So(singleSignOnOrganization.Name, ShouldEqual, "吉林分校")
	})
}

// TestResetContext 测试 ResetContext 函数是否可以重置测试上下文
func TestResetContext(t *testing.T) {
	Convey("测试 ResetContext 函数是否可以重置测试上下文", t, func() {
		// 初始化上下文 ( 对为初始化上下文的对象使用, 可以实现初始化上下文的效果 )
		unitTestTool.ResetContext()
		So(unitTestTool.HttpTestResponseRecorder.Code, ShouldEqual, http.StatusOK)
		So(unitTestTool.GinTestContext.Param("Test"), ShouldBeEmpty)
		// 修改上下文
		unitTestTool.HttpTestResponseRecorder.Code = http.StatusInternalServerError
		unitTestTool.GinTestContext.Params = gin.Params{gin.Param{Key: "Test", Value: "Value"}}
		So(unitTestTool.HttpTestResponseRecorder.Code, ShouldEqual, http.StatusInternalServerError)
		So(unitTestTool.GinTestContext.Param("Test"), ShouldEqual, "Value")
		// 重置上下文
		unitTestTool.ResetContext()
		So(unitTestTool.HttpTestResponseRecorder.Code, ShouldEqual, http.StatusOK)
		So(unitTestTool.GinTestContext.Param("Test"), ShouldBeEmpty)
	})
}

// TestGetFakePhone 测试 GetFakePhone 函数是否可以获取测试用的非重复假号码
func TestGetFakePhone(t *testing.T) {
	Convey("测试 GetFakePhone 函数是否可以获取测试用的非重复假号码", t, func() {
		So(unitTestTool.FakePhoneCount, ShouldEqual, 0)
		fakePhone := unitTestTool.GetFakePhone()
		So(unitTestTool.FakePhoneCount, ShouldEqual, 1)
		So(fakePhone, ShouldNotEqual, unitTestTool.GetFakePhone())
		So(unitTestTool.FakePhoneCount, ShouldEqual, 2)
	})
}
