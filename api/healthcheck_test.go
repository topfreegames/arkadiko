// arkadiko
// https://github.com/topfreegames/arkadiko
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"net/http"
	"testing"

	. "github.com/franela/goblin"
	. "github.com/onsi/gomega"
)

func TestHealthcheckHandler(t *testing.T) {
	g := Goblin(t)

	// special hook for gomega
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Healthcheck Handler", func() {
		g.It("Should respond with default WORKING string", func() {
			a := GetDefaultTestApp()
			status, body := Get(a, "/healthcheck", t)

			g.Assert(status).Equal(http.StatusOK)
			g.Assert(body).Equal("WORKING")
		})

		g.It("Should respond with customized WORKING string", func() {
			a := GetDefaultTestApp()
			a.Config.SetDefault("healthcheck.workingText", "OTHERWORKING")
			status, body := Get(a, "/healthcheck", t)

			g.Assert(status).Equal(http.StatusOK)
			g.Assert(body).Equal("OTHERWORKING")
		})
	})
}
