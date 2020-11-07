/*
   @Time : 2020/11/6 4:18 下午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : unit_test_tools_test
   @Software: GoLand
*/

package commons

import (
	. "github.com/smartystreets/goconvey/convey"
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
