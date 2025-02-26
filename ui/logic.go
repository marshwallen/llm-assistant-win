package ui

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"winds-assistant/common"
	"winds-assistant/workers"
	"time"
	"crypto/rand"
	"encoding/hex"
)

// 处理来自 StreamMessage 通道的消息，并在聊天窗口中更新显示内容
func ProcessStream(ctx context.Context, settings common.Settings, widgets common.Widgets, history map[string]interface{}) {
	var contentBuffer bytes.Buffer

	widgets.ChatDisplay.SetText(fmt.Sprintf("%s%s\n", 
		widgets.ChatDisplay.Text,
		common.CHAT_ASSISTANT_INFO,
		))

    err := workers.ChatStream(
		ctx,
		settings,
		widgets,
        // 回调函数
        func(content string, done bool) {
            // 流式实时输出
			widgets.ChatDisplay.SetText(widgets.ChatDisplay.Text + content)
            // 累积内容分块
            contentBuffer.WriteString(content)
            if done {
				widgets.ChatDisplay.SetText(widgets.ChatDisplay.Text + common.CHAT_END)
				history["messages"] = append(
					history["messages"].([]common.LLMMessage), 
					common.LLMMessage{Role: "Assistant", Content: contentBuffer.String()})

				widgets.ChatDisplay.Refresh()
            }
			widgets.ChatScroll.ScrollToBottom()
        },
        // 输入对话历史
        map[string]interface{}{
            "messages":    history["messages"],
        },
	)

	if err != nil {
		log.Fatal(err)
	}
}

// GenerateID 生成一个16位的16进制ID码
func GenerateID() string {
	// 获取当前时间的纳秒级时间戳
	now := time.Now().UnixNano()

	// 将时间戳转换为16进制字符串
	timeHex := fmt.Sprintf("%x", now)

	// 如果时间戳的16进制表示不足16位，则用随机数填充
	if len(timeHex) < 16 {
		randomBytes := make([]byte, 8)
		rand.Read(randomBytes) // 生成随机字节
		randomHex := hex.EncodeToString(randomBytes)
		timeHex += randomHex[:16-len(timeHex)]
	}

	// 确保ID码长度不超过16位
	if len(timeHex) > 16 {
		timeHex = timeHex[:16]
	}

	return timeHex
}
