package ttlstore

import (
	"container/heap"
	"context"
	"sync"
	"time"
)

// TTLItem see https://pkg.go.dev/container/heap
type TTLItem struct {
	Key       string
	ExpiresAt time.Time
	index     int // The index is needed by update
}

type TTLHeap []*TTLItem

func (h TTLHeap) Len() int           { return len(h) }
func (h TTLHeap) Less(i, j int) bool { return h[i].ExpiresAt.Before(h[j].ExpiresAt) }
func (h TTLHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}
func (h *TTLHeap) Push(x interface{}) {
	item := x.(*TTLItem)
	item.index = len(*h)
	*h = append(*h, item)
}

// Pop removes and returns the item with the earliest expiration time.
// This method implements heap.Interface and should be called via heap.Pop().
// The heap package automatically moves the root (minimum) element to the end
// before calling this method, so we remove and return the last element.
func (h *TTLHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*h = old[0 : n-1]
	return item
}

// Peek returns the item with the earliest expiration time without removing it.
// Returns nil if the heap is empty. This operation is O(1) since the minimum
// element is always at the root (index 0) of the min-heap.
func (h TTLHeap) Peek() *TTLItem {
	if len(h) == 0 {
		return nil
	}
	return h[0]
}

type TTLStore struct {
	mu       sync.Mutex
	heap     TTLHeap
	entries  map[string]*TTLItem
	wake     chan struct{}
	stop     chan struct{}
	DeleteFn func(key string)
}

// SetTTL sets the TTL for a key.
func (s *TTLStore) SetTTL(key string, expiresAt time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Overwrite existing key
	if old, exists := s.entries[key]; exists {
		heap.Remove(&s.heap, old.index)
		delete(s.entries, key)
	}

	item := &TTLItem{
		Key:       key,
		ExpiresAt: expiresAt,
	}
	heap.Push(&s.heap, item)
	s.entries[key] = item

	// Notify the worker to wake up
	select {
	case s.wake <- struct{}{}:
	default:
	}
}

// GetTTL returns the expiration time for a key.
func (s *TTLStore) GetTTL(key string) (time.Time, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, exists := s.entries[key]
	if !exists {
		return time.Time{}, false
	}
	return item.ExpiresAt, true
}

// run is the background worker that continuously monitors and processes expired items.
// It runs in a separate goroutine and handles three main scenarios:
// 1. Empty heap: waits for new items or stop signal
// 2. Items not yet expired: sleeps until next expiration or interruption
// 3. Expired items: removes them from heap/map and calls DeleteFn callback
func (s *TTLStore) run(ctx context.Context) {
	for {
		s.mu.Lock()
		next := s.heap.Peek()
		s.mu.Unlock()

		if next == nil {
			select {
			case <-s.wake:
				continue
			case <-ctx.Done():
				return
			}
		}

		sleep := time.Until(next.ExpiresAt)
		if sleep > 0 {
			// block goto sleep until one of the following happens: earliest item expires,
			// wake signal (a new item may expire earlier, so we continue iteration),
			// or stop signal
			select {
			case <-time.After(sleep):
			case <-s.wake:
				continue
			case <-ctx.Done():
				return
			}
		}
		// Expire items
		s.mu.Lock()
		// At this point we may have multiple items that are expired, iterate in a loop
		for {
			if s.heap.Len() == 0 || s.heap.Peek().ExpiresAt.After(time.Now()) {
				break
			}
			item := heap.Pop(&s.heap).(*TTLItem)
			delete(s.entries, item.Key)
			if s.DeleteFn != nil {
				go s.DeleteFn(item.Key)
			}
		}
		s.mu.Unlock()
	}
}

func (s *TTLStore) Stop() {
	close(s.stop)
}

func (s *TTLStore) FlushAll() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear the heap
	s.heap = TTLHeap{}
	heap.Init(&s.heap)

	// Clear the entries map
	s.entries = make(map[string]*TTLItem)
}

// NewTTLStore creates a new TTL scheduler
func NewTTLStore(ctx context.Context, deleteFn func(key string)) *TTLStore {
	s := &TTLStore{
		heap:    TTLHeap{},
		entries: make(map[string]*TTLItem),
		// Buffered channel up to 1 item to avoid blocking of the worker on wake signal
		wake:     make(chan struct{}, 1),
		stop:     make(chan struct{}),
		DeleteFn: deleteFn,
	}
	heap.Init(&s.heap)
	go s.run(ctx)
	return s
}
