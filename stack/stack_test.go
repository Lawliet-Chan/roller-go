package stack

import (
	"github.com/stretchr/testify/assert"
	"roller-go/types"
	"testing"
)

func TestStack(t *testing.T) {
	s := NewStack("test_stack")
	for i := 0; i < 3; i++ {
		trace := &types.BlockTraces{
			ID:     uint64(i),
			Traces: nil,
		}
		err := s.Append(trace)
		assert.NoError(t, err)
	}

	for i := 2; i >= 0; i-- {
		trace, err := s.Pop()
		assert.NoError(t, err)
		assert.Equal(t, uint64(i), trace.ID)
	}
}
