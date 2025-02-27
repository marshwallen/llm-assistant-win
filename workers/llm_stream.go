package workers

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
	"bufio"
    "winds-assistant/common"
    "fyne.io/fyne/v2"
    "context"
)

// 流式调用OpenAI接口的核心函数
func ChatStream(ctx context.Context, settings common.Settings, widgets common.Widgets, streamCallback func(content string, done bool),
    options ...map[string]interface{}) error {

    // 构建请求体
    body := map[string]interface{}{
        "model":  settings.Model,
    }

    // 合并可选参数
    if len(options) > 0 {
        for k, v := range options[0] {
            body[k] = v
        }
    }

    // 创建HTTP请求
    reqBody, _ := json.Marshal(body)
    req, _ := http.NewRequest("POST", settings.URL+"/api/chat", bytes.NewBuffer(reqBody))
    req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer " + settings.API_KEY)

    // 发送请求
    client := &http.Client{Timeout: 0} // 无超时限制
    resp, err := client.Do(req)
    if err != nil {
        common.ShowErrorDialog(widgets.Window, err)
        return fmt.Errorf("request failed: %v", err)
    }
    defer resp.Body.Close()

    // 检查状态码
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("error code: %d, req body: %+v", resp.StatusCode, body)
    }

    // 流式处理响应
    reader := bufio.NewReader(resp.Body)
    for {
        select {
            case <-ctx.Done():
                widgets.ChatDisplay.SetText(widgets.ChatDisplay.Text + common.CHAT_TERMINATE)
                widgets.ChatScroll.ScrollToBottom()
                return nil
            default:    
        }

        line, err := reader.ReadBytes('\n')
        if err != nil {
            if err == io.EOF {
                common.ShowErrorDialog(widgets.Window, err)
                streamCallback("", true) // 通知结束
                return nil
            }
            return fmt.Errorf("read failed: %v", err)
        }

        var ch common.Chan
        if err := json.Unmarshal(line, &ch); err != nil {
            continue // 跳过无效数据
        }
        // 回调处理
        streamCallback(ch.Message.Content, ch.Done)

        // 结束标志
        if ch.Done {
            return nil
        }
    }
}

// 获取模型列表
func GetModelList(url string, window fyne.Window) (modelList []string){
    // 发送 GET 请求
    resp, err := http.Get(url+"/api/tags")
    if err != nil {
        common.ShowErrorDialog(window, err)
        fmt.Println("req failed:", err)
        return
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        common.ShowErrorDialog(window, err)
        fmt.Println("read resp failed:", err)
    }

    var modelListResponse common.ModelListResponse
    err = json.Unmarshal(body, &modelListResponse)
    if err != nil {
        common.ShowErrorDialog(window, err)
        fmt.Println("unmarshal json failed:", err)
    }

    for _, model := range modelListResponse.Models {
        modelList = append(modelList, model.Name)
    }
    return
}