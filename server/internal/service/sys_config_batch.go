package service

import (
	"amiya-eden/internal/repository"
	"strconv"
)

type sysConfigBatch struct {
	items []repository.SysConfigUpsertItem
}

func newSysConfigBatch(capacity int) *sysConfigBatch {
	return &sysConfigBatch{items: make([]repository.SysConfigUpsertItem, 0, capacity)}
}

func (b *sysConfigBatch) AddString(key, value, desc string) *sysConfigBatch {
	b.items = append(b.items, repository.SysConfigUpsertItem{Key: key, Value: value, Desc: desc})
	return b
}

func (b *sysConfigBatch) AddBool(key string, value bool, desc string) *sysConfigBatch {
	return b.AddString(key, strconv.FormatBool(value), desc)
}

func (b *sysConfigBatch) AddInt(key string, value int, desc string) *sysConfigBatch {
	return b.AddString(key, strconv.Itoa(value), desc)
}

func (b *sysConfigBatch) AddInt64(key string, value int64, desc string) *sysConfigBatch {
	return b.AddString(key, strconv.FormatInt(value, 10), desc)
}

func (b *sysConfigBatch) AddFloat64(key string, value float64, desc string) *sysConfigBatch {
	return b.AddString(key, strconv.FormatFloat(value, 'f', -1, 64), desc)
}

func (b *sysConfigBatch) Items() []repository.SysConfigUpsertItem {
	return append([]repository.SysConfigUpsertItem(nil), b.items...)
}
