/*
   @Time : 2020/10/31 3:13 下午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : config
   @Software: GoLand
*/

package config

import (
	"github.com/jinzhu/gorm"
	"github.com/offcn-jl/gaea-back-end/commons/database/structs"
)

// Version 版本号
// 用于 ORM 模块判断是否需要进行数据库结构自动迁移
// 用于输入版本信息到响应头
// 设置为距离当前时间较为久远的项目创始时间，避免每次调试时都需要自动同步数据库 可以通过编译的方式指定版本号：go build -ldflags "-X main.VERSION=x.x.x"
var Version = "0000000 [ 2020/10/31 09:46:00 ]"

var currentConfig structs.SystemConfig

// Init 初始化
func Init(orm *gorm.DB) error {
	// 从数据库中取出最后一条配置作为当前配置
	return orm.Last(&currentConfig).Error
}

// Get 获取配置 fixme 测试
// 供系统内部调用 , 获取当前配置, 不对外直接暴露配置变量
func Get() structs.SystemConfig {
	return currentConfig
}

// Update 修改配置 fixme 测试
// 供系统内部调用 , 更新当前配置, 不对外直接暴露配置变量
func Update(orm *gorm.DB, newConfig structs.SystemConfig) {
	// 向数据库中添加新的配置
	orm.Create(&newConfig)

	// 更新内存中的配置
	currentConfig = newConfig
}
