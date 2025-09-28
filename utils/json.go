package utils

import (
	"encoding/json"
	"fmt"
)

// ToJSON 将任意结构体转换为格式化的JSON字符串
// 参数:
//   - data: 要转换的数据结构
// 返回值:
//   - string: 格式化的JSON字符串
//   - error: 转换过程中的错误
func ToJSON(data interface{}) (string, error) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("JSON序列化失败: %v", err)
	}
	return string(jsonBytes), nil
}

// ToJSONCompact 将任意结构体转换为紧凑的JSON字符串
// 参数:
//   - data: 要转换的数据结构
// 返回值:
//   - string: 紧凑的JSON字符串
//   - error: 转换过程中的错误
func ToJSONCompact(data interface{}) (string, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("JSON序列化失败: %v", err)
	}
	return string(jsonBytes), nil
}

// PrintJSON 打印格式化的JSON到控制台
// 参数:
//   - data: 要打印的数据结构
//   - prefix: 打印前缀（可选）
func PrintJSON(data interface{}, prefix ...string) {
	jsonStr, err := ToJSON(data)
	if err != nil {
		fmt.Printf("JSON打印失败: %v\n", err)
		return
	}
	
	if len(prefix) > 0 {
		fmt.Printf("%s: %s\n", prefix[0], jsonStr)
	} else {
		fmt.Println(jsonStr)
	}
}

// PrintJSONCompact 打印紧凑的JSON到控制台
// 参数:
//   - data: 要打印的数据结构
//   - prefix: 打印前缀（可选）
func PrintJSONCompact(data interface{}, prefix ...string) {
	jsonStr, err := ToJSONCompact(data)
	if err != nil {
		fmt.Printf("JSON打印失败: %v\n", err)
		return
	}
	
	if len(prefix) > 0 {
		fmt.Printf("%s: %s\n", prefix[0], jsonStr)
	} else {
		fmt.Println(jsonStr)
	}
}