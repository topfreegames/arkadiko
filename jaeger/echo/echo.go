// arkadiko
// https://github.com/topfreegames/arkadiko
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2018 Top Free Games <backend@tfgco.com>

package echo

import (
	"github.com/labstack/echo"
)

// Echo is the top-level framework instance.
type Echo struct {
	*echo.Echo
}

// New creates an instance of Echo.
func New() *Echo {
	app := echo.New()
	Instrument(app)
	return &Echo{app}
}
