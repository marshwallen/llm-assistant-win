package common

import (
	"sync"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// 在给定的父窗口中显示错误对话框
func ShowErrorDialog(parent fyne.Window, err error) {
	errLabel := widget.NewLabel(err.Error())
	errLabel.Wrapping = fyne.TextWrapWord
    dialog.ShowCustomConfirm("Error", "Confirm", "Cancel", errLabel, 
		func (confirm bool) {},
		parent,
	)
}

// 自定义滚动容器（带事件触发）
type SmartScroll struct {
    *container.Scroll
    OnScrollToTop    func()
    OnScrollToBottom func()
}

// 构造函数
func NewSmartScroll(content fyne.CanvasObject) *SmartScroll {
    s := &SmartScroll{Scroll: container.NewScroll(content)}
    s.ExtendBaseWidget(s)
    return s
}

// 滚动事件
func (s *SmartScroll) Scrolled(ev *fyne.ScrollEvent) {
    s.Scroll.Scrolled(ev) // 先执行默认滚动

    // 计算是否到达边缘（需在渲染后获取准确尺寸）
    pos := s.Offset
    contentSize := s.Content.Size()
    visibleSize := s.Size()

    // 触发滚动到底部
    if pos.Y+visibleSize.Height >= contentSize.Height-2 {
        if s.OnScrollToBottom != nil {
            s.OnScrollToBottom()
        }
    }

    // 触发滚动到顶部
    if pos.Y <= 2 {
        if s.OnScrollToTop != nil {
            s.OnScrollToTop()
        }
    }
    s.Refresh()
}

// 聊天信息分块
// -----[@@[sss]@@]-----
// sss为窗口显示区域，@@为字符缓冲区（一般来说，左右侧一样）
type ChatChunkProcessor struct {
    textBytes 		[]rune    	// 未完成分块的剩余字符
	textLength 		int			// 分块长度
	SideCacheSize	int			// 单边缓冲区大小
	WindowSize		int			// 显示宽度
	LeftPointer		int			// 左侧缓冲区指针
	RightPointer	int			// 右侧缓冲区指针
}

// 处理新输入的字符串（可多次调用）
func (cp *ChatChunkProcessor) Process(input string) {
    // 字符存储
    cp.textBytes = append(cp.textBytes, []rune(input)...)
    cp.textLength = len(cp.textBytes)    
}

// 清空分块结果
func (cp *ChatChunkProcessor) ClearChunks() {
	cp.textBytes = []rune{}
	cp.textLength = 0
	cp.LeftPointer = 0
	cp.RightPointer = 0
}

// 渲染下一块文字（返回当前块内容）
func (cp *ChatChunkProcessor) RenderNextText() string {
	// 移动右指针
	nextR := cp.RightPointer + cp.SideCacheSize
	if nextR >= cp.textLength {
		cp.RightPointer = cp.textLength-1
	}else{
		cp.RightPointer = nextR
	}

	// 移动左指针
	nextL := cp.RightPointer - 2*cp.SideCacheSize - cp.WindowSize
	if cp.LeftPointer < nextL {
		cp.LeftPointer = nextL
	}
	return string(cp.textBytes[cp.LeftPointer:cp.RightPointer+1])
	// return cp.GetString()
}

// 渲染上一块文字
func (cp *ChatChunkProcessor) RenderPreText() string {
	// 移动左指针
	nextL := cp.LeftPointer - cp.SideCacheSize
	if nextL < 0 {
		cp.LeftPointer = 0
	}else{
		cp.LeftPointer = nextL
	}

	// 移动右指针
	nextR := cp.LeftPointer + 2*cp.SideCacheSize + cp.WindowSize
	if cp.RightPointer > nextR {
		cp.RightPointer = nextR
	}
	return string(cp.textBytes[cp.LeftPointer:cp.RightPointer+1])
	// return cp.GetString()
}

func (cp *ChatChunkProcessor) RenderFinalText() string {
	cp.RightPointer = cp.textLength - 1
	nextL := cp.RightPointer - 2*cp.SideCacheSize - cp.WindowSize
	if cp.LeftPointer < nextL {
		cp.LeftPointer = nextL
	}
	return string(cp.textBytes[cp.LeftPointer:cp.RightPointer+1])	
}

// 优化频繁对字符串进行切片时带来的开销
var bufferPool = sync.Pool{
    New: func() interface{} {
        // 返回指针类型（核心修改）
        buf := make([]rune, 0, 4096)
        return &buf // 返回切片指针
    },
}

// 从 ChatChunkProcessor 中获取指定范围的字符串。
// 该方法使用缓冲池来优化内存分配，确保在获取字符串时不会频繁分配新的内存。
// 返回的字符串是从 cp.textBytes 中根据 LeftPointer 和 RightPointer 指定的范围提取的内容。
func (cp *ChatChunkProcessor) GetString() string {
    // 获取指针类型
    bufferPtr := bufferPool.Get().(*[]rune)
    
    buffer := *bufferPtr
    buffer = buffer[:0] // 清空内容但保留内存

    buffer = append(buffer, cp.textBytes[cp.LeftPointer:cp.RightPointer+1]...)
    result := string(buffer)
    
    *bufferPtr = buffer // 更新指针指向的切片
    bufferPool.Put(bufferPtr)
    
    return result
}