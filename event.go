package tsing

type Event struct {
	Status  int      // HTTP状态码
	File    string   // 源码文件名
	Line    int      // 源码行号
	Trace   []string // 跟踪信息
	Message error    // 消息文本
}

// 事件处理器类型
type EventHandler func(*Context, Event)
