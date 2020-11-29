/*
   @Time : 2020/11/8 9:35 上午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : response
   @Software: GoLand
   @Description: 用于快速的创建响应内容
*/

package response

// Struct 响应内容的结构
type Struct map[string]interface{}

// Message 创建响应的快捷方式
func Message(messageText string) Struct {
	return Struct{"Message": messageText}
}

// Data 创建带数据的响应的快捷方式
func Data(data interface{}) Struct {
	return Struct{"Message": "Success", "Data": data}
}

// Error 创建错误响应的快捷方式
func Error(messageText string, err error) Struct {
	return Struct{"Message": messageText, "Error": err.Error()}
}

// json 结构体作为接口, 对其提供方法作为 Json 类响应的快捷方式
type json struct{}

// Invalid 非法 Json 错误响应的快捷方式
func (json) Invalid(err error) Struct {
	return Error("提交的 Json 数据不正确", err)
}

// query 结构体作为接口, 对其提供方法作为 Query 类响应的快捷方式
type query struct{}

// Invalid 非法 Query 错误响应的快捷方式
func (query) Invalid(err error) Struct {
	return Error("提交的 Query 查询不正确", err)
}

// 常用的响应
var (
	// Success 成功响应
	Success = Message("Success")
	// Json Json类的响应
	Json = json{}
	// Query Query类的响应
	Query = query{}
	// Phone 手机号码类的响应
	Phone = struct {
		Invalid Struct
	}{
		Message("手机号码不正确"),
	}
)
