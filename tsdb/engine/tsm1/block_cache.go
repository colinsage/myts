package tsm1

import (
	"github.com/influxdata/influxdb/models"
	"context"
)

type BlockCache struct {

	IndexMap map[string]*TimeSerieBlockIndexList

	BlockMap map[string]*[]byte


}

type TimeSerieBlockIndexList struct {
	// sort by time range
	IndexList []*TimeSerieBlockIndex
	FieldType   models.FieldType
}

type TimeSerieBlockIndex struct {
	start  uint64
	end   uint64
}

type KeyCursorInCache struct {
	key []byte

	buf     []Value

	pos  int
	ascending bool
}

func newKeyCursorInCache( key []byte, t int64, ascending bool) *KeyCursorInCache {
	c := &KeyCursorInCache{
		key:       key,
		ascending: ascending,
	}

	return c
}

func (c *KeyCursorInCache) Close() {
	// Remove all of our in-use references since we're done
	c.buf = nil
}

func (c *KeyCursorInCache) Next() {
	if len(c.current) == 0 {
		return
	}
	// Do we still have unread values in the current block
	if !c.current[0].read() {
		return
	}
	c.current = c.current[:0]
	if c.ascending {
		c.nextAscending()
	} else {
		c.nextDescending()
	}
}