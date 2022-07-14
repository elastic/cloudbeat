// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package client

import (
	"context"

	"github.com/elastic/elastic-agent-client/v7/pkg/proto"
)

// LogClient provides Log that allows a message to be logged by Elastic Agent.
type LogClient interface {
	// Log logs a message for the unit.
	Log(ctx context.Context, message []byte) error
}

type logClient struct {
	client   *clientV2
	unitID   string
	unitType UnitType
}

// Log logs a message for the unit.
func (c *logClient) Log(ctx context.Context, message []byte) error {
	_, err := c.client.logClient.Log(ctx, &proto.LogMessageRequest{
		Token: c.client.token,
		Messages: []*proto.LogMessage{
			{
				UnitId:   c.unitID,
				UnitType: proto.UnitType(c.unitType),
				Message:  message,
			},
		},
	})
	return err
}
