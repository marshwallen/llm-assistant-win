package workers

import (
	"encoding/json"
	"regexp"
	"strings"
	"winds-assistant/tools"
)

const (
	SYSTEM_PROMPT_DEFAULT = `你是一个人工智能助手。`
	// 尝试解析用户需求以调取工具
	SYSTEM_PROMPT_WITH_TOOLS = `
		你是一个 Windows 系统上的人工智能助手。你需要分析用户的输入，然后以规定的格式返回将使用的工具。获取到信息后，你可以回答用户的问题。

		1. 
		1.1. 如果用户提到了 <分析日志> 等类似的需求，你可以使用 <get_win_event> 工具来获取日志。
		1.2. 此外，你还要解析用户需要分析的日志类型(Application, Security, System), 分析天数(StartTime, 为正数(默认1), 表示往前分析多少天)和最大事件数(MaxEvents, 默认50)。
		1.3. 最后, 必需只返回一个类似的json内容, 除此之外不要说任何其他内容:
		{
			"tools": {
				"get_win_event": {
					"logName": "Application",
					"startTime": 1,
					"maxEvents": 50
				}
			}            
		}
                                                           
		2. 
		2.1. 如果用户提到了 <分析文件> 等类似的需求，你可以使用 <get_file_tree> 工具来获取文件树。
		2.2. 此外，你还要解析用户想分析的盘符(如 C:/)。默认是 C:/。
		2.3. 最后, 必需只返回一个类似的json内容, 除此之外不要说任何其他内容:
		{
			"tools": {
				"get_file_tree": {
					"disk": ["C:/", "D:/"]
				}
			}
		}

		3. 
		3.1. 如果用户提到了 <分析系统、硬件监控、CPU、GPU、内存、硬盘> 等类似的需求，你可以使用 <get_sys_health> 工具来获取系统信息。
		3.2. 此外，你还要解析用户想分析的时间范围(Minutes, 为正数, 表示往前分析多少分钟)。默认是1, 表示分析最近一分钟。
		3.3. 用户对每个指标返回一个列表，分别代表每个指标在每个时间间隔的值。
		3.4. 最后, 必需只返回一个类似的json内容, 除此之外不要说任何其他内容:
		{
			"tools": {
				"get_sys_health": {                          
					"minutes": 1
				}
			}
		}

		4. 如果用户同时用到了以上工具中的几种, 你需要将每种工具的内容添加到 "tools" 字典中, 并返回以下类似的json内容:
		{
			"tools": {
				"get_sys_health": {...},
				"get_win_event": {...},
			}
		}

		5. 如果你确信用户不想获取以上的各种信息，你可以根据用户需求随意返回任何内容,无需遵从任何规范格式。
		`
	USER_PROMPT_LAST = `现在，你已经查到了上述问题的资料, 现在可以回答我的问题了:`
)

// 解析模型返回
func AgentParser(rawOutput string) (useTool bool, output string){
	useTool = true
	index := strings.Index(rawOutput, "</think>")
	if index != -1 {
		rawOutput = rawOutput[index + len("</think>"):]
	}
	re := regexp.MustCompile("(?m)^\\s*```[\\w\\s]*\\n|\\n\\s*```\\s*$")
	rawOutput = re.ReplaceAllString(rawOutput, "\n")
	rawOutput = strings.TrimPrefix(rawOutput, "\n")
	rawOutput = strings.TrimSuffix(rawOutput, "\n")
	rawOutput = strings.TrimSpace(rawOutput)

	// 解析 JSON 到 map
	var data map[string]interface{}
	err := json.Unmarshal([]byte(rawOutput), &data)
	if err != nil {
		useTool = false
		output = rawOutput
		return
	}

	// 逐个调用工具
	toolsMap, ok := data["tools"].(map[string]interface{})
	if !ok {
		useTool = false
		output = rawOutput
		return
	}
	for k, v := range toolsMap {
		switch k {
		// 获取 Windows 事件查看器日志
		case "get_win_event":
			if q, ok := v.(map[string]interface{}); ok {
				_o, _ := tools.QueryEvents(q)
				output += _o
			} else {
				output = "Invalid type for get_win_event"
			}
		// 获取文件树
		case "get_file_tree":
			if q, ok := v.(map[string]interface{}); ok {
				diskList, _ := q["disk"].([]interface{})
				for _, disk := range diskList {
					_disk, _ := disk.(string)
					_o, _ := tools.GetFileTree(_disk, 3, 10, false)
					output += _o
				}
			} else {
				output = "Invalid type for get_file_tree"
			}
		case "get_sys_health":
			if q, ok := v.(map[string]interface{}); ok {
				_o, _ := tools.GetSysHealthData(q)
				output += _o
			} else {
				output = "Invalid type for get_sys_health"
			}
		default:
			useTool = false
			output = "Invalid tool name"
		}
	}

	return
}