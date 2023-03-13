package goevent

import (
	"container/list"
	"errors"
)

var ErrUnregister = errors.New("unregister")

// EventHandler 一个事件处理起需要实现的接口
type EventHandler[ET, EV comparable] interface {
	HandleEvent(Event[ET, EV]) error
}

// EventCallback 事件回调
// 用来支持将任何函数注册为事件处理器
type EventCallback[ET, EV comparable] func(Event[ET, EV]) error

// CallbackEventHandler 将EventCallback封装为Handler
type CallbackEventHandler[ET comparable, EV comparable] struct {
	cb EventCallback[ET, EV]
}

func newCallbackEventHandler[ET, EV comparable](cb EventCallback[ET, EV]) *CallbackEventHandler[ET, EV] {
	return &CallbackEventHandler[ET, EV]{cb: cb}
}

func (h *CallbackEventHandler[ET, EV]) HandleEvent(evt Event[ET, EV]) error {
	return h.cb(evt)
}

// eventHandler 事件处理器entry
type eventHandler[ET, EV, HK comparable] struct {
	key  HK                   // key
	h    EventHandler[ET, EV] // 事件处理起
	once bool                 // 是否只处理一次
}

func newEventHandler[ET, EV, HK comparable](key HK, h EventHandler[ET, EV], once bool) *eventHandler[ET, EV, HK] {
	if h == nil {
		panic("nil EventHandler")
	}
	return &eventHandler[ET, EV, HK]{
		key:  key,
		h:    h,
		once: once,
	}
}

func (h *eventHandler[ET, EV, HK]) call(evt Event[ET, EV]) error {
	return h.h.HandleEvent(evt)
}

// eventHandlers 事件处理器容器
// 每一个独立的事件，都有与之对应的处理器容器来维护相关的处理器
type eventHandlers[ET, EV, HK comparable] struct {
	calling        bool
	handlerList    *list.List
	handlerMap     map[HK]*list.Element
	pendingRemList *list.List
}

func newEventHandlers[ET, EV, HK comparable]() *eventHandlers[ET, EV, HK] {
	return &eventHandlers[ET, EV, HK]{
		handlerList: list.New(),
		handlerMap:  map[HK]*list.Element{},
	}
}

func (hs *eventHandlers[ET, EV, HK]) add(key HK, h EventHandler[ET, EV], once bool) {
	if elem, ok := hs.handlerMap[key]; ok {
		hEntry := elem.Value.(*eventHandler[ET, EV, HK])
		hEntry.h = h
		hEntry.once = once
	} else {
		hEntry := newEventHandler[ET, EV, HK](key, h, once)
		elem = hs.handlerList.PushBack(hEntry)
		hs.handlerMap[key] = elem
	}
}

func (hs *eventHandlers[ET, EV, HK]) rem(key HK) {
	elem, ok := hs.handlerMap[key]
	if !ok {
		return
	}
	if hs.calling {
		if hs.pendingRemList == nil {
			hs.pendingRemList = list.New()
		}
		hs.pendingRemList.PushBack(key)
	} else {
		hs.handlerList.Remove(elem)
		delete(hs.handlerMap, key)
	}
}

func (hs *eventHandlers[ET, EV, HK]) empty() bool {
	return hs.handlerList.Len() == 0
}

func (hs *eventHandlers[ET, EV, HK]) call(evtID EventID[ET, EV], generator interface{}, param interface{}) error {
	event := Event[ET, EV]{
		eventID:   evtID,
		generator: generator,
		param:     param,
	}

	hs.calling = true
	elem := hs.handlerList.Front()
	for elem != nil {
		handler := elem.Value.(*eventHandler[ET, EV, HK])
		err := handler.call(event)
		if err != nil && err != ErrUnregister {
			return err
		}
		next := elem.Next()
		if handler.once || err == ErrUnregister {
			hs.handlerList.Remove(elem)
			delete(hs.handlerMap, handler.key)
		}
		elem = next
	}
	hs.calling = false

	if hs.pendingRemList != nil {
		elem := hs.pendingRemList.Front()
		for elem != nil {
			key := elem.Value.(HK)
			hs.rem(key)
			elem = elem.Next()
		}
		hs.pendingRemList.Init()
		hs.pendingRemList = nil
	}

	return nil
}

func (hs *eventHandlers[ET, EV, HK]) clear() {
	hs.calling = false
	hs.handlerList.Init()
	hs.handlerMap = map[HK]*list.Element{}
	if hs.pendingRemList != nil {
		hs.pendingRemList.Init()
		hs.pendingRemList = nil
	}
}

type eventTypeHandlers[ET, EV, HK comparable] struct {
	valHandlers map[EV]*eventHandlers[ET, EV, HK]
	*eventHandlers[ET, EV, HK]
}

func newEventTypeHandlers[ET, EV, HK comparable]() *eventTypeHandlers[ET, EV, HK] {
	return &eventTypeHandlers[ET, EV, HK]{
		valHandlers:   nil,
		eventHandlers: nil,
	}
}

func (hs *eventTypeHandlers[ET, EV, HK]) add(key HK, h EventHandler[ET, EV], once bool) {
	if hs.eventHandlers == nil {
		hs.eventHandlers = newEventHandlers[ET, EV, HK]()
	}
	hs.eventHandlers.add(key, h, once)
}

func (hs *eventTypeHandlers[ET, EV, HK]) rem(key HK) {
	if hs.eventHandlers == nil {
		return
	}
	hs.eventHandlers.rem(key)
	if hs.eventHandlers.empty() {
		hs.eventHandlers = nil
	}
}

func (hs *eventTypeHandlers[ET, EV, HK]) addVal(ev EV, key HK, h EventHandler[ET, EV], once bool) {
	if hs.valHandlers == nil {
		hs.valHandlers = map[EV]*eventHandlers[ET, EV, HK]{}
	}

	container := hs.valHandlers[ev]
	if container == nil {
		container = newEventHandlers[ET, EV, HK]()
		hs.valHandlers[ev] = container
	}

	container.add(key, h, once)
}

func (hs *eventTypeHandlers[ET, EV, HK]) remVal(ev EV, key HK) {
	container := hs.valHandlers[ev]
	if container != nil {
		container.rem(key)
		if container.empty() {
			delete(hs.valHandlers, ev)
		}
		if len(hs.valHandlers) == 0 {
			hs.valHandlers = nil
		}
	}
}

func (hs *eventTypeHandlers[ET, EV, HK]) empty() bool {
	return hs.eventHandlers == nil && hs.valHandlers == nil
}

func (hs *eventTypeHandlers[ET, EV, HK]) call(evtID EventID[ET, EV], generator interface{}, param interface{}) error {
	if hs.eventHandlers != nil {
		err := hs.eventHandlers.call(evtID, generator, param)
		if hs.eventHandlers.empty() {
			hs.eventHandlers = nil
		}
		if err != nil {
			return err
		}
	}

	if hs.valHandlers != nil {
		container := hs.valHandlers[evtID.Val]
		if container != nil {
			err := container.call(evtID, generator, param)
			if container.empty() {
				delete(hs.valHandlers, evtID.Val)
				if len(hs.valHandlers) == 0 {
					hs.valHandlers = nil
				}
			}
			return err
		}
	}

	return nil
}

func (hs *eventTypeHandlers[ET, EV, HK]) clear() {
	for _, v := range hs.valHandlers {
		v.clear()
	}
	hs.valHandlers = nil

	if hs.eventHandlers != nil {
		hs.eventHandlers.clear()
		hs.eventHandlers = nil
	}
}
