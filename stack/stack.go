package stack

import (
	"encoding/binary"
	"encoding/json"
	"github.com/Lawliet-Chan/roller-go/types"
	"github.com/scroll-tech/go-ethereum/log"
	"go.etcd.io/bbolt"
)

type Stack struct {
	kvdb *bbolt.DB
}

var bucket = []byte("stack")

func NewStack(path string) *Stack {
	kvdb, err := bbolt.Open(path, 0666, nil)
	if err != nil {
		log.Crit("cannot open kvdb for stack", "error", err)
	}
	err = kvdb.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucket)
		return err
	})
	if err != nil {
		log.Crit("init stack failed", "error", err)
	}
	return &Stack{kvdb: kvdb}
}

func (s *Stack) Append(traces *types.BlockTraces) error {
	byt, err := json.Marshal(traces)
	if err != nil {
		return err
	}
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, traces.ID)
	return s.kvdb.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucket).Put(key, byt)
	})
}

func (s *Stack) Pop() (*types.BlockTraces, error) {
	var value []byte
	err := s.kvdb.Update(func(tx *bbolt.Tx) error {
		var key []byte
		bu := tx.Bucket(bucket)
		c := bu.Cursor()
		key, value = c.Last()
		return bu.Delete(key)
	})
	if err != nil {
		return nil, err
	}
	var traces = &types.BlockTraces{}
	err = json.Unmarshal(value, traces)
	return traces, err
}
