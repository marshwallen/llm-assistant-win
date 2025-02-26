package tools

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"
)

var metrics = map[string]string{
	"cpu_clock": "MHz", 
	"mem_used": "bytes", 
	"disk_used": "bytes",
	"gpu_util": "%", 
	"gpu_mem_used": "MB", 
	"gpu_mem_clock": "MHz",
	"gpu_core_clock": "MHz", 
	"gpu_temp": "°C",
	"gpu_power": "W",
}

// 读取 CSV 文件的后 N 行中的某一列的值
func readCSVLastNColumn(filename string, columnIndex int, n int) ([]string, error) {
	// 打开 CSV 文件
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %w", err)
	}
	defer file.Close()

	// 创建 CSV 读取器
	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read csv failed: %w", err)
	}

	// 检查列索引是否有效
	if columnIndex < 0 || columnIndex >= len(rows[0]) {
		return nil, fmt.Errorf("invalid column index: %d", columnIndex)
	}

	// 计算起始行
	start := len(rows) - n
	if start < 0 {
		start = 0
	}

	// 提取指定列的值
	var values []string
	for i := len(rows) - 1; i >= start; i-- {
		values = append(values, rows[i][columnIndex])
	}
	return values, nil
}

const dateFormat = "20060102"

func GetSysHealthData(p map[string]interface{}) (out string, err error){
	_minutes, _ := p["minutes"].(float64)
	minutes := int(_minutes)

	currentDay := time.Now().Local().Format(dateFormat)
	for metric, unit := range metrics {
		values, err := readCSVLastNColumn(fmt.Sprintf("data/%s_%s.csv", metric, currentDay), 3, minutes*12)
		if err != nil {
			return "", err
		}
		out += fmt.Sprintf("%s 利用情况趋势如下 (每个指标间隔 5s, 单位为 %s): %s\n", metric, unit, fmt.Sprint(values))
	}
	return
}