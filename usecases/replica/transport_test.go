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
	ctx, cancel := context.WithDeadline(ctx, time.Now())
	defer cancel()
	err := &Error{Err: ctx.Err()}
	assert.True(t, err.Timeout())
	err = err.Clone()
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestReplicationErrorMarshal(t *testing.T) {
	rawErr := Error{Code: StatusClassNotFound, Msg: "Article", Err: errors.New("error cannot be marshalled")}
	bytes, err := json.Marshal(&rawErr)
	assert.Nil(t, err)
	got := NewError(0, "")
	assert.Nil(t, json.Unmarshal(bytes, got))
	want := &Error{Code: StatusClassNotFound, Msg: "Article"}
	assert.Equal(t, want, got)
}

func TestReplicationErrorStatus(t *testing.T) {
	tests := []struct {
		code StatusCode
		desc string
	}{
		{-1, ""},
		{StatusOK, "ok"},
		{StatusClassNotFound, "class not found"},
		{StatusShardNotFound, "shard not found"},
		{StatusNotFound, "not found"},
		{StatusAlreadyExisted, "already existed"},
		{StatusConflict, "conflict"},
		{StatusPreconditionFailed, "precondition failed"},
		{StatusReadOnly, "read only"},
	}
	for _, test := range tests {
		got := statusText(test.code)
		if got != test.desc {
			t.Errorf("statusText(%d) want %v got %v", test.code, test.desc, got)
		}
	}
}
