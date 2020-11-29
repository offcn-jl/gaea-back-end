/*
   @Time : 2020/11/5 8:19 上午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : orm_test
   @Software: GoLand
   @Description: ORM 组件 单元测试
*/

package orm

import (
	"database/sql/driver"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/offcn-jl/gaea-back-end/commons/config"
	"github.com/offcn-jl/gaea-back-end/commons/database/structs"
	"github.com/offcn-jl/gaea-back-end/commons/utt"
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
	"time"
)

// 实现 driver.Valuer 接口 供 TestCustomGormLogFormatter 使用
type testVal struct{ val interface{} }

func (u testVal) Value() (driver.Value, error) { return u.val, nil }

// TestCustomGormLogFormatter 测试自定义的 Gorm LogFormatter
func TestCustomGormLogFormatter(t *testing.T) {
	Convey("测试自定义的 Gorm LogFormatter", t, func() {
		// 测试 普通日志
		logs := gorm.LogFormatter("日志级别", "日志内容")
		So(fmt.Sprint(logs), ShouldContainSubstring, "日志内容")

		// 测试 ? 形式的构造参数
		logs = gorm.LogFormatter("sql", "调用方源码路径及代码的行数", time.Now().Sub(time.Now().Add(-1*time.Second)), "参数 ?" /* 带有 ? SQL 语句 */, []interface{}{"参数1"} /* SQL 语句的构造参数 */, int64(0) /* 受影响的行数 */)
		So(fmt.Sprint(logs), ShouldContainSubstring, "参数 '参数1'")

		// 测试 $n 形式的构造参数 ( n 从 1 开始 $0 不会进行替换 )
		logs = gorm.LogFormatter("sql", "调用方源码路径及代码的行数", time.Now().Sub(time.Now().Add(-1*time.Second)), "参数 $0 $1" /* 带有 $n 的 SQL 语句 */, []interface{}{"参数1", "参数2"} /* SQL 语句的构造参数 */, int64(0) /* 受影响的行数 */)
		So(fmt.Sprint(logs), ShouldContainSubstring, "参数 $0 '参数1'")

		// 测试 非法 构造参数
		logs = gorm.LogFormatter("sql", "调用方源码路径及代码的行数", time.Now().Sub(time.Now().Add(-1*time.Second)), "", []interface{}{nil} /* SQL 语句的构造参数 */, int64(0) /* 受影响的行数 */)
		So(fmt.Sprint(logs), ShouldContainSubstring, "NULL")

		// 测试 各种类型的构造参数
		logs = gorm.LogFormatter("sql", "调用方源码路径及代码的行数", time.Now().Sub(time.Now().Add(-1*time.Second)), "参数 ? ? ? ? ? ? ? ? ?" /* 带有 ? SQL 语句 */, []interface{}{time.Time{}, time.Unix(1604741402, 0).In(time.FixedZone("CST", 8*3600)), []byte("Byte 文字参数"), make([]byte, 10), testVal{"字符串 类型 driver.Value"}, testVal{nil}, 0} /* SQL 语句的构造参数 */, int64(0) /* 受影响的行数 */)
		So(fmt.Sprint(logs), ShouldContainSubstring, " 参数 '0000-00-00 00:00:00' '2020-11-07 17:30:02' 'Byte 文字参数' '<binary>' '字符串 类型 driver.Value' NULL 0")
	})
}

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
	utt.ORM = MySQL.Gaea
	utt.CloseORM()
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
	utt.ORM = MySQL.Gaea
	utt.CloseORM()
}
