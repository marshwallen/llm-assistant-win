package workers

import (
	"encoding/json"
	"fmt"
	"time"
	"winds-assistant/common"
	"bytes"
)

// 解析各API返回的数据
// 函数注册
var ParserFuncRegister = map[string]func([]byte) (string, bool){
	"ollama": ParserOllamaResp,
	"qwen":   ParserQwenResp,
	"volcengine": ParserVolcengineResp,
}

// ** Ollama 解析块 **
type RespOllama struct {
    Model     string    			`json:"model"`
    CreatedAt time.Time 			`json:"created_at,omitempty"` 			// 使用RFC3339Nano时间格式
    Message   common.LLMMessage   	`json:"message"`
    Done      bool      			`json:"done"`
}

func ParserOllamaResp(resp []byte) (content string, done bool) {
	// 判断空值
	if len(resp) == 0 {
		return content, false
	}

	var ch RespOllama
	if err := json.Unmarshal(resp, &ch); err != nil {
		fmt.Printf("json.Unmarshal error: %+v\n", err)
		return content, true
	}
	content = ch.Message.Content
	done = ch.Done
	return
}

// ** 通义千问(Qwen)解析块 **
type RespQwen struct {
	Choices []struct {	
        Delta struct {
            Content string `json:"content"`
        } `json:"delta"`
    } `json:"choices"`
}

func ParserQwenResp(resp []byte) (content string, done bool) {
	// 处理前缀
	resp = bytes.TrimPrefix(resp, []byte("data: "))

	// 判断空值
	if len(resp) == 0 {
		return content, false
	}

	// 处理流结束标记
	if string(resp) == "[DONE]" {
		return content, true
	}

	// 解析数据流
	var ch RespQwen
	if err := json.Unmarshal(resp, &ch); err != nil {
		fmt.Printf("json.Unmarshal error: %+v\n", err)
		return content, true
	}
	content = ch.Choices[0].Delta.Content
	done = false

	return
}

// ** 字节火山引擎 **
type RespVolcengine struct {
    Choices []struct {
        Delta struct {
            Content string `json:"content"`
        } `json:"delta"`
    } `json:"choices"`
}

func ParserVolcengineResp(resp []byte) (content string, done bool) {
	// 处理前缀
	resp = bytes.TrimPrefix(resp, []byte("data: "))

	// 判断空值
	if len(resp) == 0 {
		return content, false
	}

	// 处理流结束标记
	if string(resp) == "[DONE]" {
		return content, true
	}

	// 解析数据流
	var ch RespVolcengine
	if err := json.Unmarshal(resp, &ch); err != nil {
		fmt.Printf("json.Unmarshal error: %+v\n", err)
		return content, true
	}
	content = ch.Choices[0].Delta.Content
	done = false

	return
}	