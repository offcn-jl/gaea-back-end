/*
   @Time : 2020/11/2 9:11 上午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : orm
   @Software: GoLand
*/

package orm

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/offcn-jl/gaea-back-end/commons/config"
	"github.com/offcn-jl/gaea-back-end/commons/database/structs"
	"github.com/offcn-jl/gaea-back-end/commons/logger"
	"os"
	"time"
)

// MySQL 对外暴露的数据库实例
var MySQL struct {
	Gaea *gorm.DB
}

// Init 初始化数据库
func Init() error {
	// 校验是否配置 MYSQL_DSN
	if os.Getenv("MYSQL_DSN_GAEA") == "" {
		// 未配置，返回错误
		return errors.New("未配置 MYSQL_DSN_GAEA , 请在环境变量中配置 MYSQL_DSN_GAEA ( 例 : user:password@tcp(hostname)/database?charset=utf8mb4&parseTime=True&loc=Local )")
	} else {
		// 在调试模式时，输出环境变量中配置的 MYSQL_DSN_GAEA
		logger.DebugToString("os.Getenv(\"MYSQL_DSN_GAEA\")", os.Getenv("MYSQL_DSN_GAEA"))
		// 已配置, 初始化 MYSQL 客户端
		var err error
		if MySQL.Gaea, err = gorm.Open("mysql", os.Getenv("MYSQL_DSN_GAEA")); err != nil {
			return err
		} else {
			// 判断初始化是否成功
			if MySQL.Gaea.Error != nil {
				return MySQL.Gaea.Error
			} else {
				// 初始化成功后，判断是否需要进行数据库表结构迁移
				if len(config.Version) != 31 {
					logger.Log("应用版本号有误, 无法计算构建时间, 开始进行数据库结构自动迁移.")
					autoMigrate()
				} else {
					// 通过版本号, 计算构建时间
					builtTime, _ := time.ParseInLocation("2006/01/02 15:04:05", config.Version[10:29], time.Local) // 使用 parseInLocation 将字符串格式化返回本地时区时间
					// 判断构建时间是否超过一小时
					if time.Since(builtTime).Hours() > 1 {
						// 超过一小时, 不需要进行迁移
						logger.Log("应用构建于 " + fmt.Sprint(time.Since(builtTime)) + " 前, 跳过数据库结构自动迁移.")
					} else {
						// 没有超过一小时, 进行表结构自动迁移
						logger.Log("应用构建时间没有超过一小时, 开始进行数据库结构自动迁移.")
						autoMigrate()
					}
				}
			}
			return nil
		}
	}
}

func autoMigrate() {
	MySQL.Gaea.AutoMigrate(
		// system 系统表
		&structs.SystemConfig{},
	)
}
