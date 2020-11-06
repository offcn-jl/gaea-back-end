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

// UnitTestTool 单元测试工具
type UnitTestTool struct {
	ORM *gorm.DB
}

// tableList 单元测试需要使用的数据库表列表
var tableList = []interface{}{
	// system 系统表
	structs.SystemConfig{},
}

// CreatORM 单元测试工具 创建 ORM
func (t *UnitTestTool) CreatORM() {
	// 使用正确的 DSN 初始化 MYSQL 客户端
	t.ORM, _ = gorm.Open("mysql", os.Getenv("UNIT_TEST_MYSQL_DSN_GAEA"))
}

// InitORM 单元测试工具 初始化 ORM
func (t *UnitTestTool) InitORM() {
	for _, table := range tableList {
		t.ORM.AutoMigrate(table)
	}
}

// CloseORM 单元测试工具 关闭 ORM
func (t *UnitTestTool) CloseORM() {
	for _, table := range tableList {
		t.ORM.DropTableIfExists(table)
	}

	// 关闭 ORM 的连接
	t.ORM.Close()
}
