package replica

import (
	"encoding/json"
	"fmt"
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
	fmt.Println(rawErr.Error())

	bytes, err := json.Marshal(&rawErr)
	assert.Nil(t, err)
	got := NewError(0, "")
	assert.Nil(t, json.Unmarshal(bytes, got))
	want := &Error{Code: StatusClassNotFound, Msg: "Article"}
	assert.Equal(t, want, got)
}
