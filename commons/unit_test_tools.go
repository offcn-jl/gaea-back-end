/*
   @Time : 2020/11/5 11:26 上午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : unit_test_tools
   @Software: GoLand
*/

package commons

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/offcn-jl/gaea-back-end/commons/database/structs"
	"os"
)

// tableList 单元测试需要使用的数据库表列表
var tableList = []interface{}{
	// system 系统表
	&structs.SystemConfig{},
}

// TestToolCreatORM 单元测试工具 创建 ORM
func TestToolCreatORM() *gorm.DB {
	// 使用正确的 DSN 初始化 MYSQL 客户端
	orm, _ := gorm.Open("mysql", os.Getenv("UNIT_TEST_MYSQL_DSN_GAEA"))
	return orm
}

// TestToolInitORM 单元测试工具 初始化 ORM
func TestToolInitORM(orm *gorm.DB) {
	for _, table := range tableList {
		orm.AutoMigrate(table)
	}
}

// TestToolRestORM 单元测试工具 重置 ORM
func TestToolRestORM(orm *gorm.DB) {
	for _, table := range tableList {
		orm.DropTableIfExists(table)
	}

	// 关闭 ORM 的连接
	orm.Close()
}
