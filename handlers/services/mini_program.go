/*
   @Time : 2021/7/1 9:37 上午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : mini_program.go
   @Package : services
   @Description: 小程序内部接口
*/

package services

import (
	"github.com/gin-gonic/gin"
	"github.com/offcn-jl/gaea-back-end/commons/database/orm"
	"github.com/offcn-jl/gaea-back-end/commons/database/structs"
	"github.com/offcn-jl/gaea-back-end/commons/response"
	"github.com/offcn-jl/gaea-back-end/commons/utils"
	"net/http"
)

// MiniProgramPhotoProcessingConfigList 按照查询参数获取照片处理配置列表
func MiniProgramPhotoProcessingConfigList(c *gin.Context) {
	page := utils.StringToInt(c.Query("page"))
	if page == 0 {
		page = 1
	}
	limit := utils.StringToInt(c.Query("limit"))
	if limit == 0 || limit > 100 {
		limit = 10
	}

	// 基本查询语句
	sql := "SELECT id,`name`,millimeter_width,millimeter_height,pixel_width,pixel_height FROM mini_program_photo_processing_configs WHERE (deleted_at > NOW() OR deleted_at IS NULL) AND "
	// 符合条件的数据总量
	total := 0
	// 定义数据结构及用于保存数据的数组
	var list []struct {
		ID               uint   // ID
		Name             string // 照片处理名称
		MillimeterWidth  uint   // MM 毫米 宽度
		MillimeterHeight uint   // MM 毫米 高度
		PixelWidth       uint   // PX 像素 宽度
		PixelHeight      uint   // PX 像素 高度
	}

	// 按照参数拼接查询条件
	if c.Query("search") != "" {
		// search 模糊搜索 名称、尺寸
		orm.MySQL.Gaea.Debug().Raw(sql+"( `name` Like ? OR millimeter_width Like ? OR millimeter_height Like ? OR pixel_width Like ? OR pixel_height Like ? )", "%"+c.Query("search")+"%", "%"+c.Query("search")+"%", "%"+c.Query("search")+"%", "%"+c.Query("search")+"%", "%"+c.Query("search")+"%").Offset((page - 1) * limit).Limit(limit).Order("id DESC").Scan(&list)
		orm.MySQL.Gaea.Model(structs.MiniProgramPhotoProcessingConfig{}).Where("(deleted_at > NOW() OR deleted_at IS NULL) AND `name` Like ? OR millimeter_width Like ? OR millimeter_height Like ? OR pixel_width Like ? OR pixel_height Like ?", "%"+c.Query("search")+"%", "%"+c.Query("search")+"%", "%"+c.Query("search")+"%", "%"+c.Query("search")+"%", "%"+c.Query("search")+"%").Count(&total)
	} else if c.Query("project") != "" {
		// project 强匹配 项目
		orm.MySQL.Gaea.Raw(sql+"project = ?", c.Query("project")).Offset((page - 1) * limit).Limit(limit).Order("id DESC").Scan(&list)
		orm.MySQL.Gaea.Model(structs.MiniProgramPhotoProcessingConfig{}).Where("(deleted_at > NOW() OR deleted_at IS NULL) AND project = ?", c.Query("project")).Count(&total)

	} else {
		// 未配置条件 强匹配 热门
		orm.MySQL.Gaea.Raw(sql+"hot = ?", true).Offset((page - 1) * limit).Limit(limit).Order("id DESC").Scan(&list)
		orm.MySQL.Gaea.Model(structs.MiniProgramPhotoProcessingConfig{}).Where("(deleted_at > NOW() OR deleted_at IS NULL) AND hot = ?", true).Count(&total)
	}

	// 返回数据
	c.JSON(http.StatusOK, response.PaginationData(list, total))
}

// MiniProgramPhotoProcessingConfig 获取照片处理配置
func MiniProgramPhotoProcessingConfig(c *gin.Context) {
	info := struct {
		Name             string // 照片处理名称
		Project          string // 项目
		CRMEventFormID   uint   // CRM 活动表单 ID
		CRMEventFormSID  string // CRM 活动表单 SID
		MillimeterWidth  uint   // MM 毫米 宽度
		MillimeterHeight uint   // MM 毫米 高度
		PixelWidth       uint   // PX 像素 宽度
		PixelHeight      uint   // PX 像素 高度
		BackgroundColors string // 背景色列表
		Description      string // 备注
	}{}

	orm.MySQL.Gaea.Unscoped().Model(structs.MiniProgramPhotoProcessingConfig{}).Where("deleted_at > NOW() AND id = ?", c.Param("ID")).Scan(&info)

	// 检查是否成功获取到配置
	if info.Name == "" {
		c.JSON(http.StatusNotFound, response.Message("ID 为 "+c.Param("ID")+" 的照片处理不存在"))
		return
	}

	c.JSON(http.StatusOK, response.Data(info))
}
