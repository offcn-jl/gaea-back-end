/*
   @Time : 2021/6/29 10:11 上午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : occ.go
   @Package : services
   @Description: OCC 相关业务
*/

package services

import (
	"crypto/md5"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/offcn-jl/gaea-back-end/commons/config"
	"github.com/offcn-jl/gaea-back-end/commons/response"
	"net/http"
	"time"
)

// OCCGetSign 获取 OCC 接口调用 Sign
func OCCGetSign(c *gin.Context) {
	// 校验 AccessToken 是否合法
	if c.Query("access-token") != config.Get().ServicesAccessToken {
		c.JSON(http.StatusBadRequest, response.Message("AccessToken 不正确"))
		return
	}

	// 返回 Sign
	c.JSON(http.StatusOK, response.Data(fmt.Sprintf("%x", md5.Sum([]byte(config.Get().OffcnOCCKey+time.Now().Format("20060102"))))))
}
