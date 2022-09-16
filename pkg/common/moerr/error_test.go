// Copyright 2021 - 2022 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package moerr

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func pf1() {
	panic("foo")
}

func pf2(a, b int) int {
	return a / b
}

func pf3() {
	panic(NewInternalError("%s %s %s %d", "foo", "bar", "zoo", 2))
}

func PanicF(i int) (err *Error) {
	defer func() {
		if e := recover(); e != nil {
			err = NewPanicError(e)
		}
	}()
	switch i {
	case 1:
		pf1()
	case 2:
		foo := pf2(1, 0)
		panic(foo)
	case 3:
		pf3()
	default:
		return nil
	}
	return
}

func TestWrap(t *testing.T) {
	err := NewError(SUCCESS, "foo").Wrap(
		io.EOF,
		io.ErrNoProgress,
		&os.PathError{
			Op:   "read",
			Path: "foo",
		},
	)
	require.True(t, errors.Is(err, io.EOF))
	require.True(t, errors.Is(err, io.ErrNoProgress))
	var pathError *os.PathError
	require.True(t, errors.As(err, &pathError))
	require.Equal(t, "read", pathError.Op)
	require.Equal(t, "foo", pathError.Path)
}

func TestPanicError(t *testing.T) {
	for i := 0; i <= 3; i++ {
		err := PanicF(i)
		if i == 0 {
			if err != nil {
				t.Errorf("No panic should be OK")
			}
		} else {
			if err == nil {
				t.Errorf("Uncaught panic")
			}
			if err.Ok() {
				t.Errorf("Caught OK panic")
			}
		}
	}
}

func TestNew(t *testing.T) {
	type args struct {
		code Code
		args []any
	}
	tests := []struct {
		name        string
		args        args
		wantCode    Code
		wantState   SqlState
		wantMessage Message
	}{

		{
			name:     "DIVISION_BY_ZERO",
			args:     args{code: DIVIVISION_BY_ZERO, args: []any{}},
			wantCode: DIVIVISION_BY_ZERO, wantState: MySQLDefaultSqlState, wantMessage: "division by zero",
		},
		{
			name:     "OUT_OF_RANGE",
			args:     args{code: OUT_OF_RANGE, args: []any{"double", "bigint"}},
			wantCode: OUT_OF_RANGE, wantState: MySQLDefaultSqlState, wantMessage: "overflow from double to bigint",
		},
		{
			name:     "DATA_TRUNCATED",
			args:     args{code: DATA_TRUNCATED, args: []any{"decimal128"}},
			wantCode: DATA_TRUNCATED, wantState: MySQLDefaultSqlState, wantMessage: "decimal128 data truncated",
		},
		{
			name:     "BAD_CONFIGURATION",
			args:     args{code: BAD_CONFIGURATION, args: []any{"log"}},
			wantCode: BAD_CONFIGURATION, wantState: MySQLDefaultSqlState, wantMessage: "invalid log configuration",
		},
		{
			name:     "LOG_SERVICE_NOT_READY",
			args:     args{code: LOG_SERVICE_NOT_READY, args: []any{}},
			wantCode: LOG_SERVICE_NOT_READY, wantState: MySQLDefaultSqlState, wantMessage: "log service not ready",
		},
		{
			name:     "ErrClientClosed",
			args:     args{code: ErrClientClosed, args: []any{}},
			wantCode: ErrClientClosed, wantState: MySQLDefaultSqlState, wantMessage: "client closed",
		},
		{
			name:     "ErrBackendClosed",
			args:     args{code: ErrBackendClosed, args: []any{}},
			wantCode: ErrBackendClosed, wantState: MySQLDefaultSqlState, wantMessage: "backend closed",
		},
		{
			name:     "ErrStreamClosed",
			args:     args{code: ErrStreamClosed, args: []any{}},
			wantCode: ErrStreamClosed, wantState: MySQLDefaultSqlState, wantMessage: "stream closed",
		},
		{
			name:     "ErrNoAvailableBackend",
			args:     args{code: ErrNoAvailableBackend, args: []any{}},
			wantCode: ErrNoAvailableBackend, wantState: MySQLDefaultSqlState, wantMessage: "no available backend",
		},
		{
			name:     "ErrTxnClosed",
			args:     args{code: ErrTxnClosed, args: []any{}},
			wantCode: ErrTxnClosed, wantState: MySQLDefaultSqlState, wantMessage: "the transaction has been committed or aborted",
		},
		{
			name:     "ErrTxnWriteConflict",
			args:     args{code: ErrTxnWriteConflict, args: []any{}},
			wantCode: ErrTxnWriteConflict, wantState: MySQLDefaultSqlState, wantMessage: "write conflict",
		},
		{
			name:     "ErrMissingTxn",
			args:     args{code: ErrMissingTxn, args: []any{}},
			wantCode: ErrMissingTxn, wantState: MySQLDefaultSqlState, wantMessage: "missing txn",
		},
		{
			name:     "ErrUnresolvedConflict",
			args:     args{code: ErrUnresolvedConflict, args: []any{}},
			wantCode: ErrUnresolvedConflict, wantState: MySQLDefaultSqlState, wantMessage: "unresolved conflict",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.args.code, tt.args.args...)

			var code Code
			require.True(t, errors.As(got, &code))
			require.Equal(t, tt.wantCode, code)

			var state SqlState
			require.True(t, errors.As(got, &state))
			require.Equal(t, tt.wantState, state)

			var msg Message
			require.True(t, errors.As(got, &msg))
			require.Equal(t, tt.wantMessage, msg)
			require.Equal(t, string(tt.wantMessage), got.Error())

		})
	}
}

func TestNewError(t *testing.T) {
	type args struct {
		code Code
		msg  string
	}
	tests := []struct {
		name string
		args args
		want *Error
	}{
		{
			name: "DIVISION_BY_ZERO",
			want: New(DIVIVISION_BY_ZERO),
			args: args{code: DIVIVISION_BY_ZERO, msg: "division by zero"},
		},
		{
			name: "OUT_OF_RANGE",
			want: New(OUT_OF_RANGE, "double", "bigint"),
			args: args{code: OUT_OF_RANGE, msg: "overflow from double to bigint"},
		},
		{
			name: "DATA_TRUNCATED",
			want: New(DATA_TRUNCATED, "decimal128"),
			args: args{code: DATA_TRUNCATED, msg: "decimal128 data truncated"},
		},
		{
			name: "BAD_CONFIGURATION",
			want: New(BAD_CONFIGURATION, "log"),
			args: args{code: BAD_CONFIGURATION, msg: "invalid log configuration"},
		},
		{
			name: "LOG_SERVICE_NOT_READY",
			want: New(LOG_SERVICE_NOT_READY),
			args: args{code: LOG_SERVICE_NOT_READY, msg: "log service not ready"},
		},
		{
			name: "ErrClientClosed",
			want: New(ErrClientClosed),
			args: args{code: ErrClientClosed, msg: "client closed"},
		},
		{
			name: "ErrBackendClosed",
			want: New(ErrBackendClosed),
			args: args{code: ErrBackendClosed, msg: "backend closed"},
		},
		{
			name: "ErrStreamClosed",
			want: New(ErrStreamClosed),
			args: args{code: ErrStreamClosed, msg: "stream closed"},
		},
		{
			name: "ErrNoAvailableBackend",
			want: New(ErrNoAvailableBackend),
			args: args{code: ErrNoAvailableBackend, msg: "no available backend"},
		},
		{
			name: "ErrTxnClosed",
			want: New(ErrTxnClosed),
			args: args{code: ErrTxnClosed, msg: "the transaction has been committed or aborted"},
		},
		{
			name: "ErrTxnWriteConflict",
			want: New(ErrTxnWriteConflict),
			args: args{code: ErrTxnWriteConflict, msg: "write conflict"},
		},
		{
			name: "ErrMissingTxn",
			want: New(ErrMissingTxn),
			args: args{code: ErrMissingTxn, msg: "missing txn"},
		},
		{
			name: "ErrUnresolvedConflict",
			want: New(ErrUnresolvedConflict),
			args: args{code: ErrUnresolvedConflict, msg: "unresolved conflict"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewError(tt.args.code, tt.args.msg)
			require.Equal(t, tt.want, got)
			var wantMessage Message
			require.True(t, errors.As(tt.want, &wantMessage))
			require.Equal(t, string(wantMessage), got.Error())
			var wantCode Code
			require.True(t, errors.As(tt.want, &wantCode))
			require.Equal(t, IsMoErrCode(got, wantCode), true)
		})
	}
}

func TestNew_panic(t *testing.T) {
	type args struct {
		code Code
		msg  string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "panic",
			args: args{code: 65534},
			want: "not exist MOErrorCode: 65534",
		},
	}
	defer func() {
		var err any
		if err = recover(); err != nil {
			require.Equal(t, tests[0].want, err.(error).Error())
			t.Logf("err: %+v", err)
		}
	}()
	for _, tt := range tests {
		got := New(tt.args.code, tt.args.msg)
		require.Equal(t, nil, got)
	}
}

func TestNew_MyErrorCode(t *testing.T) {
	type args struct {
		code Code
		args []any
	}
	tests := []struct {
		name string
		args args
		want Code
	}{
		{
			name: "hasMysqlErrorCode",
			args: args{code: ER_NO_DB_ERROR, args: []any{}},
			want: 1046,
		},
	}
	for _, tt := range tests {
		got := New(tt.args.code, tt.args.args...)
		var code Code
		require.True(t, errors.As(got, &code))
		require.Equal(t, code, tt.want)
	}
}

func TestNewError_panic(t *testing.T) {
	type args struct {
		code Code
		msg  string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "panic",
			args: args{code: 65534, msg: "not exist error code"},
			want: "not exist MOErrorCode: 65534",
		},
	}
	defer func() {
		var err any
		if err = recover(); err != nil {
			require.Equal(t, tests[0].want, err.(error).Error())
			t.Logf("err: %+v", err)
		}
	}()
	for _, tt := range tests {
		got := NewError(tt.args.code, tt.args.msg)
		require.Equal(t, nil, got)
	}
}

func TestNewInfo(t *testing.T) {
	type args struct {
		msg string
	}
	tests := []struct {
		name string
		args args
		want *Error
	}{
		{
			name: "normal",
			args: args{msg: "info msg"},
			want: New(INFO, "info msg"),
		},
	}
	for _, tt := range tests {
		got := NewInfo(tt.args.msg)
		require.Equal(t, tt.want, got)
		var wantMessage Message
		require.True(t, errors.As(tt.want, &wantMessage))
		require.Equal(t, string(wantMessage), got.Error())
		var wantCode Code
		require.True(t, errors.As(tt.want, &wantCode))
		require.Equal(t, IsMoErrCode(got, wantCode), true)
	}
}

func TestNewWarn(t *testing.T) {
	type args struct {
		msg string
	}
	tests := []struct {
		name string
		args args
		want *Error
	}{
		{
			name: "normal",
			args: args{msg: "error msg"},
			want: New(WARN, "error msg"),
		},
	}
	for _, tt := range tests {
		got := NewWarn(tt.args.msg)
		require.Equal(t, tt.want, got)
		var wantMessage Message
		require.True(t, errors.As(tt.want, &wantMessage))
		require.Equal(t, string(wantMessage), got.Error())
		var wantCode Code
		require.True(t, errors.As(tt.want, &wantCode))
		var code Code
		require.True(t, errors.As(got, &code))
		require.Equal(t, wantCode, code)
		var state SqlState
		require.True(t, errors.As(got, &state))
		require.Equal(t, SqlState("HY000"), state)
		require.Equal(t, IsMoErrCode(got, wantCode), true)
	}
}

func TestIsMoErrCode(t *testing.T) {
	type args struct {
		e  error
		rc Code
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "not Error",
			args: args{e: errors.New("raw error"), rc: INFO},
			want: false,
		},
		{
			name: "End Error",
			args: args{e: New(ErrEnd, "max value of MOError"), rc: ErrEnd},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsMoErrCode(tt.args.e, tt.args.rc); got != tt.want {
				t.Errorf("IsMoErrCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewWithContext(t *testing.T) {
	type args struct {
		ctx  context.Context
		code Code
		args []any
	}
	tests := []struct {
		name        string
		args        args
		wantCode    Code
		wantState   SqlState
		wantMessage Message
	}{
		{
			name:     "normal",
			args:     args{ctx: context.Background(), code: DIVIVISION_BY_ZERO, args: []any{}},
			wantCode: DIVIVISION_BY_ZERO, wantState: MySQLDefaultSqlState, wantMessage: "division by zero",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewWithContext(tt.args.ctx, tt.args.code, tt.args.args...)

			var code Code
			require.True(t, errors.As(got, &code))
			require.Equal(t, tt.wantCode, code)

			var state SqlState
			require.True(t, errors.As(got, &state))
			require.Equal(t, tt.wantState, state)

			var msg Message
			require.True(t, errors.As(got, &msg))
			require.Equal(t, tt.wantMessage, msg)
			require.Equal(t, string(tt.wantMessage), got.Error())

		})
	}
}
