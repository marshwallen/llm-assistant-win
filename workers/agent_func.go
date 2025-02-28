package workers

import (
	"winds-assistant/tools"
)

// ** Register Agent Tools
var ToolsFuncRegister = map[string]func(map[string]interface{}) (string){
	"get_win_event": getWinEvent,
	"get_file_tree": getFileTree,
	"get_sys_health": getSysHealth,
	"get_sys_process": getSysProcess,
}

// ** Register Agent Tools Prompt
var ToolsPromptRegister = []string{
	GET_WIN_EVENT_PROMPT,
	GET_FILE_TREE_PROMPT,
	GET_SYS_HEALTH_PROMPT,
	GET_SYS_PROCESS_PROMPT,
}

// 在这里写 Agent Tools 的函数入口
// return type 必须为 string

// 查询 Windows 事件日志并返回结果
const GET_WIN_EVENT_PROMPT = `
工具 <get_win_event> 使用规则：
1. 如果用户提到了 <分析日志> 等类似的需求，你可以使用 <get_win_event> 工具来获取日志。
2. 此外，你还要解析用户需要分析的日志类型(Application, Security, System), 分析天数(StartTime, 为正数(默认1), 表示往前分析多少天)和最大事件数(MaxEvents, 默认50)。
3. 最后, 只返回如下类似的json内容, 除此之外不要说任何其他内容, 不要有多余的符号如 Markdown 代码块标识符, 无效换行和空白等:
{
	"tools": {
		"get_win_event": {
			"logName": "Application",
			"startTime": 1,
			"maxEvents": 50
		}
	}            
}`
func getWinEvent(q map[string]interface{}) string{
	logName, _ := q["logName"].(string)
    _s, _ := q["startTime"].(float64)
    _m, _ := q["maxEvents"].(float64)
    startTime := int(_s)
    maxEvents := int(_m)

	_o, _ := tools.QueryEvents(logName, startTime, maxEvents)
	output := "<get_win_event> 返回结果：" + _o
	return output
}

const GET_FILE_TREE_PROMPT = `
工具 <get_file_tree> 使用规则：
1. 如果用户提到了 <分析文件> 等类似的需求，你可以使用 <get_file_tree> 工具来获取文件树。
2. 此外，你还要解析用户想分析的盘符(如 C:/)。默认是 C:/。如有多个，请用字符串列表的形式表示。
3. 最后, 只返回如下类似的json内容, 除此之外不要说任何其他内容, 不要有多余的符号如 Markdown 代码块标识符, 无效换行和空白等:
{
	"tools": {
		"get_file_tree": {
			"disk": ["A:/", "B:/"]
		}
	}
}`
// 根据提供的磁盘列表获取文件树结构
func getFileTree(q map[string]interface{}) string{
	diskList, _ := q["disk"].([]interface{})

	output := "<get_file_tree> 返回结果："
	for _, d := range diskList {
		disk, _ := d.(string)
		_o, _ := tools.GetFileTree(disk, 3, 10, false)
		output += _o
	}
	return output
}

const GET_SYS_HEALTH_PROMPT = `
工具 <get_sys_health> 使用规则：
1. 如果用户提到了 <分析系统、硬件监控、CPU、GPU、内存、硬盘> 等类似的需求，你可以使用 <get_sys_health> 工具来获取系统信息。
2. 此外，你还要解析用户想分析的时间范围(Minutes, 为正数, 表示往前分析多少分钟)。默认是1, 表示分析最近一分钟。
3. 用户对每个指标返回一个列表，分别代表每个指标在每个时间间隔的值。
4. 你需要结合多个指标对系统状态进行分析。此外，因为这些是时间序列，你需要额外地进行一些数学上的分析。
5. 最后, 只返回如下类似的json内容, 除此之外不要说任何其他内容, 不要有多余的符号如 Markdown 代码块标识符, 无效换行和空白等:
{
	"tools": {
		"get_sys_health": {                          
			"minutes": 1
		}
	}
}`
// 根据提供的参数获取系统健康数据
func getSysHealth(q map[string]interface{}) string{
	_m, _ := q["minutes"].(float64)
	minutes := int(_m)
	_o, _ := tools.GetSysHealthData(minutes)
	output := "<get_sys_health> 返回结果：" + _o
	return output
}

const GET_SYS_PROCESS_PROMPT = `
工具 <get_sys_process> 使用规则：
1. 如果用户提到了 <分析系统进程、查看运行情况> 等类似的需求，你可以使用 <get_sys_process> 工具来获取系统进程信息。
2. 你需要结合多个指标对系统状态进行分析, 如进程的CPU, 内存 ,执行路径 ,启动参数 ,运行状态 ,线程/句柄数 ,IO 统计等详细信息, 查找潜在问题。
3. 最后, 只返回如下类似的json内容, 除此之外不要说任何其他内容, 不要有多余的符号如 Markdown 代码块标识符, 无效换行和空白等:
{
	"tools": {
		"get_sys_process": {                          
			"enable": true
		}
	}
}
`
// 根据传入的参数判断是否启用获取系统进程信息
func getSysProcess(q map[string]interface{}) string{
	enable, _ := q["enable"].(bool)
	if enable {
		// 获取系统进程信息
		_o := tools.GetSysProcessStr()
		output := "<get_sys_process> 返回结果：" + _o
		return output
	}
	return ""
}