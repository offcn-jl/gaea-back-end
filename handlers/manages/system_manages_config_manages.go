/*
   @Time : 2020/12/23 11:16 上午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : system_manages_config_manages
   @Description: 系统管理 - 配置管理
*/

package manages

import (
	"github.com/gin-gonic/gin"
	"github.com/offcn-jl/gaea-back-end/commons/config"
	"github.com/offcn-jl/gaea-back-end/commons/database/orm"
	"github.com/offcn-jl/gaea-back-end/commons/database/structs"
	"github.com/offcn-jl/gaea-back-end/commons/logger"
	"github.com/offcn-jl/gaea-back-end/commons/response"
	"github.com/offcn-jl/gaea-back-end/commons/utils"
	"net/http"
)

// SystemManagesConfigManagesPaginationGetConfig 分页获取配置列表
func SystemManagesConfigManagesPaginationGetConfig(c *gin.Context) {
	var configList []structs.SystemConfig
	orm.MySQL.Gaea.Offset((utils.StringToInt(c.Param("Page")) - 1) * utils.StringToInt(c.Param("Limit"))).Limit(utils.StringToInt(c.Param("Limit"))).Order("id DESC").Find(&configList)
	total := 0
	orm.MySQL.Gaea.Model(structs.SystemConfig{}).Count(&total)
	c.JSON(http.StatusOK, response.PaginationData(configList, total))
}

// SystemManagesConfigManagesUpdateConfig 修改配置
func SystemManagesConfigManagesUpdateConfig(c *gin.Context) {
	// 绑定数据
	requestJsonMap := structs.SystemConfig{}
	// 绑定数据
	if err := c.ShouldBindJSON(&requestJsonMap); err != nil {
		// 绑定数据错误
		logger.Error(err)
		c.JSON(http.StatusBadRequest, response.Json.Invalid(err))
		return
	}

	// 修改配置
	config.Update(orm.MySQL.Gaea, requestJsonMap)

	c.JSON(http.StatusOK, response.Success)
}
