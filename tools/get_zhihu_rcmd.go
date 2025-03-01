package tools

import (
	"context"
	"fmt"
	"net/http"
	"github.com/PuerkitoBio/goquery"
	"time"
	"strconv"
	"encoding/json"
	"strings"
	"path/filepath"
	"gopkg.in/yaml.v3"
	"os"
)

// 定义一些响应体
type ZhihuResponse struct {
    Title       	string
    Author      	string
    Description 	string
	Link			string
    Upvotes     	int
    Comments    	int
    ItemID      	string
}

type DataZop struct {
	AuthorName 		string 	`json:"authorName"`
	ItemId		 	string 	`json:"itemId"`
	Title 			string 	`json:"title"`
	Type			string 	`json:"type"`
}

// 工具配置
type ZhihuToolCfg struct {
	Cookie 			string			`json:"cookie"`
}

// 根据 Cookie 获取知乎首页推荐
// 采用解析 HTML 的方式
func GetZhihuRcmd(ctx context.Context, cookie string) (items []ZhihuResponse, err error) {
	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 10 * time.Second, // 设置超时时间
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.zhihu.com", nil)
	if err != nil {
		return items, fmt.Errorf("create req failed: %w", err)
	}

	// 请求头
	req.Header = http.Header{
		"User-Agent": []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"},
		"Cookie":     []string{cookie},
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return items, fmt.Errorf("req failed: %w", err)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return items, fmt.Errorf("error code: %d", resp.StatusCode)
	}

	// 创建 goquery 文档
    doc, err := goquery.NewDocumentFromReader(resp.Body)
    if err != nil {
        return items, fmt.Errorf("create goquery doc failed: %w", err)
    }

    // 遍历推荐内容容器
    doc.Find(".ListShortcut .Topstory-recommend .TopstoryItem").Each(func(i int, s *goquery.Selection) {
        // 提取基础数据
        item := ZhihuResponse{}

        if link, exists := s.Find(".ContentItem-title a").Attr("href"); exists {
            item.Link = "https://" + strings.TrimPrefix(link, "//")
        }

        // 从data-zop属性提取结构化数据
		var datazop DataZop
        if zop, exists := s.Find(".ContentItem").Attr("data-zop"); exists {
			err := json.Unmarshal([]byte(zop), &datazop)
			if err != nil {
				return
			}
			item.Title = datazop.Title
			item.Author = datazop.AuthorName
            item.ItemID = datazop.ItemId
        }

        // 从meta标签获取数值
        s.Find("meta[itemprop]").Each(func(_ int, meta *goquery.Selection) {
            switch meta.AttrOr("itemprop", "") {
            case "upvoteCount":
                item.Upvotes, _ = strconv.Atoi(meta.AttrOr("content", "0"))
            case "commentCount":
                item.Comments, _ = strconv.Atoi(meta.AttrOr("content", "0"))
            }
        })

        // 描述文本处理
        descSel := s.Find(".RichText.ztext")
        if descSel.Length() > 0 {
            item.Description = descSel.First().Text()
        }

        items = append(items, item)
    })
	return
}

// 根据提供的 cookie 和轮数获取知乎推荐内容的字符串
func GetZhihuRcmdStr(rounds int) (r string){
	// 读取工具配置文件
    data, err := os.ReadFile(filepath.Clean("config/agent_get_zhihu_rcmd.yaml"))
    if err != nil {
        fmt.Printf("error read yaml: %v", err)
        return
    }

    // 解析 Yaml
	var config ZhihuToolCfg
    if err := yaml.Unmarshal(data, &config); err != nil {
        fmt.Printf("error extract yaml: %v", err)
        return
    }

	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()
	for i := 0; i < rounds; i++ {
		zhihuresp, err := GetZhihuRcmd(ctx, config.Cookie)
		if err != nil {
			return
		}

		for _, item := range zhihuresp {
			r += fmt.Sprintf("[%s] %s | 作者: %s | 链接: %s | 描述: %s | 赞同: %v | 评论: %v\n",
				item.ItemID, item.Title, item.Author, item.Link, item.Description, item.Upvotes, item.Comments)
		}
		time.Sleep(500 * time.Millisecond)
	}
	return
}
