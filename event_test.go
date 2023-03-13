package goevent

import (
	"math/rand"
	"testing"
)

type testET int
type testEV int64
type testHK int64
type testEventID = EventID[testET, testEV]
type testEvent = Event[testET, testEV]

type testEventHandler struct {
	h func(testEvent) error
}

func (h *testEventHandler) HandleEvent(evt testEvent) error {
	return h.h(evt)
}

func TestTypeHandler(t *testing.T) {
	dispatcher := NewEventDispatcher[testET, testEV, testHK]()
	eventType := testET(1)
	value := 0

	key1 := testHK(1)
	add1 := 1
	handler1 := &testEventHandler{
		h: func(e testEvent) error {
			value += add1
			return nil
		},
	}

	key2 := testHK(2)
	add2 := 2
	handler2 := &testEventHandler{
		h: func(e testEvent) error {
			value += add2
			return nil
		},
	}

	dispatcher.AddTypeHandler(eventType, key1, handler1, false)
	dispatcher.AddTypeHandler(eventType, key2, handler2, false)
	if err := dispatcher.Dispatch(testEventID{eventType, 1}, nil, nil); err != nil {
		t.Fatal("there must no error")
	}
	if value != add1+add2 {
		t.Fatal("value must be", add1+add2)
	}

	value = 0
	dispatcher.RemTypeHandler(eventType, key2)
	if err := dispatcher.Dispatch(testEventID{eventType, 1}, nil, nil); err != nil {
		t.Fatal("there must no error")
	}
	if value != add1 {
		t.Fatal("value must be", add1)
	}

	value = 0
	dispatcher.RemTypeHandler(eventType, key1)
	if err := dispatcher.Dispatch(testEventID{eventType, 1}, nil, nil); err != nil {
		t.Fatal("there must no error")
	}
	if value != 0 {
		t.Fatal("value must be", 0)
	}
	if len(dispatcher.typesHandlers) != 0 {
		t.Fatal("handlers must clear")
	}

	value = 0
	dispatcher.AddTypeHandler(eventType, key2, handler2, true)
	if err := dispatcher.Dispatch(testEventID{eventType, 1}, nil, nil); err != nil {
		t.Fatal("there must no error")
	}
	if value != add2 {
		t.Fatal("value must be", add2)
	}
	if len(dispatcher.typesHandlers) != 0 {
		t.Fatal("handlers must clear")
	}
}

func TestHandler(t *testing.T) {
	dispatcher := NewEventDispatcher[testET, testEV, testHK]()
	eventType1 := testET(1)
	eventType2 := testET(2)
	eventVal1 := testEV(1)
	eventVal2 := testEV(2)
	defaultValue := 0
	value := defaultValue

	key1 := testHK(1)
	add1 := rand.Int()
	handler1 := &testEventHandler{
		h: func(e testEvent) error {
			value += add1
			return nil
		},
	}

	key2 := testHK(2)
	add2 := rand.Int()
	handler2 := &testEventHandler{
		h: func(e testEvent) error {
			value += add2
			return nil
		},
	}

	dispatcher.AddHandler(testEventID{eventType1, eventVal1}, key1, handler1, false)
	dispatcher.AddHandler(testEventID{eventType2, eventVal2}, key2, handler2, false)
	if err := dispatcher.Dispatch(testEventID{eventType1, eventVal1}, nil, nil); err != nil {
		t.Fatal("there must no error")
	}
	if err := dispatcher.Dispatch(testEventID{eventType2, eventVal2}, nil, nil); err != nil {
		t.Fatal("there must no error")
	}
	if value != add1+add2 {
		t.Fatal("value must be", add1+add2)
	}

	value = 0
	if err := dispatcher.Dispatch(testEventID{eventType1, 0}, nil, nil); err != nil {
		t.Fatal("there must no error")
	}
	if err := dispatcher.Dispatch(testEventID{eventType2, 0}, nil, nil); err != nil {
		t.Fatal("there must no error")
	}
	if value != defaultValue {
		t.Fatal("value must be", defaultValue)
	}

	value = 0
	if err := dispatcher.Dispatch(testEventID{eventType1, eventVal1}, nil, nil); err != nil {
		t.Fatal("there must no error")
	}
	if value != add1 {
		t.Fatal("value must be", add1)
	}

	value = 0
	dispatcher.AddTypeHandler(eventType1, key1, handler1, false)
	if err := dispatcher.Dispatch(testEventID{eventType1, eventVal1}, nil, nil); err != nil {
		t.Fatal("there must no error")
	}
	if value != add1+add1 {
		t.Fatal("value must be", add1+add1)
	}

	value = 0
	if err := dispatcher.Dispatch(testEventID{eventType2, eventVal2}, nil, nil); err != nil {
		t.Fatal("there must no error")
	}
	if value != add2 {
		t.Fatal("value must be", add2)
	}

	value = 0
	dispatcher.AddTypeHandler(eventType2, key2, handler2, false)
	if err := dispatcher.Dispatch(testEventID{eventType2, eventVal2}, nil, nil); err != nil {
		t.Fatal("there must no error")
	}
	if value != add2+add2 {
		t.Fatal("value must be", add2+add2)
	}

	value = 0
	if err := dispatcher.Dispatch(testEventID{eventType1, eventVal1}, nil, nil); err != nil {
		t.Fatal("there must no error")
	}
	if err := dispatcher.Dispatch(testEventID{eventType2, eventVal2}, nil, nil); err != nil {
		t.Fatal("there must no error")
	}
	if value != add1*2+add2*2 {
		t.Fatal("value must be", add1*2+add2*2)
	}

	value = 0
	dispatcher.Clear()
	if err := dispatcher.Dispatch(testEventID{eventType1, eventVal1}, nil, nil); err != nil {
		t.Fatal("there must no error")
	}
	if err := dispatcher.Dispatch(testEventID{eventType2, eventVal2}, nil, nil); err != nil {
		t.Fatal("there must no error")
	}
	if value != defaultValue {
		t.Fatal("value must be", defaultValue)
	}
}

func BenchmarkAddHandler(b *testing.B) {
	dispatcher := NewEventDispatcher[testET, testEV, testHK]()
	maxEventType := 100
	maxEventVal := 1000
	handler := &testEventHandler{h: func(event testEvent) error {
		return nil
	}}

	b.Run("build", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			dispatcher.AddHandler(
				testEventID{testET(rand.Intn(maxEventType)), testEV(rand.Intn(maxEventVal))},
				testHK(i),
				handler,
				false,
			)
		}
	})

	dispatcher.Clear()
	for i := 0; i < maxEventType; i++ {
		dispatcher.AddTypeHandler(testET(i), 0, handler, false)
	}
	b.Run("dispatch type", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			dispatcher.Dispatch(
				testEventID{testET(rand.Intn(maxEventType)), testEV(rand.Intn(maxEventVal))},
				nil,
				nil,
			)
		}
	})

	mm := map[testET]*testEventHandler{}
	for i := 0; i < maxEventType; i++ {
		mm[testET(i)] = handler
	}
	b.Run("dispatch type map", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			evt := testEvent{
				eventID: testEventID{testET(rand.Intn(maxEventType)), testEV(rand.Intn(maxEventVal))},
			}
			mm[testET(rand.Intn(maxEventType))].HandleEvent(evt)
		}
	})
}
