package kafka

import (
	"fmt"
	"sync"
)

type NotifyHandler interface {
	Topic() string
	Push(msg []byte) error
}

type NotifyHandlerRegistry struct {
	mu       sync.Mutex
	handlers map[string]NotifyHandler
}

func NewNotifyHandlerRegistry() *NotifyHandlerRegistry {
	return &NotifyHandlerRegistry{
		handlers: make(map[string]NotifyHandler),
	}
}

func (r *NotifyHandlerRegistry) Register(handler NotifyHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[handler.Topic()] = handler
}

func (r *NotifyHandlerRegistry) GetHandler(topic string) (NotifyHandler, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	handler, exists := r.handlers[topic]
	return handler, exists
}

func (r *NotifyHandlerRegistry) Push(topic string, msg []byte) error {
	handler, exists := r.GetHandler(topic)
	if !exists {
		return fmt.Errorf("topic %s 没有注册对应的处理器", topic)
	}
	return handler.Push(msg)
}

func (r *NotifyHandlerRegistry) RegisterAll(handlers []NotifyHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, handler := range handlers {
		r.Register(handler)
	}
}
