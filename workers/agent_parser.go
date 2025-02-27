package workers

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

const (
	// DEFAULT
	SYSTEM_PROMPT_DEFAULT = `你是一个人工智能助手。`

	// 尝试解析用户需求以调取工具
	SYSTEM_PROMPT_WITH_TOOLS_BASE = `
	你是一个 Windows 系统上的人工智能助手。你需要分析用户的输入，然后以规定的格式返回将使用的工具。获取到信息后，你可以回答用户的问题。
	注意: 在用户使用工具获取到资料后, 你要回答用户之前提出的问题, 不能再返回json格式的数据, 以免再次触发工具调用。
	注意: 如果你确信用户不想使用工具获取信息, 可以根据用户需求随意返回任何内容, 无需遵从任何规范格式。\n`

	// 使用工具后的用户提示
	USER_PROMPT_WITH_TOOLS= `现在，现在可以回答我的问题了，通过工具获取的上述问题的资料如下:\n`
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
		if q, ok := v.(map[string]interface{}); ok {
			f, exists := ToolsFuncRegister[k]
			if !exists {
				output += fmt.Sprintf("Invalid tool %s\n", k)
				continue
			}else{
				output += f(q)
			}
		} else {
			output += fmt.Sprintf("Invalid type for %s\n", k)
		}
	}
	return
}