package utils

import (
	"encoding/json"
	"fmt"
)

// 自定义类型，用于存储float64或string
type Float64OrString struct {
	IsFloat  bool    // 标志是否为float64类型
	FloatVal float64 // 存储float64值
	StrVal   string  // 存储string值
}

// 实现json.Unmarshaler接口，处理解析逻辑
func (f *Float64OrString) UnmarshalJSON(data []byte) error {
	// 先尝试解析为float64
	var floatVal float64
	if err := json.Unmarshal(data, &floatVal); err == nil {
		f.IsFloat = true
		f.FloatVal = floatVal
		f.StrVal = ""
		return nil
	}

	// 解析float64失败，尝试解析为string
	var strVal string
	if err := json.Unmarshal(data, &strVal); err == nil {
		f.IsFloat = false
		f.StrVal = strVal
		f.FloatVal = 0
		return nil
	}

	// 两种类型都解析失败
	return fmt.Errorf("无法解析为float64或string: %s", data)
}

// 实现json.Marshaler接口（序列化）
func (f *Float64OrString) MarshalJSON() ([]byte, error) {
	if f.IsFloat {
		// 序列化float64值为JSON数字
		return json.Marshal(f.FloatVal)
	} else {
		// 序列化string值为JSON字符串（自动处理引号和转义）
		return json.Marshal(f.StrVal)
	}
}
