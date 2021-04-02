/*
   @Time : 2020/11/28 5:17 下午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : mini_program
   @Software: GoLand
   @Description: 微信小程序
*/

package wechat

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/offcn-jl/gaea-back-end/commons/config"
	"github.com/offcn-jl/gaea-back-end/commons/logger"
	"github.com/offcn-jl/gaea-back-end/commons/request"
)

// MiniProgramCreateQrCode 创建微信小程序码
//
// page 类型：string; 默认值：主页; 必填：否; 说明：必须是已经发布的小程序存在的页面（否则报错），例如 pages/index/index, 根路径前不要填加 /,不能携带参数（参数请放在scene字段里），如果不填写这个字段，默认跳主页面
//
// scene 类型：string; 默认值：无; 必填：是; 说明：最大32个可见字符，只支持数字，大小写英文以及部分特殊字符：!#$&'()*+,/:;=?@-._~，其它字符请自行编码为合法字符（因不支持%，中文无法使用 urlencode 处理，请使用其他编码方式）
//
// width 类型：number; 默认值：430; 必填：否; 说明：二维码的宽度，单位 px，最小 280px，最大 1280px
//
// autoColor 类型：boolean; 默认值：false; 必填：否; 说明：自动配置线条颜色，如果颜色依然是黑色，则说明不建议配置主色调，默认 false
//
// lineColor 类型：Object; 默认值：{"r":0,"g":0,"b":0}; 必填：否; 说明：auto_color 为 false 时生效，使用 rgb 设置颜色 例如 {"r":"xxx","g":"xxx","b":"xxx"} 十进制表示
//
// isHyaline 类型：boolean; 默认值：false; 必填：否; 说明：是否需要透明底色，为 true 时，生成透明底色的小程序
func MiniProgramCreateQrCode(appID, page, scene string, width uint, autoColor bool, lineColor map[string]uint, isHyaline bool) ([]byte, error) {
	// 获取访问令牌
	if accessToken, err := MiniProgramGetAccessToken(appID); err != nil {
		return nil, err
	} else {
		// 调用接口
		if responseData, err := request.PostSendJsonReceiveBytes("https://api.weixin.qq.com/wxa/getwxacodeunlimit?access_token="+accessToken, map[string]interface{}{"page": page, "scene": scene, "width": width, "auto_color": autoColor, "line_color": lineColor, "is_hyaline": isHyaline}); err != nil {
			return nil, err
		} else {
			// 解码 body
			var responseJsonMap map[string]interface{}
			if err := json.Unmarshal(responseData, &responseJsonMap); err != nil {
				// 不能格式化为 Json 代表返回来的是buffer，所以请求成功
				return responseData, nil
			} else {
				// 成功被格式化为 Json 代表请求出错
				logger.DebugToJson("responseJsonMap", responseJsonMap)
				return nil, errors.New("创建小程序码失败 [ " + fmt.Sprint(responseJsonMap["errcode"]) + " ] " + responseJsonMap["errmsg"].(string))
			}
		}
	}
}

// MiniProgramGetAccessToken 获取微信小程序访问令牌
// 调用生产环境的接口获取微信小程序访问令牌的快捷方式
func MiniProgramGetAccessToken(appID string) (string, error) {
	if responseJson, err := request.GetSendQueryReceiveJson("https://api.gaea.jilinoffcn.com/release/services/authentication/mini-program/get/access-token", map[string]string{"access-token": config.Get().ServicesAccessToken, "app-id": appID}); err != nil {
		return "", err
	} else {
		if responseJson["Message"] != "Success" {
			return "", errors.New(responseJson["Message"].(string))
		}
		return responseJson["Data"].(string), nil
	}
}
