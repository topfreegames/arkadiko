// arkadiko
// https://github.com/topfreegames/arkadiko
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package lib

import "context"

// ArkadikoInterface defines the interface for the arkadiko client
type ArkadikoInterface interface {
	SendMQTT(ctx context.Context, topic string, payload interface{}, retained bool) (*SendMQTTResponse, error)
}
