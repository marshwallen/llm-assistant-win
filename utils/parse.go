package utils

import (
    "fmt"
    "strconv"
    "strings"
)

// 解析带百分比的浮点数（如 "75 %" → 75.0）
func ParseFloat(s string) float64 {
    s = strings.TrimSpace(s)
    // 移除所有非数字、小数点、负号以外的字符（如 "%", "W" 等）
    cleaned := strings.Map(func(r rune) rune {
        if (r >= '0' && r <= '9') || r == '.' || r == '-' {
            return r
        }
        return -1 // 过滤其他字符
    }, s)
    
    value, err := strconv.ParseFloat(cleaned, 64)
    if err != nil {
        return 0
    }
    return value
}

// 解析带单位的无符号整数（如 "1500 MHz" → 1500）
func ParseUint(s string) uint64 {
    s = strings.TrimSpace(s)
    // 提取数字部分
    var numStr strings.Builder
    for _, c := range s {
        if c >= '0' && c <= '9' {
            numStr.WriteRune(c)
        } else if numStr.Len() > 0 {
            break // 遇到非数字字符时停止
        }
    }
    
    value, err := strconv.ParseUint(numStr.String(), 10, 64)
    if err != nil {
        return 0
    }
    return value
}

// 解析显存容量（如 "8192 MiB" → 8192）
func ParseMemory(s string) uint64 {
    var value uint64
    // 直接提取开头的数字部分（忽略单位和空格）
    fmt.Sscanf(s, "%d", &value)
    return value
}