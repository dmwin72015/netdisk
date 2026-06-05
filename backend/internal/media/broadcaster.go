package media

import "sync"

type ProgressEvent struct {
	UserID    int64  `json:"-"`
	MediaSlug string `json:"mediaSlug"`
	Progress  int    `json:"progress"`
	Status    string `json:"status"`
	ErrorMsg  string `json:"errorMsg,omitempty"`
}

type Broadcaster struct {
	mu          sync.RWMutex
	subscribers []chan ProgressEvent
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{}
}

func (b *Broadcaster) Subscribe() chan ProgressEvent {
	b.mu.Lock()
	defer b.mu.Unlock()
	ch := make(chan ProgressEvent, 64)
	b.subscribers = append(b.subscribers, ch)
	return ch
}

func (b *Broadcaster) Unsubscribe(ch chan ProgressEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for i, sub := range b.subscribers {
		if sub == ch {
			b.subscribers = append(b.subscribers[:i], b.subscribers[i+1:]...)
			close(ch)
			return
		}
	}
}

func (b *Broadcaster) Publish(event ProgressEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.subscribers {
		select {
		case ch <- event:
		default:
		}
	}
}
