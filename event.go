package goevent

// EventID 事件ID
// 用于唯一标识一个事
type EventID[ET, EV comparable] struct {
	Type ET
	Val  EV
}

// Event 产生的事件
type Event[ET, EV comparable] struct {
	eventID   EventID[ET, EV] // 事件ID
	param     interface{}     // 参数
	generator interface{}     // The generator of the event
}

func (e *Event[ET, EV]) EventID() EventID[ET, EV] { return e.eventID }

func (e *Event[ET, EV]) Param() interface{} { return e.param }

func (e *Event[ET, EV]) Generator() interface{} { return e.generator }
