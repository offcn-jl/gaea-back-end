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
	"net/http"
	"strings"
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

	// 检查 CORS 并在通过检查后放行所有 OPTIONS 请求
	router.Use(func(c *gin.Context) {
		// 跳过 /favicon.ico
		if c.Request.URL.Path == "/favicon.ico" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		// 不是标准路径
		if len(strings.Split(c.Request.URL.Path, "/")) < 3 {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"Msg": "路径有误"})
			return
		}

		// 匹配检查规则
		allowOrigins := ""
		switch strings.Split(c.Request.URL.Path, "/")[2] {
		case "services":
			allowOrigins = config.Get().CORSRuleServices
		case "manages":
			allowOrigins = config.Get().CORSRuleManages
		case "events":
			allowOrigins = config.Get().CORSRuleEvents
		default:
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"Msg": "路径有误"})
			return
		}

		// 跨域检查
		// 作用仅用来防止浏览器端的非法调用，所以不严格校验未包含 origin 头的情况
		allowOriginsArray := strings.Split(allowOrigins, ",")
		pass := false
		for _, origin := range allowOriginsArray {
			// 遍历配置中的跨域头，寻找匹配项
			if c.GetHeader("origin") == origin {
				c.Header("Access-Control-Allow-Origin", origin)
				pass = true
				// 只要有一个跨域头匹配就跳出循环
				break
			}
		}

		// 未通过检查
		if !pass {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"Msg": "请求未通过跨域检查"})
			return
		}

		// 通过跨域检查后，放行所有 OPTIONS 方法，并添加按照客户端的请求添加 Allow Headers
		if c.Request.Method == "OPTIONS" {
			// 请求首部  Access-Control-Request-Headers 出现于 preflight request （预检请求）中，用于通知服务器在真正的请求中会采用哪些请求首部。
			c.Header("Access-Control-Allow-Headers", c.GetHeader("Access-Control-Request-Headers")) // 放行预检请求通知的请求首部。
			// https://cloud.tencent.com/developer/section/1189896
			c.Header("Access-Control-Allow-Methods", c.GetHeader("Access-Control-Request-Method")) // 放行预检请求通知的请求首部。
			c.AbortWithStatus(http.StatusNoContent)
		}
	})

	// 默认路由组
	defaultGroup := router.Group(basePath)

	// 内部服务路由组
	// 用于对接第三方平台、为内部的工具提供服务等
	servicesGroup := defaultGroup.Group("/services")
	{
		servicesGroup.GET("")
	}

	// 管理平台路由组
	// 用于管理平台
	managesGroup := defaultGroup.Group("/manages")
	{
		managesGroup.GET("")
	}

	// 活动 ( 外部服务 ) 路由组
	// 用于专题页、为专题页服务的模块等
	eventsGroup := defaultGroup.Group("/events")
	{
		eventsGroup.GET("")
	}

	// 未匹配到路由的路径返回统一的 404 响应
	router.Use(func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"Msg": "路径有误"})
	})

	return router
}
