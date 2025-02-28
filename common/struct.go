package common

import (
	"context"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
)

type LLMMessage struct {
    Role 		   string `json:"role"`
    Content 	   string `json:"content"`
}

type Model struct {
    Name           string `json:"name"`
    ModifiedAt     string `json:"modified_at"`
    Size           int64  `json:"size"`
    Digest         string `json:"digest"`
}
// 模型列表
type ModelListResponse struct {
    Models []Model `json:"models"`
}

// 用户设置
type Settings struct {
    BackendName    string                  // 后端名称
    BackendCfg     BackendConfig           // 后端配置
    ModelList      []string                // 模型列表
    CancelFunc     context.CancelFunc      // 上下文取消函数
    DialogID       string                  // 对话 ID
    EnableAgent    bool                    // 是否启用Agent调用系统能力
    SysPrompt      string                  // 系统 Prompt
    Running        bool                    // 是否正在对话
}

// 配置文件解析
type LLMConfig struct {
    Backend        map[string]BackendConfig `yaml:"backend"`     // 后端配置
    Default        string                   `yaml:"default_backend"`     // 默认后端
}

type BackendConfig struct {
    BaseURL        string `yaml:"base_url"`    // LLM后端地址
    APIKey         string `yaml:"api_key"`     // API Key
    Model          string `yaml:"model"`       // 模型名称
}

type GPUInfoStat struct {
	Index		   uint64  `json:"index"`		 // GPU序号
	Name		   string  `json:"name"`         // GPU名称
    Utilization    float64 `json:"gpu_util"`     // GPU利用率(%) 
    MemUsed        uint64  `json:"mem_used"`     // 显存使用量(MB)
    MemTotal       uint64  `json:"mem_total"`    // 显存总量(MB)
    CoreClock      uint64  `json:"core_clock"`   // 核心频率(MHz)
    MemClock       uint64  `json:"mem_clock"`    // 显存频率(MHz)
    Temperature    uint64  `json:"temp"`         // 温度(℃)
    PowerDraw      float64 `json:"power"`        // 实时功耗(W)
    Vendor         string  `json:"vendor"`       // 厂商(NVIDIA)
}

type CPUInfoStat struct{
    Base           cpu.InfoStat                  // CPU信息
    Percent        float64                       // CPU利用率
}

type MetricData struct {
	CPU            CPUInfoStat
	Mem            *mem.VirtualMemoryStat
	Disk           *disk.UsageStat
	GPU            []GPUInfoStat
	Time           int64
}

type Widgets struct {
    Window         fyne.Window
	MainSplit 	   *container.Split
	ChatDisplay    *widget.Label
	ChatScroll 	   *SmartScroll
	InputEntry 	   *widget.Entry
    ChatChunk      *ChatChunkProcessor
}
