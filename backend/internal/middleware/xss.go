package middleware

import (
	"bytes"
	"encoding/json"
	"html"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
)

// XSSProtectionMiddleware 防止XSS攻击的中间件
func XSSProtectionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只对POST、PUT、PATCH请求进行XSS防护
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			// 读取请求体
			body, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "读取请求体失败"})
				c.Abort()
				return
			}

			// 如果请求体为空，直接继续
			if len(body) == 0 {
				c.Next()
				return
			}

			// 检查Content-Type
			contentType := c.GetHeader("Content-Type")
			if strings.Contains(contentType, "application/json") {
				// 处理JSON请求
				cleanedBody, err := sanitizeJSON(body)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "JSON格式错误"})
					c.Abort()
					return
				}

				// 重新设置请求体
				c.Request.Body = io.NopCloser(bytes.NewBuffer(cleanedBody))
			} else if strings.Contains(contentType, "application/x-www-form-urlencoded") {
				// 处理表单数据
				cleanedBody := sanitizeFormData(body)
				c.Request.Body = io.NopCloser(bytes.NewBuffer(cleanedBody))
			}
		}

		c.Next()
	}
}

// sanitizeJSON 清理JSON数据中的XSS内容
func sanitizeJSON(data []byte) ([]byte, error) {
	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return nil, err
	}

	// 递归清理JSON数据
	cleanedData := sanitizeValue(jsonData)

	// 重新序列化
	cleanedBytes, err := json.Marshal(cleanedData)
	if err != nil {
		return nil, err
	}

	return cleanedBytes, nil
}

// sanitizeFormData 清理表单数据中的XSS内容
func sanitizeFormData(data []byte) []byte {
	formData := string(data)

	// 分割键值对
	pairs := strings.Split(formData, "&")
	var cleanedPairs []string

	for _, pair := range pairs {
		if pair == "" {
			continue
		}

		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			key := html.EscapeString(kv[0])
			value := html.EscapeString(kv[1])
			cleanedPairs = append(cleanedPairs, key+"="+value)
		} else {
			cleanedPairs = append(cleanedPairs, pair)
		}
	}

	return []byte(strings.Join(cleanedPairs, "&"))
}

// sanitizeValue 递归清理值中的XSS内容
func sanitizeValue(value interface{}) interface{} {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case string:
		// 对字符串进行HTML转义
		return html.EscapeString(v)
	case []interface{}:
		// 处理数组
		cleanedArray := make([]interface{}, len(v))
		for i, item := range v {
			cleanedArray[i] = sanitizeValue(item)
		}
		return cleanedArray
	case map[string]interface{}:
		// 处理对象
		cleanedMap := make(map[string]interface{})
		for key, val := range v {
			// 清理键名
			cleanedKey := html.EscapeString(key)
			// 递归清理值
			cleanedMap[cleanedKey] = sanitizeValue(val)
		}
		return cleanedMap
	default:
		// 对于其他类型，尝试使用反射进行深度清理
		return sanitizeWithReflection(value)
	}
}

// sanitizeWithReflection 使用反射深度清理结构体
func sanitizeWithReflection(value interface{}) interface{} {
	val := reflect.ValueOf(value)

	// 如果是指针，获取指向的值
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
	}

	// 根据类型进行处理
	switch val.Kind() {
	case reflect.String:
		return html.EscapeString(val.String())
	case reflect.Slice:
		// 处理切片
		result := reflect.MakeSlice(val.Type(), val.Len(), val.Cap())
		for i := 0; i < val.Len(); i++ {
			element := val.Index(i)
			if element.CanInterface() {
				cleanedElement := sanitizeWithReflection(element.Interface())
				result.Index(i).Set(reflect.ValueOf(cleanedElement))
			}
		}
		return result.Interface()
	case reflect.Struct:
		// 处理结构体
		result := reflect.New(val.Type()).Elem()
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			fieldType := val.Type().Field(i)

			// 只处理可导出的字段
			if fieldType.PkgPath == "" && field.CanInterface() {
				cleanedField := sanitizeWithReflection(field.Interface())
				if result.Field(i).CanSet() {
					result.Field(i).Set(reflect.ValueOf(cleanedField))
				}
			}
		}
		return result.Interface()
	default:
		// 对于其他类型，直接返回原值
		return value
	}
}

// XSSWhitelistMiddleware 白名单中间件，跳过某些路径的XSS检查
func XSSWhitelistMiddleware(skipPaths []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查当前路径是否在白名单中
		for _, path := range skipPaths {
			if strings.HasPrefix(c.Request.URL.Path, path) {
				c.Next()
				return
			}
		}

		// 不在白名单中，执行XSS防护
		XSSProtectionMiddleware()(c)
	}
}
