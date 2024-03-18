package internal

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"
)

type cursorFile struct {
	Feeds map[string]map[string]time.Time
}

type Cursor struct {
	feeds map[string]map[string]time.Time // notifier:url:updated_at
	mu    sync.RWMutex
}

func NewCursor(filename string) (*Cursor, error) {
	f, err := os.Open(filename)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		f, err = os.Create(filename)
		if err != nil {
			return nil, err
		}
	}
	defer f.Close()

	if stat, err := f.Stat(); err != nil {
		return nil, err
	} else if stat.Size() == 0 {
		return &Cursor{
			feeds: make(map[string]map[string]time.Time),
		}, nil
	}

	c := cursorFile{
		Feeds: make(map[string]map[string]time.Time),
	}
	if err := json.NewDecoder(f).Decode(&c); err != nil {
		return nil, err
	}
	return &Cursor{
		feeds: c.Feeds,
		mu:    sync.RWMutex{},
	}, nil
}

func (c *Cursor) GetNotifierCursor(notifierName string) map[string]time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if v, ok := c.feeds[notifierName]; ok {
		return v
	}
	return map[string]time.Time{}
}

func (c *Cursor) SetNotifierCursor(notifierName string, cursor map[string]time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.feeds[notifierName] = cursor
}

func (c *Cursor) Write(filename string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(&cursorFile{
		Feeds: c.feeds,
	})
}
