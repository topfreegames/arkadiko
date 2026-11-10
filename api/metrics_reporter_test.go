// arkadiko
// https://github.com/topfreegames/arkadiko
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/arkadiko/api"
)

var _ = Describe("Metrics", func() {
	It("Should count a disconnection by its error", func() {
		Expect(func() {
			api.NewMetrics().DisconnectionCounter.WithLabelValues("EOF").Inc()
		}).NotTo(Panic())
	})
})
