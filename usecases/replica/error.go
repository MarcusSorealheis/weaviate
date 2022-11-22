//                           _       _
// __      _____  __ ___   ___  __ _| |_ ___
// \ \ /\ / / _ \/ _` \ \ / / |/ _` | __/ _ \
//  \ V  V /  __/ (_| |\ V /| | (_| | ||  __/
//   \_/\_/ \___|\__,_| \_/ |_|\__,_|\__\___|
//
//  Copyright © 2016 - 2022 SeMI Technologies B.V. All rights reserved.
//
//  CONTACT: hello@semi.technology
//

package replica

import (
	"fmt"
	"net/http"
)

type StatusCode int

const (
	StatusOK            = iota
	StatusClassNotFound = iota + 100
	StatusShardNotFound
	StatusNotFound
	StatusFound
	StatusReadOnly = http.StatusAccepted
)

// Error
type Error struct {
	Code StatusCode `json:"code"`
	Msg  string     `json:"msg,omitempty"`
	Err  error      `json:"-"`
}

// Empty checks whether e is an empty error which equivalent to e == nil
func (e *Error) Empty() bool {
	return e.Code == StatusOK && e.Msg == "" && e.Err == nil
}

func NewError(code StatusCode, msg string, err error) *Error {
	return &Error{code, msg, err}
}

// Unwrap underlying error
func (e *Error) Unwrap() error { return e.Err }

func (e *Error) Error() string { return fmt.Sprintf("%d %q, :%v", e.Code, e.Msg, e.Err) }

func (e *Error) Timeout() bool {
	t, ok := e.Err.(interface {
		Timeout() bool
	})
	return ok && t.Timeout()
}
