/*
  @Time : 2020/11/5 11:47 上午
  @Author : ShadowWalker
  @Email : master@rebeta.cn
  @File : config_test
  @Software: GoLand
@Description: 配置组件 单元测试
*/

package config

import (
	"github.com/offcn-jl/gaea-back-end/commons/database/structs"
	"github.com/offcn-jl/gaea-back-end/commons/utt"
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
)

// TestInit 测试 Init 函数是否可以完成初始化配置
func TestInit(t *testing.T) {
	Convey("测试 初始化配置", t, func() {
		Convey("测试 未初始化 ORM 时 抛出 PANIC [ runtime error: invalid memory address or nil pointer dereference ]", func() {
			So(func() { Init(utt.ORM) }, ShouldPanic)
		})
		Convey("测试 DSN 配置有误 时 抛出 PANIC [ runtime error: invalid memory address or nil pointer dereference ]", func() {
			rightDSN := os.Getenv("UNIT_TEST_MYSQL_DSN_GAEA")
			os.Setenv("UNIT_TEST_MYSQL_DSN_GAEA", "INVALID_DSN")
			utt.CreatORM()
			So(func() { Init(utt.ORM) }, ShouldPanic)
			os.Setenv("UNIT_TEST_MYSQL_DSN_GAEA", rightDSN)
		})

		Convey("测试 不存在记录 时 返回错误 [ record not found ]", func() {
			// 使用正确的 DSN 创建 ORM
			utt.CreatORM()
			// 初始化 ORM
			utt.InitORM()
			So(Init(utt.ORM), ShouldBeError, "record not found")
		})

		Convey("测试 存在记录 时 返回值为空并且成功取出配置", func() {
			// 创建一条记录
			utt.ORM.Create(&structs.SystemConfig{DisableDebug: true})
			So(Init(utt.ORM), ShouldBeEmpty)
			So(currentConfig.DisableDebug, ShouldBeTrue)
		})
	})

	// 恢复配置状态
	currentConfig = structs.SystemConfig{}

	// 在程序结束时重置数据库
	utt.CloseORM()
}

// TestGet 测试 Get 函数是否可以按照预期获取配置
func TestGet(t *testing.T) {
	Convey("测试 Get 函数是否可以按照预期获取配置", t, func() {
		Convey("测试 未初始化配置 ( 或初始化失败 ) 时, 获取到的配置为默认配置", func() {
			So(Get().DisableDebug, ShouldBeFalse)
		})

		Convey("测试 初始化配置后 获取到的时数据库中的最后一条配置", func() {
			// 使用正确的 DSN 创建 ORM
			utt.CreatORM()
			// 初始化 ORM
			utt.InitORM()
			// 创建一条记录
			utt.ORM.Create(&structs.SystemConfig{DisableDebug: true})
			// 初始化配置
			So(Init(utt.ORM), ShouldBeEmpty)
			// 获取配置
			So(Get().DisableDebug, ShouldBeTrue)
			// 在程序结束时重置数据库
			utt.CloseORM()
		})
	})
}

// TestUpdate 测试 Update 函数是否可以完成修改配置
func TestUpdate(t *testing.T) {
	Convey("测试 Update 函数是否可以完成修改配置", t, func() {
		// 使用正确的 DSN 创建 ORM
		utt.CreatORM()
		// 初始化 ORM
		utt.InitORM()
		// 创建一条记录
		utt.ORM.Create(&structs.SystemConfig{DisableDebug: true})
		// 初始化配置
		So(Init(utt.ORM), ShouldBeEmpty)
		// 获取配置
		So(Get().DisableDebug, ShouldBeTrue)
		// 修改配置
		Update(utt.ORM, structs.SystemConfig{DisableDebug: false})
		// 获取配置
		So(Get().DisableDebug, ShouldBeFalse)
		// 在程序结束时重置数据库
		utt.CloseORM()
	})
}
