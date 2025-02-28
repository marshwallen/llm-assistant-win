package utils

import (
    "fmt"
    "os"
    "path/filepath"
    "gopkg.in/yaml.v3"
    "winds-assistant/common"
)

const CONFIG_FILE = "config/llm_settings.yaml"

// 从指定路径加载配置文件并解析为 LLMConfig 结构体。
func LoadCfg() (config *common.LLMConfig, err error) {
    // 读取文件内容
    data, err := os.ReadFile(filepath.Clean(CONFIG_FILE))
    if err != nil {
        fmt.Printf("error read yaml: %v", err)
        return nil, err
    }

    // 解析 Yaml
    if err := yaml.Unmarshal(data, &config); err != nil {
        fmt.Printf("error extract yaml: %v", err)
        return nil, err
    }
    return
}

// 将给定的 LLMConfig 配置序列化为 YAML 格式并保存到指定的配置文件中
func SaveCfg(config *common.LLMConfig) {
    yamlData, err := yaml.Marshal(&config)
	if err != nil {
		fmt.Printf("error marshaling yaml: %v", err)
		return
	}

	// 写入文件
	err = os.WriteFile(CONFIG_FILE, yamlData, 0644)
	if err != nil {
		fmt.Printf("error writing file: %v", err)
		return
	}
}