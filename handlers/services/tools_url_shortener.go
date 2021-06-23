/*
   @Time : 2021/4/2 1:42 下午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : tools_url_shortener.go
   @Package : services
   @Description: 工具 短链接生成器 ( 长链接转短链接 ) [ 临时使用 ]
*/

package services

import (
	"github.com/gin-gonic/gin"
	"github.com/offcn-jl/gaea-back-end/commons/config"
	"github.com/offcn-jl/gaea-back-end/commons/database/orm"
	"github.com/offcn-jl/gaea-back-end/commons/database/structs"
	"github.com/offcn-jl/gaea-back-end/commons/logger"
	"github.com/offcn-jl/gaea-back-end/commons/response"
	"github.com/offcn-jl/gaea-back-end/commons/utils"
	"net/http"
	"os"
)

// ToolsUrlShortenerCreateShortLink 新建短链接
func ToolsUrlShortenerCreateShortLink(c *gin.Context) {
	// 校验 AccessToken 是否合法
	if c.Query("access-token") != config.Get().ServicesAccessToken {
		c.JSON(http.StatusBadRequest, response.Message("AccessToken 不正确"))
		return
	}

	// 绑定参数
	requestInfo := struct {
		URL string `binding:"required"`
	}{}
	if err := c.ShouldBindJSON(&requestInfo); err != nil {
		logger.Error(err)
		c.JSON(http.StatusBadRequest, response.Json.Invalid(err))
		return
	}

	basePath := "https://offcn.ltd/test/"
	if os.Getenv("GIN_MODE") == "release" {
		basePath = "https://offcn.ltd/"
	}

	// 检查是否已经存在短链记录
	checkInfo := structs.ToolsUrlShortener{}
	orm.MySQL.Gaea.Where("url = ?", requestInfo.URL).Find(&checkInfo)
	if checkInfo.ID != 0 {
		c.JSON(http.StatusOK, response.Data(map[string]interface{}{
			"Repetitive": true,                                        // 是否重复
			"ShortUrl":   basePath + utils.Base62Encode(checkInfo.ID), // 短链接
		}))
		return
	}

	saveInfo := structs.ToolsUrlShortener{
		CreatedUserID: 1,
		UpdatedUserID: 1,
		URL:           requestInfo.URL,
	}

	// 创建记录
	orm.MySQL.Gaea.Create(&saveInfo)

	// 返回创建成功
	c.JSON(http.StatusOK, response.Data(map[string]interface{}{
		"Repetitive": false,                                      // 是否重复
		"ShortUrl":   basePath + utils.Base62Encode(saveInfo.ID), // 短链接
	}))
}
