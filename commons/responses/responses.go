/*
   @Time : 2020/11/8 9:35 上午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : responses
   @Software: GoLand
   @Description: 用于快速的创建响应内容
*/

package responses

// ResponseStruct 响应内容的结构体
type ResponseStruct map[string]interface{}

// Message 创建响应的快捷方式
func Message(messageText string) ResponseStruct {
	return ResponseStruct{"Message": messageText}
}

// Data 创建带数据的响应的快捷方式
func Data(data interface{}) ResponseStruct {
	return ResponseStruct{"Message": "Success", "Data": data}
}

// Error 创建错误响应的快捷方式
func Error(messageText string, err error) ResponseStruct {
	return ResponseStruct{"Message": messageText, "Error": err.Error()}
}

// json 结构体作为接口, 对其提供方法作为 Json 类响应的快捷方式
type json struct{}

// Invalid 非法 Json 错误响应的快捷方式
func (json) Invalid(err error) ResponseStruct {
	return Error("提交的 Json 数据不正确", err)
}

// 常用的响应
var (
	// Success 成功响应
	Success = Message("Success")
	// Json Json类的响应
	Json = json{}
	// Phone 手机号码类的响应
	Phone = struct {
		Invalid ResponseStruct
	}{
		Message("手机号码不正确"),
	}
)
