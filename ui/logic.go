package ui

import (
	"bytes"
	"context"
	"fmt"
	"winds-assistant/common"
	"winds-assistant/workers"
	"time"
	"crypto/rand"
	"encoding/hex"
	"container/list"
)

const (
	// 用于维护历史记录的链表（维护最近的 HISTORY_LIST_LENGTH 个记录）
	HISTORY_LIST_LENGTH = 5
)

// 处理来自 StreamMessage 通道的消息，并在聊天窗口中更新显示内容
func ProcessStream(ctx context.Context, settings *common.Settings, widgets common.Widgets, history *list.List) {
	settings.Running = true
	var contentBuffer bytes.Buffer

    err := workers.ChatReqStream(
		ctx,
		settings,
		widgets,
        // 回调函数
        func(content string, done bool) {
            // 流式实时输出
			widgets.ChatChunk.Process(content)
        	widgets.ChatDisplay.SetText(widgets.ChatChunk.RenderFinalText())
			widgets.ChatScroll.ScrollToBottom()
			
            // 累积内容分块
            contentBuffer.WriteString(content)
            if done {
				useTool, midOutput := workers.AgentParser(contentBuffer.String())
				// 如果 useTool 为 True，则 midOutput 为使用工具搜索后的结果
				if useTool{
					widgets.ChatChunk.Process(common.CHAT_AGENT_MID + midOutput)
        			widgets.ChatDisplay.SetText(widgets.ChatChunk.RenderNextText())
					UpdateHistory(history, common.LLMMessage{Role: "midresult", Content: midOutput})
				// 否则，midOutput 为直接返回的 Assistant 的回答
				}else{
					widgets.ChatChunk.Process(common.CHAT_END)
        			widgets.ChatDisplay.SetText(widgets.ChatChunk.RenderNextText())
					
					UpdateHistory(history, common.LLMMessage{Role: "assistant", Content: contentBuffer.String()})
				}
				contentBuffer.Reset() 
				settings.Running = false
            }
			widgets.ChatDisplay.SetText(widgets.ChatChunk.RenderFinalText())
			widgets.ChatScroll.ScrollToBottom()
        },
        // 输入对话历史
        GenerateHistoryMessage(history, settings.SysPrompt),
	)

	if err != nil {
		settings.Running = false
		common.ShowErrorDialog(widgets.Window, err)
	}
}

// 两次聊天流以调取工具
func ProcessStreamWithTools(ctx context.Context, settings *common.Settings, widgets common.Widgets, history *list.List) {
	ProcessStream(ctx, settings, widgets, history)
	lastMessage := history.Back().Value.(common.LLMMessage)

	if lastMessage.Role == "midresult" {
		history.Remove(history.Back())
		UpdateHistory(history, common.LLMMessage{Role: "user", Content: workers.USER_PROMPT_WITH_TOOLS + lastMessage.Content})
		ProcessStream(ctx, settings, widgets, history)
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

// 滚动维护 History 链表（不包括System）
func UpdateHistory(history *list.List, message common.LLMMessage) {
	// 使用滑动窗口截断历史记录
	if HISTORY_LIST_LENGTH <= 0 {
		return
	}

	if message.Role != "System" {
		if history.Len() < HISTORY_LIST_LENGTH {
			history.PushBack(message)
		}else{
			history.Remove(history.Front())
			history.PushBack(message)
		}
	}
}

// 生成带或者不带System的历史记录，以便传入请求体
func GenerateHistoryMessage(history *list.List, systemPrompt string) map[string]interface{} {
	var historyMessage []common.LLMMessage
	historyMessage = append(historyMessage, common.LLMMessage{Role: "system", Content: systemPrompt})
	for e := history.Front(); e != nil; e = e.Next() {
		historyMessage = append(historyMessage, e.Value.(common.LLMMessage))
	}
	return map[string]interface{}{
		"messages": historyMessage,
	}
}