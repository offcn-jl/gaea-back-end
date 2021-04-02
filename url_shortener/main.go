/*
   @Time : 2021/4/1 7:59 下午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : main.go
   @Package : main
   @Description: 短链接生成服务 入口函数及主程序初始化逻辑
*/

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/offcn-jl/gaea-back-end/commons/config"
	"github.com/offcn-jl/gaea-back-end/commons/database/orm"
	"github.com/offcn-jl/gaea-back-end/commons/database/structs"
	"github.com/offcn-jl/gaea-back-end/commons/logger"
	"github.com/offcn-jl/gaea-back-end/commons/utils"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
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

	// 关闭 Gin 的控制台彩色输出
	gin.DisableConsoleColor()

	// 使用默认配置初始化路由
	router := gin.Default()

	// 添加版本号
	router.Use(func(c *gin.Context) {
		c.Header("Server", "Gaea - URL Shortener - "+config.Version)
	})

	// 判断运行环境配置路径
	path := "/test/:ID"
	if os.Getenv("GIN_MODE") == "release" {
		path = "/:ID"
	}

	// 添加路由
	router.Any(path, func(c *gin.Context) {
		if c.Request.URL.String() == "/favicon.ico" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		// 获取链接配置
		urlInfo := structs.ToolsUrlShortener{}
		orm.MySQL.Gaea.Unscoped().Where("id = ? OR custom_id = ?", utils.Base62Decode(c.Param("ID")), c.Param("ID")).Find(&urlInfo)

		// 读取请求 body
		requestBody, _ := ioutil.ReadAll(c.Request.Body)

		// 未找到页面
		if urlInfo.ID == 0 {
			/// 保存请求记录
			orm.MySQL.Gaea.Create(&structs.ToolsUrlShortenerRedirectLog{
				UrlID:         uint(utils.Base62Decode(c.Param("ID"))),
				CustomID:      c.Param("ID"),
				RequestProto:  c.Request.Proto,
				RequestMethod: c.Request.Method,
				RequestUrl:    c.Request.Host + c.Request.RequestURI,
				RequestBody:   string(requestBody),
				IP:            c.ClientIP(),
				UserAgent:     c.Request.UserAgent(),
				Status:        "NotFound",
			})
			c.Data(http.StatusNotFound, config.Get().ToolsUrlShortenerNotFoundContentType, []byte(config.Get().ToolsUrlShortenerNotFoundData))
			return
		}

		// 页面已禁用
		if urlInfo.DeletedAt != nil {
			// 保存请求记录
			orm.MySQL.Gaea.Create(&structs.ToolsUrlShortenerRedirectLog{
				UrlID:         urlInfo.ID,
				CustomID:      urlInfo.CustomID,
				URL:           urlInfo.URL,
				RequestProto:  c.Request.Proto,
				RequestMethod: c.Request.Method,
				RequestUrl:    c.Request.Host + c.Request.RequestURI,
				RequestBody:   string(requestBody),
				IP:            c.ClientIP(),
				UserAgent:     c.Request.UserAgent(),
				Status:        "Disabled",
			})
			c.Data(http.StatusNotFound, config.Get().ToolsUrlShortenerDisabledContentType, []byte(config.Get().ToolsUrlShortenerDisabledData))
			return
		}

		// 保存请求记录
		orm.MySQL.Gaea.Create(&structs.ToolsUrlShortenerRedirectLog{
			UrlID:         urlInfo.ID,
			CustomID:      urlInfo.CustomID,
			URL:           urlInfo.URL,
			RequestProto:  c.Request.Proto,
			RequestMethod: c.Request.Method,
			RequestUrl:    c.Request.Host + c.Request.RequestURI,
			RequestBody:   string(requestBody),
			IP:            c.ClientIP(),
			UserAgent:     c.Request.UserAgent(),
			Status:        "RedirectSuccess",
		})

		// 跳转页面
		if strings.Contains(urlInfo.URL, "?") {
			c.Redirect(http.StatusTemporaryRedirect, urlInfo.URL+"&"+c.Request.URL.RawQuery)
		} else {
			c.Redirect(http.StatusTemporaryRedirect, urlInfo.URL+"?"+c.Request.URL.RawQuery)
		}
	})

	// 初始化路由后启动监听
	if err := router.Run(); err != nil {
		// 处理错误
		logger.Panic(err)
	}
}
