/*
   @Time : 2020/11/5 8:19 上午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : orm_test
   @Software: GoLand
*/

package orm

import (
	"github.com/jinzhu/gorm"
	"github.com/offcn-jl/gaea-back-end/commons"
	"github.com/offcn-jl/gaea-back-end/commons/config"
	"github.com/offcn-jl/gaea-back-end/commons/database/structs"
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
	"time"
)

// Test_autoMigrate 测试 autoMigrate 函数是否可以完成表结构自动迁移
func Test_autoMigrate(t *testing.T) {
	Convey("测试 autoMigrate 函数是否可以完成表结构自动迁移", t, func() {
		Convey("测试 未初始化 MYSQL 客户端 产生 PANIC [ runtime error: invalid memory address or nil pointer dereference ]", func() {
			So(autoMigrate, ShouldPanic)
		})

		// 使用错误的 DSN 初始化 MYSQL 客户端
		var err error
		MySQL.Gaea, err = gorm.Open("mysql", os.Getenv("MYSQL_DSN_GAEA"))
		// 开发环境与 Github Actions 环境的端口不同 ( Github Actions 的 MySQL 确实在监听 tcp 127.0.0.1:3306, 本条测试存在歧义，所以跳过 )
		//Convey("测试 使用错误的 DSN 初始化 MYSQL 客户端 时 产生 错误 [ dial tcp 127.0.0.1:3306: connect: connection refused ]", func() {
		//	So(err, ShouldBeError, "dial tcp 127.0.0.1:3306: connect: connection refused")
		//})
		Convey("测试 使用错误的 DSN 初始化 MYSQL 客户端 时 客户端未包含任何错误", func() {
			So(MySQL.Gaea.Error, ShouldBeEmpty)
		})
		Convey("测试 使用错误的 DSN 初始化 MYSQL 客户端 后, 运行自动迁移 产生 PANIC [ sql: database is closed ]", func() {
			So(autoMigrate, ShouldPanic)
		})

		// 使用正确的 DSN 初始化 MYSQL 客户端
		MySQL.Gaea, err = gorm.Open("mysql", os.Getenv("UNIT_TEST_MYSQL_DSN_GAEA"))

		Convey("使用错误的 DSN 初始化 MYSQL 客户端 并 自动迁移 后 系统表不存在", func() {
			So(MySQL.Gaea.HasTable(&structs.SystemConfig{}), ShouldBeFalse)
		})
		Convey("测试 使用正确的 DSN 初始化 MYSQL 客户端 时 不会产生 错误", func() {
			So(err, ShouldBeEmpty)
		})
		Convey("测试 使用正确的 DSN 初始化 MYSQL 客户端 时 客户端未包含任何错误", func() {
			So(MySQL.Gaea.Error, ShouldBeEmpty)
		})
		Convey("测试 使用正确的 DSN 初始化 MYSQL 客户端 后, 运行自动迁移 不会产生 PANIC", func() {
			So(autoMigrate, ShouldNotPanic)
		})
		Convey("使用正确的 DSN 初始化 MYSQL 客户端 并 自动迁移 后 系统表存在", func() {
			So(MySQL.Gaea.HasTable(&structs.SystemConfig{}), ShouldBeTrue)
		})
	})

	// 在程序结束时重置数据库
	commons.TestToolRestORM(MySQL.Gaea)
}

// TestInit 测试 Init 函数是否能够完成初始化数据库
func TestInit(t *testing.T) {
	Convey("测试 Init 函数是否能够完成初始化数据库", t, func() {
		Convey("测试 未配置 MYSQL_DSN_GAEA 时 返回错误", func() {
			So(Init(), ShouldBeError)
		})

		Convey("测试 配置 错误的 MYSQL_DSN_GAEA 时 返回错误", func() {
			os.Setenv("MYSQL_DSN_GAEA", "INVALID_DSN")
			So(Init(), ShouldBeError, "invalid DSN: missing the slash separating the database name")
		})

		// 配置 正确的 MYSQL_DSN_GAEA
		os.Setenv("MYSQL_DSN_GAEA", os.Getenv("UNIT_TEST_MYSQL_DSN_GAEA"))

		Convey("测试 版本号配置有误", func() {
			rightVersion := config.Version
			config.Version = ""
			So(Init(), ShouldBeEmpty)
			config.Version = rightVersion
		})

		Convey("测试 构建时间超过一个小时", func() {
			So(Init(), ShouldBeEmpty)
		})

		Convey("测试 构建时间未超过一个小时", func() {
			rightVersion := config.Version
			config.Version = time.Now().Format("0000000 [ 2006/01/02 15:04:05 ]")
			So(Init(), ShouldBeEmpty)
			config.Version = rightVersion
		})

		Convey("测试 初始化完成后 是否可以执行查询", func() {
			So(MySQL.Gaea.HasTable(&structs.SystemConfig{}), ShouldBeTrue)
		})
	})

	// 在程序结束时重置数据库
	commons.TestToolRestORM(MySQL.Gaea)
}
