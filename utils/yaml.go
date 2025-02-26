package utils

import (
    "os"
    "path/filepath"
    "gopkg.in/yaml.v3"
    "winds-assistant/common"
)

// 从指定路径加载配置文件并解析为 LLMConfig 结构体。
func LoadCfg() (*common.LLMConfig, error) {
    // 读取文件内容
    data, err := os.ReadFile(filepath.Clean("config/llm_settings.yaml"))
    if err != nil {
        return nil, err
    }

    // 解析YAML
    var cfg common.LLMConfig
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }

    return &cfg, nil
}