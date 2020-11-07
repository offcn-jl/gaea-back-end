/*
   @Time : 2020/11/2 9:11 上午
   @Author : ShadowWalker
   @Email : master@rebeta.cn
   @File : orm
   @Software: GoLand
*/

package orm

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/offcn-jl/gaea-back-end/commons/config"
	"github.com/offcn-jl/gaea-back-end/commons/database/structs"
	"github.com/offcn-jl/gaea-back-end/commons/logger"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"time"
	"unicode"
)

func init() {
	// 替换 gorm 的 LogFormatter , 移除其中输出日志颜色的逻辑, 优化 gorm 日志的可读性
	gorm.LogFormatter = func(values ...interface{}) (messages []interface{}) {
		isPrintable := func(s string) bool {
			for _, r := range s {
				if !unicode.IsPrint(r) {
					return false
				}
			}
			return true
		}
		if len(values) > 1 {
			var (
				sql             string
				formattedValues []string
				level           = values[0]
				currentTime     = "\nGORM -> [" + gorm.NowFunc().Format("2006-01-02 15:04:05") + "]"
				source          = fmt.Sprintf("(%v)", values[1])
			)
			messages = []interface{}{source, currentTime}
			if len(values) == 2 {
				//remove the line break
				currentTime = currentTime[1:]
				//remove the brackets
				source = fmt.Sprintf("%v", values[1])

				messages = []interface{}{currentTime, source}
			}
			if level == "sql" {
				// duration
				messages = append(messages, fmt.Sprintf(" [%.2fms] ", float64(values[2].(time.Duration).Nanoseconds()/1e4)/100.0))
				// sql
				for _, value := range values[4].([]interface{}) {
					indirectValue := reflect.Indirect(reflect.ValueOf(value))
					if indirectValue.IsValid() {
						value = indirectValue.Interface()
						if t, ok := value.(time.Time); ok {
							if t.IsZero() {
								formattedValues = append(formattedValues, fmt.Sprintf("'%v'", "0000-00-00 00:00:00"))
							} else {
								formattedValues = append(formattedValues, fmt.Sprintf("'%v'", t.Format("2006-01-02 15:04:05")))
							}
						} else if b, ok := value.([]byte); ok {
							if str := string(b); isPrintable(str) {
								formattedValues = append(formattedValues, fmt.Sprintf("'%v'", str))
							} else {
								formattedValues = append(formattedValues, "'<binary>'")
							}
						} else if r, ok := value.(driver.Valuer); ok {
							if value, err := r.Value(); err == nil && value != nil {
								formattedValues = append(formattedValues, fmt.Sprintf("'%v'", value))
							} else {
								formattedValues = append(formattedValues, "NULL")
							}
						} else {
							switch value.(type) {
							case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
								formattedValues = append(formattedValues, fmt.Sprintf("%v", value))
							default:
								formattedValues = append(formattedValues, fmt.Sprintf("'%v'", value))
							}
						}
					} else {
						formattedValues = append(formattedValues, "NULL")
					}
				}
				// differentiate between $n placeholders or else treat like ?
				if regexp.MustCompile(`\$\d+`).MatchString(values[3].(string)) {
					sql = values[3].(string)
					for index, value := range formattedValues {
						placeholder := fmt.Sprintf(`\$%d([^\d]|$)`, index+1)
						sql = regexp.MustCompile(placeholder).ReplaceAllString(sql, value+"$1")
					}
				} else {
					formattedValuesLength := len(formattedValues)
					for index, value := range regexp.MustCompile(`\?`).Split(values[3].(string), -1) {
						sql += value
						if index < formattedValuesLength {
							sql += formattedValues[index]
						}
					}
				}
				messages = append(messages, sql)
				messages = append(messages, fmt.Sprintf(" \n[%v] ", strconv.FormatInt(values[5].(int64), 10)+" rows affected or returned "))
			} else {
				//messages = append(messages, "\033[31;1m")
				messages = append(messages, values[2:]...)
				//messages = append(messages, "\033[0m")
			}
		}
		return
	}
}

// MySQL 对外暴露的数据库实例
var MySQL struct {
	Gaea *gorm.DB
}

// Init 初始化数据库
func Init() error {
	// 校验是否配置 MYSQL_DSN
	if os.Getenv("MYSQL_DSN_GAEA") == "" {
		// 未配置，返回错误
		return errors.New("未配置 MYSQL_DSN_GAEA , 请在环境变量中配置 MYSQL_DSN_GAEA ( 例 : user:password@tcp(hostname)/database?charset=utf8mb4&parseTime=True&loc=Local )")
	}
	// 在调试模式时，输出环境变量中配置的 MYSQL_DSN_GAEA
	logger.DebugToString("os.Getenv(\"MYSQL_DSN_GAEA\")", os.Getenv("MYSQL_DSN_GAEA"))
	// 已配置, 初始化 MYSQL 客户端
	var err error
	if MySQL.Gaea, err = gorm.Open("mysql", os.Getenv("MYSQL_DSN_GAEA")); err != nil {
		return err
	}
	// 初始化成功后，判断是否需要进行数据库表结构迁移
	if len(config.Version) != 31 {
		logger.Log("应用版本号有误, 无法计算构建时间, 开始进行数据库结构自动迁移.")
		autoMigrate()
	} else {
		// 通过版本号, 计算构建时间
		builtTime, _ := time.ParseInLocation("2006/01/02 15:04:05", config.Version[10:29], time.Local) // 使用 parseInLocation 将字符串格式化返回本地时区时间
		// 判断构建时间是否超过一小时
		if time.Since(builtTime).Hours() > 1 {
			// 超过一小时, 不需要进行迁移
			logger.Log("应用构建于 " + fmt.Sprint(time.Since(builtTime)) + " 前, 跳过数据库结构自动迁移.")
		} else {
			// 没有超过一小时, 进行表结构自动迁移
			logger.Log("应用构建时间没有超过一小时, 开始进行数据库结构自动迁移.")
			autoMigrate()
		}
	}
	return nil
}

// autoMigrate 可以完成表结构自动迁移
func autoMigrate() {
	MySQL.Gaea.AutoMigrate(
		// system 系统
		&structs.SystemConfig{},
		// SingleSignOn 单点登陆
		structs.SingleSignOnLoginModule{},
		structs.SingleSignOnVerificationCode{},
		structs.SingleSignOnUser{},
		structs.SingleSignOnSession{},
		structs.SingleSignOnSuffix{},
		structs.SingleSignOnOrganization{},
		structs.SingleSignOnCRMRoundLog{},
		structs.SingleSignOnErrorLog{},
	)
}
