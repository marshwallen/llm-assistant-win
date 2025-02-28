package workers

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "winds-assistant/common"
    "fyne.io/fyne/v2"
    "context"
)

// 流式调用OpenAI接口的核心函数
func ChatReqStream(ctx context.Context, settings *common.Settings, widgets common.Widgets, streamCallback func(content string, done bool),
    options ...map[string]interface{}) error {

    // 构建请求体
    body := map[string]interface{}{
        "model":  settings.BackendCfg.Model,
        "stream": true,
    }

    // 合并可选参数
    if len(options) > 0 {
        for k, v := range options[0] {
            body[k] = v
        }
    }

    // 创建HTTP请求
    reqBody, _ := json.Marshal(body)

    var req *http.Request
    if settings.BackendName == "ollama" {
        req, _ = http.NewRequestWithContext(ctx, "POST", settings.BackendCfg.BaseURL + "/api/chat", bytes.NewBuffer(reqBody))
    }else{
        req, _ = http.NewRequestWithContext(ctx, "POST", settings.BackendCfg.BaseURL, bytes.NewBuffer(reqBody))
    }
    
    req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer " + settings.BackendCfg.APIKey)

    // 发送请求
    client := &http.Client{Timeout: 0} // 无超时限制
    resp, err := client.Do(req)
    if err != nil {
        common.ShowErrorDialog(widgets.Window, fmt.Errorf("request failed: %v", err))
        return err
    }
    defer resp.Body.Close()

    // 检查状态码
    if resp.StatusCode != http.StatusOK {
        common.ShowErrorDialog(widgets.Window, fmt.Errorf("error code: %d, req body: %+v, resp body: %+v", resp.StatusCode, string(reqBody), resp.Body))
        return err
    }

    var leftover []byte
    buf := make([]byte, 4096)
    for {
        select {
        case <-ctx.Done():
            settings.Running = false
            widgets.ChatChunk.Process(common.CHAT_TERMINATE)
            widgets.ChatDisplay.SetText(widgets.ChatChunk.RenderFinalText())
            widgets.ChatScroll.ScrollToBottom()
            return nil
        default:
        }

        n, err := resp.Body.Read(buf)
        if err != nil && err != io.EOF {
            common.ShowErrorDialog(widgets.Window, fmt.Errorf("read failed: %v", err))
            return err
        }

        eof := err == io.EOF
        if n > 0 {
            leftover = append(leftover, buf[:n]...)
        }

        // 处理所有完整行
        for {
            newlineIdx := bytes.IndexByte(leftover, '\n')
            if newlineIdx == -1 {
                break
            }

            line := leftover[:newlineIdx]
            leftover = leftover[newlineIdx+1:]

            content, done := ParserFuncRegister[settings.BackendName](line)
            streamCallback(content, done)
            if done {
                return nil
            }
        }

        // 处理EOF时的剩余数据
        if eof {
            if len(leftover) > 0 {
                content, done := ParserFuncRegister[settings.BackendName](leftover)
                streamCallback(content, done)
                if done {
                    return nil
                }
            }
            break
        }
    }
    return nil
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