// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package client

import (
	"bytes"
	"context"
	"errors"

	"github.com/elastic/elastic-agent-client/v7/pkg/proto"
)

// ArtifactsClient provides Fetch that allows artifacts to be fetched from the artifact store.
type ArtifactsClient interface {
	// Fetch fetches the artifact from the artifact store.
	Fetch(ctx context.Context, id string, sha256 string) ([]byte, error)
}

type artifactsClient struct {
	client *clientV2
}

// Fetch fetches the artifact from the artifact store.
func (c *artifactsClient) Fetch(ctx context.Context, id string, sha256 string) ([]byte, error) {
	var data bytes.Buffer
	client, err := c.client.artifactClient.Fetch(ctx, &proto.ArtifactFetchRequest{
		Token:  c.client.token,
		Id:     id,
		Sha256: sha256,
	})
	if err != nil {
		return nil, err
	}
	for {
		msg, err := client.Recv()
		if err != nil {
			return nil, err
		}
		switch res := msg.ContentEof.(type) {
		case *proto.ArtifactFetchResponse_Content:
			if _, err := data.Write(res.Content); err != nil {
				return nil, err
			}
		case *proto.ArtifactFetchResponse_Eof:
			return data.Bytes(), nil
		default:
			return nil, errors.New("unknown ArtifactContentEof type")
		}
	}
}
