// arkadiko
// https://github.com/topfreegames/arkadiko
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"encoding/json"
	"net/http"
	"testing"

	. "github.com/franela/goblin"
	. "github.com/onsi/gomega"
)

func TestAuthorizeUsersHandler(t *testing.T) {
	g := Goblin(t)

	// special hook for gomega
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })
	g.Describe("Should authorize users", func() {
		g.It("Should respond with 200 for a valid message", func() {
			a := GetDefaultTestApp()
			var jsonPayload JSON
			testJSON := `{"userId": "felipe", "rooms": ["room1", "room2"]}`
			json.Unmarshal([]byte(testJSON), &jsonPayload)
			res := PostJSON(a, "/authorize_user", t, jsonPayload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
		})

		g.It("Should respond with 400 if malformed JSON", func() {
			a := GetDefaultTestApp()
			var jsonPayload JSON
			testJSON := `{"message": "hello"}}`
			json.Unmarshal([]byte(testJSON), &jsonPayload)
			res := PostJSON(a, "/authorize_user", t, jsonPayload)

			g.Assert(res.Raw().StatusCode).Equal(400)
		})

		g.It("Should respond with 400 if no userId is provided", func() {
			a := GetDefaultTestApp()
			var jsonPayload JSON
			testJSON := `{"message": "hello", "rooms": ["room1"]}`
			json.Unmarshal([]byte(testJSON), &jsonPayload)
			res := PostJSON(a, "/authorize_user", t, jsonPayload)

			g.Assert(res.Raw().StatusCode).Equal(400)
		})

		g.It("Should respond with 400 if no rooms is provided", func() {
			a := GetDefaultTestApp()
			var jsonPayload JSON
			testJSON := `{"userId": "hello", "roo": ["room1"]}`
			json.Unmarshal([]byte(testJSON), &jsonPayload)
			res := PostJSON(a, "/authorize_user", t, jsonPayload)

			g.Assert(res.Raw().StatusCode).Equal(400)
		})

	})
}
