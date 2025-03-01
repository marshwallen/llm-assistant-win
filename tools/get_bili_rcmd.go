package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"os"
	"gopkg.in/yaml.v3"
	"path/filepath"
)

// B站响应体定义
type BiliResponse struct {
	Code    int    					`json:"code"`
	Message string 					`json:"message"`
	TTL     int    					`json:"ttl"`
	Data    Data   					`json:"data"`
}

type Data struct {
	Item []VideoItem 				`json:"item"`
}

type VideoItem struct {
	Bvid    string     				`json:"bvid"`
	Title   string     				`json:"title"`
	Duration int       				`json:"duration"`
	Pubdate int64      				`json:"pubdate"`
	Owner   OwnerInfo  				`json:"owner"`
	Stat    VideoStat  				`json:"stat"`
	RcmdReason RcmdReasonInfo 		`json:"rcmd_reason"`
}

type OwnerInfo struct {
	Mid  int64  					`json:"mid"`
	Name string 					`json:"name"`
	Face string 					`json:"face"`
}

type VideoStat struct {
	View    int 					`json:"view"`
	Like    int 					`json:"like"`
	Danmaku int 					`json:"danmaku"`
}

type RcmdReasonInfo struct {
	Content 		string 			`json:"content"`
	ReasonType 		int 			`json:"reason_type"`
}

// 工具配置
type BiliToolCfg struct {
	Cookie 			string			`json:"cookie"`
}

func GetBiliRcmd(ctx context.Context, cookie string) (biliresp *BiliResponse, err error) {
	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 10 * time.Second, // 设置超时时间
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.bilibili.com/x/web-interface/index/top/feed/rcmd", nil)
	if err != nil {
		return nil, fmt.Errorf("create req failed: %w", err)
	}

	// 请求头
	req.Header = http.Header{
		"User-Agent": []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"},
		"Referer":    []string{"https://www.bilibili.com/"},
		"Cookie":     []string{cookie},
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("req failed: %w", err)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error code: %d", resp.StatusCode)
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read resp failed: %w", err)
	}

	// 解析JSON
	var result BiliResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshwal JSON failed: %w", err)
	}

	// 检查业务码
	if result.Code != 0 {
		return nil, fmt.Errorf("wrong bilibili api: %d - %s", 
			result.Code, result.Message)
	}
	return &result, nil
}

func GetBiliRcmdStr(enable_cookie bool, rounds int) (r string) {
	// 读取工具配置文件
    data, err := os.ReadFile(filepath.Clean("config/agent_get_bili_rcmd.yaml"))
    if err != nil {
        fmt.Printf("error read yaml: %v", err)
        return
    }

    // 解析 Yaml
	var config BiliToolCfg
    if err := yaml.Unmarshal(data, &config); err != nil {
        fmt.Printf("error extract yaml: %v", err)
        return
    }

	// 从结构体中提取关键字，并拼接成字符串
	// rounds 表示获取几轮推荐
	for i := 0; i < rounds; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
		defer cancel()

		var biliresp *BiliResponse
		var err error
		if enable_cookie {
			biliresp, err = GetBiliRcmd(ctx, config.Cookie)
		}else{
			biliresp, err = GetBiliRcmd(ctx, "")
		}
		
		if err != nil {
			return
		}

		for _, item := range biliresp.Data.Item {
			bvid := item.Bvid												// 视频BV号
			title := item.Title												// 视频标题
			duration := item.Duration										// 视频时长(s)
			pubdate := time.Unix(item.Pubdate, 0).Format(time.RFC3339Nano)	// 发布时间
			owner := item.Owner.Name										// UP主名称
			uid := item.Owner.Mid											// UP主 UID
			view := item.Stat.View											// 观看数
			like := item.Stat.Like											// 点赞数
			danmaku := item.Stat.Danmaku									// 弹幕数
			rcmdReason := item.RcmdReason.Content							// 推荐理由

			r += fmt.Sprintf("[%s] %s | 发布时间: %s | UP主: %s(%d) | 视频时长(秒): %v | 观看数: %d | 点赞数: %d | 弹幕数: %d | 推荐理由: %s\n",
				bvid, title, pubdate, owner, uid, duration, view, like, danmaku, rcmdReason)
		}
		time.Sleep(500 * time.Millisecond)
	}
	return
}