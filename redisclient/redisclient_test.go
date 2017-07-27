// arkadiko
// https://github.com/topfreegames/arkadiko
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package redisclient_test

import (
	"github.com/garyburd/redigo/redis"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/topfreegames/arkadiko/redisclient"
)

var _ = Describe("Redis Client", func() {
	l, _ := test.NewNullLogger()
	logger := l.WithFields(log.Fields{})

	Describe("Specs", func() {
		Describe("Redis Client", func() {
			It("It should set and get without error", func() {
				rc := redisclient.GetRedisClient("localhost", 4444, "", logger)
				_, err := rc.Pool.Get().Do("set", "teste", 1)
				Expect(err).To(BeNil())
				res, err := redis.Int(rc.Pool.Get().Do("get", "teste"))
				Expect(err).To(BeNil())
				Expect(res).To(Equal(1))
			})
		})
	})

	Describe("Perf", func() {
		Measure("it should set item fast", func(b Benchmarker) {
			rc := redisclient.GetRedisClient("localhost", 4444, "", logger)
			r := rc.Pool.Get()

			runtime := b.Time("runtime", func() {
				_, err := r.Do("set", "teste", 1)
				Expect(err).NotTo(HaveOccurred())
			})

			Expect(runtime.Seconds()).Should(BeNumerically("<", 0.05), "Redis client set shouldn't take too long.")
		}, 20)
	})
})
