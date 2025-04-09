/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package utils

import (
	"errors"
	"github.com/go-pg/pg"
	"io"
	"net"
	"os"
)

const (
	PgNetworkErrorLogPrefix string = "PG_NETWORK_ERROR"
	PgQueryFailLogPrefix    string = "PG_QUERY_FAIL"
	PgQuerySlowLogPrefix    string = "PG_QUERY_SLOW"
)

const (
	FAIL    string = "FAIL"
	SUCCESS string = "SUCCESS"
)

type ErrorType string

func (e ErrorType) String() string {
	return string(e)
}

const (
	NetworkErrorType ErrorType = "NETWORK_ERROR"
	SyntaxErrorType  ErrorType = "SYNTAX_ERROR"
	TimeoutErrorType ErrorType = "TIMEOUT_ERROR"
	NoErrorType      ErrorType = "NA"
)

func getErrorType(err error) ErrorType {
	if err == nil {
		return NoErrorType
	} else if errors.Is(err, os.ErrDeadlineExceeded) {
		return TimeoutErrorType
	} else if isNetworkError(err) {
		return NetworkErrorType
	}
	return SyntaxErrorType
}

func isNetworkError(err error) bool {
	if err == io.EOF {
		return true
	}
	_, ok := err.(net.Error)
	return ok
}

func isIntegrityViolationError(err error) bool {
	pgErr, ok := err.(pg.Error)
	if !ok {
		return false
	}
	return pgErr.IntegrityViolation()
}
