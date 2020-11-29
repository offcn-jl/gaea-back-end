/*
   @Time : 2020/11/28 5:06 下午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : request
   @Software: GoLand
   @Description: 请求工具
*/

package request

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// GetSendQueryReceiveBytes 用 GET 发送 QueryString 类型的请求并接受 Bytes 类型的响应
func GetSendQueryReceiveBytes(path string, query map[string]string) ([]byte, error) {
	// 解析请求
	if urlObject, err := url.Parse(path); err != nil {
		return nil, err
	} else {
		// 将参数 map 拼接到 query 字符串中
		queryObject := urlObject.Query()
		for queryKey, queryValue := range query {
			queryObject.Set(queryKey, queryValue)
		}
		urlObject.RawQuery = queryObject.Encode()
		// 发送 GET 请求
		if responseData, err := http.Get(urlObject.String()); err != nil {
			// 发送 GET 请求出错
			return nil, err
		} else {
			defer responseData.Body.Close() // 函数退出时关闭 body
			if responseData.StatusCode != 200 {
				// 请求出错
				return nil, errors.New("发送 GET 请求出错. 状态码: " + fmt.Sprint(responseData.StatusCode))
			} else {
				// 读取 body
				if responseBytes, err := ioutil.ReadAll(responseData.Body); err != nil {
					return nil, err
				} else {
					return responseBytes, nil
				}
			}
		}
	}
}

// GetSendQueryReceiveJson 用 GET 发送 QueryString 类型的请求并接受 Json 类型的响应
func GetSendQueryReceiveJson(path string, query map[string]string) (map[string]interface{}, error) {
	if responseBytes, err := GetSendQueryReceiveBytes(path, query); err != nil {
		return nil, err
	} else {
		// 解码 responseBytes
		var responseJson map[string]interface{}
		if err := json.Unmarshal(responseBytes, &responseJson); err != nil {
			return nil, err
		} else {
			// 返回请求回来的 Json 的 Map
			return responseJson, nil
		}
	}
}
