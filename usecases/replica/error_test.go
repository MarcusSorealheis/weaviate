package replica

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestReplicationErrorTimeout(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithDeadline(context.Background(), time.Now())
	defer cancel()
	err := &Error{Err: ctx.Err()}
	assert.True(t, err.Timeout())
	err = err.Clone()
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestReplicationErrorMarshal(t *testing.T) {
	rawErr := Error{Code: StatusFound, Msg: "message", Err: errors.New("error cannot be marshalled")}
	bytes, err := json.Marshal(&rawErr)
	assert.Nil(t, err)
	got := NewError(0, "")
	assert.Nil(t, json.Unmarshal(bytes, got))
	want := &Error{Code: StatusFound, Msg: "message"}
	assert.Equal(t, want, got)
}
