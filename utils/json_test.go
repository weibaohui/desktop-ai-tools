package utils

import (
	"testing"
)

// TestToJSON 测试ToJSON函数
func TestToJSON(t *testing.T) {
	// 测试数据
	testData := map[string]interface{}{
		"name":    "test",
		"age":     25,
		"active":  true,
		"details": map[string]string{"city": "Beijing"},
	}

	// 调用ToJSON函数
	jsonStr, err := ToJSON(testData)
	if err != nil {
		t.Fatalf("ToJSON失败: %v", err)
	}

	// 验证结果不为空
	if jsonStr == "" {
		t.Fatal("JSON字符串为空")
	}

	t.Logf("格式化JSON输出:\n%s", jsonStr)
}

// TestToJSONCompact 测试ToJSONCompact函数
func TestToJSONCompact(t *testing.T) {
	// 测试数据
	testData := map[string]interface{}{
		"name": "test",
		"age":  25,
	}

	// 调用ToJSONCompact函数
	jsonStr, err := ToJSONCompact(testData)
	if err != nil {
		t.Fatalf("ToJSONCompact失败: %v", err)
	}

	// 验证结果不为空
	if jsonStr == "" {
		t.Fatal("JSON字符串为空")
	}

	t.Logf("紧凑JSON输出: %s", jsonStr)
}

// TestPrintJSON 测试PrintJSON函数
func TestPrintJSON(t *testing.T) {
	// 测试数据
	testData := map[string]interface{}{
		"message": "这是一个测试",
		"status":  "success",
	}

	// 测试带前缀的打印
	t.Log("测试PrintJSON函数（带前缀）:")
	PrintJSON(testData, "测试数据")

	// 测试不带前缀的打印
	t.Log("测试PrintJSON函数（不带前缀）:")
	PrintJSON(testData)
}