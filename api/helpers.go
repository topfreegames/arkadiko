// arkadiko
// https://github.com/topfreegames/arkadiko
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"fmt"

	"github.com/labstack/echo"
	newrelic "github.com/newrelic/go-agent"
)

// FailWith fails with the specified message
func FailWith(status int, message string, c echo.Context) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	return c.String(status, fmt.Sprintf(`{"success":false,"reason":"%s"}`, message))
}

// SucceedWith sends payload to user with status 200
func SucceedWith(payload map[string]interface{}, c echo.Context) error {
	if len(payload) == 0 {
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		return c.String(200, `{"success":true}`)
	}
	payload["success"] = true
	return c.JSON(200, payload)
}

//GetTX returns new relic transaction
func GetTX(c echo.Context) newrelic.Transaction {
	tx := c.Get("txn")
	if tx == nil {
		return nil
	}

	return tx.(newrelic.Transaction)
}

//WithSegment adds a segment to new relic transaction
func WithSegment(name string, c echo.Context, f func() error) error {
	tx := GetTX(c)
	if tx == nil {
		return f()
	}
	segment := newrelic.StartSegment(tx, name)
	defer segment.End()
	return f()
}
