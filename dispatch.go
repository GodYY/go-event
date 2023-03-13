package goevent

// EventDispatcher 事件派发器
// 事件产生前，用于注册事件处理器
// 事件产生时，用于将事件派发到对应的处理起进行处理
type EventDispatcher[ET, EV, HK comparable] struct {
	typesHandlers map[ET]*eventTypeHandlers[ET, EV, HK]
	isCalling     bool
}

func NewEventDispatcher[ET, EV, HK comparable]() *EventDispatcher[ET, EV, HK] {
	return &EventDispatcher[ET, EV, HK]{
		typesHandlers: map[ET]*eventTypeHandlers[ET, EV, HK]{},
	}
}

func (d *EventDispatcher[ET, EV, HK]) AddTypeHandler(evtType ET, key HK, h EventHandler[ET, EV], once bool) {
	typeHandler := d.addORGetTypeHandlers(evtType)
	typeHandler.add(key, h, once)
}

func (d *EventDispatcher[ET, EV, HK]) AddHandler(evtId EventID[ET, EV], key HK, h EventHandler[ET, EV], once bool) {
	typeHandlers := d.addORGetTypeHandlers(evtId.Type)
	typeHandlers.addVal(evtId.Val, key, h, once)
}

func (d *EventDispatcher[ET, EV, HK]) RemTypeHandler(evtType ET, key HK) {
	typeHandlers := d.typesHandlers[evtType]
	if typeHandlers == nil {
		return
	}
	typeHandlers.rem(key)
	if typeHandlers.empty() {
		delete(d.typesHandlers, evtType)
	}
}

func (d *EventDispatcher[ET, EV, HK]) RemHandler(evtId EventID[ET, EV], key HK) {
	typeHandlers := d.typesHandlers[evtId.Type]
	if typeHandlers == nil {
		return
	}
	typeHandlers.remVal(evtId.Val, key)
	if typeHandlers.empty() {
		delete(d.typesHandlers, evtId.Type)
	}
}

func (d *EventDispatcher[ET, EV, HK]) Clear() {
	for _, v := range d.typesHandlers {
		v.clear()
	}
	d.typesHandlers = map[ET]*eventTypeHandlers[ET, EV, HK]{}
	d.isCalling = false
}

func (d *EventDispatcher[ET, EV, HK]) Dispatch(evtId EventID[ET, EV], generator interface{}, param ...interface{}) error {
	typeHandlers := d.typesHandlers[evtId.Type]
	if typeHandlers == nil {
		return nil
	}

	if d.isCalling {
		panic("nested Dispatch calls")
	}

	d.isCalling = true
	err := typeHandlers.call(evtId, generator, param)
	if typeHandlers.empty() {
		delete(d.typesHandlers, evtId.Type)
	}
	d.isCalling = false
	return err
}

func (d *EventDispatcher[ET, EV, HK]) addORGetTypeHandlers(evtType ET) *eventTypeHandlers[ET, EV, HK] {
	typeHandlers := d.typesHandlers[evtType]
	if typeHandlers == nil {
		typeHandlers = newEventTypeHandlers[ET, EV, HK]()
		d.typesHandlers[evtType] = typeHandlers
	}
	return typeHandlers
}
