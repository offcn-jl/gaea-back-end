/*
   @Time : 2020/11/6 4:05 下午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : verify
   @Software: GoLand
   @Description: 验证工具
*/

package verify

import (
	"crypto/md5"
	"crypto/sha1"
	"errors"
	"fmt"
	"github.com/offcn-jl/gaea-back-end/commons/config"
	"github.com/offcn-jl/gaea-back-end/commons/database/orm"
	"github.com/offcn-jl/gaea-back-end/commons/database/structs"
	"github.com/offcn-jl/gaea-back-end/commons/logger"
	"github.com/offcn-jl/gaea-back-end/commons/request"
	"regexp"
	"strconv"
	"time"
)

// Phone 用来验证手机号码是否有效
// 中国(严谨), 根据工信部2019年最新公布的手机号段
// 摘自 https://any86.github.io/any-rule/
// 去除首尾的 /
func Phone(phone string) bool {
	regular := `^(?:(?:\+|00)86)?1(?:(?:3[\d])|(?:4[5-7|9])|(?:5[0-3|5-9])|(?:6[5-7])|(?:7[0-8])|(?:8[\d])|(?:9[1|8|9]))\d{8}$`
	return regexp.MustCompile(regular).MatchString(phone)
}

// MisToken 校验 MIS 口令码 是否合法
func MisToken(misToken string) (bool, error) {
	// 获取当前系统中的 MIS 口令码
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	if responseJsonMap, err := request.GetSendQueryReceiveJson(config.Get().OffcnMisURL, map[string]string{"appid": config.Get().OffcnMisAppID, "sign": fmt.Sprintf("%x", sha1.Sum([]byte("appid="+config.Get().OffcnMisAppID+"&code="+fmt.Sprintf("%x", md5.Sum([]byte(config.Get().OffcnMisCode)))+"&noncestr=gaea&timestamp="+timestamp+"&token="+config.Get().OffcnMisToken+"&url=http://chaos.jilinoffcn.com/"))), "noncestr": "gaea", "timestamp": timestamp}); err != nil {
		// 请求失败
		return false, err
	} else {
		// 判断是否获取成功
		if responseJsonMap["status"].(float64) != 1 {
			logger.DebugToJson("响应内容", responseJsonMap)
			// 获取 MIS TOKEN 失败
			if responseJsonMap["msg"] != nil {
				return false, errors.New(responseJsonMap["msg"].(string))
			} else {
				return false, errors.New("请求 MIS 口令码 失败")
			}
		} else {
			// 判断 MIS TOKEN 是否有效
			if misToken != responseJsonMap["data"] {
				// MIS Token 不匹配
				return false, nil
			} else {
				// 对 UUID 对应的 Session 进行更新 MIS Token 的操作
				return true, nil
			}
		}
	}
}

// IsSubordinateRole 检查是否是下属角色
// 参数1: superior, 上级
// 参数2: subordinate, 下属
func IsSubordinateRole(superior, subordinate uint) bool {
	// 检查传入的上级
	if superior == 0 {
		// 避免传入空上级导致提权
		return false
	}

	// 取出全部下属角色
	var subordinateRoles []structs.SystemRole
	orm.MySQL.Gaea.Where("superior_id = ?", superior).Find(&subordinateRoles)

	// 遍历全部下属角色, 如果找到匹配的, 则返回真, 并结束遍历
	if len(subordinateRoles) != 0 {
		// 遍历当前一层的
		for index := range subordinateRoles {
			// 匹配后返回真
			if subordinateRoles[index].ID == subordinate {
				return true
			}
		}
		// 递归遍历更深层的
		for index := range subordinateRoles {
			// 递归结果为真时, 返回真
			if IsSubordinateRole(subordinateRoles[index].ID, subordinate) {
				return true
			}
		}
	}

	// 遍历了全部下属角色并且没找到匹配的结果, 返回假
	return false
}

// PasswordComplexity 检查密码复杂度是否符合标准
func PasswordComplexity(password string) (bool, error) {
	// 检查密码长度
	if len(password) < 8 {
		return false, errors.New("密码长度不足 8 位")
	}

	// 检查密码中是否包含数字
	if match, _ := regexp.MatchString(`[0-9]+`, password); !match {
		return false, errors.New("密码中应当包含数字")
	}

	// 检查密码中是否包含小写字母
	if match, _ := regexp.MatchString(`[a-z]+`, password); !match {
		return false, errors.New("密码中应当包含小写字母")
	}

	// 检查密码中是否包含大写字母
	if match, _ := regexp.MatchString(`[A-Z]+`, password); !match {
		return false, errors.New("密码中应当包含大写字母")
	}

	// 检查密码中是否包含特殊符号
	if match, _ := regexp.MatchString(`[~!@#$%^&*?_-]+`, password); !match {
		return false, errors.New("密码中应当包含特殊符号，如 ~!@#$%^&*?_- ")
	}

	// 通过检查
	return true, nil
}
