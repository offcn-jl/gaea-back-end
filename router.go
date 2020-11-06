/*
   @Time : 2020/10/31 9:59 上午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : router
   @Software: GoLand
*/

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/offcn-jl/gaea-back-end/commons/config"
)

// initRouter 初始化路由
func initRouter(basePath string) *gin.Engine {
	// 关闭 Gin 的控制台彩色输出
	gin.DisableConsoleColor()

	// 使用默认配置初始化路由
	router := gin.Default()

	// 添加版本号
	router.Use(func(c *gin.Context) {
		c.Header("Server", "Gaea - "+config.Version)
	})

	// 默认路由组
	defaultGroup := router.Group(basePath)

	// 内部服务路由组
	servicesGroup := defaultGroup.Group("/services")
	{
		servicesGroup.GET("")
	}

	// 管理平台路由组
	managesGroup := defaultGroup.Group("/manages")
	{
		managesGroup.GET("")
	}

	// 活动 ( 外部服务 ) 路由组
	eventsGroup := defaultGroup.Group("/events")
	{
		eventsGroup.GET("")
	}

	return router
}
