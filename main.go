/*
   @Time : 2020/10/31 9:55 上午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : main
   @Software: GoLand
   @Description: 入口函数及主程序初始化逻辑
*/

package main

import (
	"github.com/offcn-jl/gaea-back-end/commons/config"
	"github.com/offcn-jl/gaea-back-end/commons/database/orm"
	"github.com/offcn-jl/gaea-back-end/commons/logger"
	"github.com/offcn-jl/gaea-back-end/commons/router"
	"os"
)

// init 初始化操作
func init() {
	defer func() {
		// 捕获 PANIC
		if err := recover(); err != nil {
			logger.Log("捕获到了 PANIC 产生的异常. 未定义任何处理逻辑, 主进程已结束.")
			// 使用 defer 配合空 select 阻塞进程退出
			// 防止进程退出后, 容器被清理, 无法进行调试
			select {}
		}
	}()

	// 初始化数据库
	if err := orm.Init(); err != nil {
		// 初始化失败
		logger.Log("初始化数据库失败.")
		// 抛出异常
		logger.Panic(err)
	}

	// 初始化系统配置
	if err := config.Init(orm.MySQL.Gaea); err != nil {
		// 初始化失败
		logger.Log("初始化系统配置失败.")
		// 抛出异常
		logger.Panic(err)
	}

	// 打印初始化完成信息及版本号到控制台
	logger.Log("Gaea 项目后端初始化完成, 当前版本 : " + config.Version)
}

// main 入口
func main() {
	// 在入口程序结束时关闭数据库连接
	defer func() { orm.MySQL.Gaea.Close() }()

	// 定义基础路径
	mode := "/release"
	// 判断运行环境, 如果不是 release 则修改基础路径为 /test
	if os.Getenv("GIN_MODE") != "release" {
		mode = "/test"
	}

	// 初始化路由后启动监听
	if err := router.InitRouter(mode).Run(); err != nil {
		// 处理错误
		logger.Panic(err)
	}
}
