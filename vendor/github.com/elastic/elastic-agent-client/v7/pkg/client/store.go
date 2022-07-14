// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package client

import (
	"context"
	"errors"

	"github.com/elastic/elastic-agent-client/v7/pkg/proto"
)

// ErrStoreTxReadOnly is an error when write actions are performed on a read-only transaction.
var ErrStoreTxReadOnly = errors.New("transaction is read-only")

// ErrStoreTxBroken is an error when an action on a transaction has failed causing the whole transaction to be broken.
var ErrStoreTxBroken = errors.New("transaction is broken")

// ErrStoreTxDiscarded is an error when Commit is called on an already discarded transaction.
var ErrStoreTxDiscarded = errors.New("transaction was already discarded")

// ErrStoreTxCommitted is an error action is performed on an already committed transaction.
var ErrStoreTxCommitted = errors.New("transaction was already committed")

// StoreTxClient provides actions allowed for a started transaction.
type StoreTxClient interface {
	// GetKey fetches a value from its key in the store.
	GetKey(ctx context.Context, name string) ([]byte, bool, error)
	// SetKey sets a value for a key in the store.
	SetKey(ctx context.Context, name string, value []byte, ttl uint64) error
	// DeleteKey deletes a key from the store.
	DeleteKey(ctx context.Context, name string) error
	// Commit commits the transaction.
	Commit(ctx context.Context) error
	// Discard discards the transaction.
	//
	// Can be called even if already committed, in that case it does nothing.
	Discard(ctx context.Context) error
}

// StoreClient provides access to the key-value store from Elastic Agent for this unit.
type StoreClient interface {
	// BeginTx starts a transaction for the key-value store.
	BeginTx(ctx context.Context, write bool) (StoreTxClient, error)
}

type storeClientTx struct {
	client *storeClient
	txID   string
	write  bool

	brokenErr error
	discarded bool
	committed bool
}

type storeClient struct {
	client   *clientV2
	unitID   string
	unitType UnitType
}

// BeginTx starts a transaction for the key-value store.
func (c *storeClient) BeginTx(ctx context.Context, write bool) (StoreTxClient, error) {
	txType := proto.StoreTxType_READ_ONLY
	if write {
		txType = proto.StoreTxType_READ_WRITE
	}
	res, err := c.client.storeClient.BeginTx(ctx, &proto.StoreBeginTxRequest{
		Token:    c.client.token,
		UnitId:   c.unitID,
		UnitType: proto.UnitType(c.unitType),
		Type:     txType,
	})
	if err != nil {
		return nil, err
	}
	return &storeClientTx{
		client: c,
		txID:   res.Id,
		write:  write,
	}, nil
}

// GetKey fetches a value from its key in the store.
func (c *storeClientTx) GetKey(ctx context.Context, name string) ([]byte, bool, error) {
	if c.brokenErr != nil {
		return nil, false, ErrStoreTxBroken
	}
	if c.discarded {
		return nil, false, ErrStoreTxDiscarded
	}
	if c.committed {
		return nil, false, ErrStoreTxCommitted
	}
	res, err := c.client.client.storeClient.GetKey(ctx, &proto.StoreGetKeyRequest{
		Token: c.client.client.token,
		TxId:  c.txID,
		Name:  name,
	})
	if err != nil {
		c.brokenErr = err
		return nil, false, err
	}
	switch res.Status {
	case proto.StoreGetKeyResponse_FOUND:
		return res.Value, true, nil
	case proto.StoreGetKeyResponse_NOT_FOUND:
		return nil, false, nil
	}
	err = errors.New("unknown StoreGetKeyResponseStatus")
	c.brokenErr = err
	return nil, false, err
}

// SetKey sets a value for a key in the store.
func (c *storeClientTx) SetKey(ctx context.Context, name string, value []byte, ttl uint64) error {
	if c.brokenErr != nil {
		return ErrStoreTxBroken
	}
	if c.discarded {
		return ErrStoreTxDiscarded
	}
	if c.committed {
		return ErrStoreTxCommitted
	}
	if !c.write {
		return ErrStoreTxReadOnly
	}
	_, err := c.client.client.storeClient.SetKey(ctx, &proto.StoreSetKeyRequest{
		Token: c.client.client.token,
		TxId:  c.txID,
		Name:  name,
		Value: value,
		Ttl:   ttl,
	})
	if err != nil {
		c.brokenErr = err
		return err
	}
	return nil
}

// DeleteKey deletes a key from the store.
func (c *storeClientTx) DeleteKey(ctx context.Context, name string) error {
	if c.brokenErr != nil {
		return ErrStoreTxBroken
	}
	if c.discarded {
		return ErrStoreTxDiscarded
	}
	if c.committed {
		return ErrStoreTxCommitted
	}
	if !c.write {
		return ErrStoreTxReadOnly
	}
	_, err := c.client.client.storeClient.DeleteKey(ctx, &proto.StoreDeleteKeyRequest{
		Token: c.client.client.token,
		TxId:  c.txID,
		Name:  name,
	})
	if err != nil {
		c.brokenErr = err
		return err
	}
	return err
}

// Commit commits the transaction.
func (c *storeClientTx) Commit(ctx context.Context) error {
	if c.brokenErr != nil {
		return ErrStoreTxBroken
	}
	if c.discarded {
		return ErrStoreTxDiscarded
	}
	if c.committed {
		return nil
	}
	_, err := c.client.client.storeClient.CommitTx(ctx, &proto.StoreCommitTxRequest{
		Token: c.client.client.token,
		TxId:  c.txID,
	})
	if err != nil {
		c.brokenErr = err
		return err
	}
	c.committed = true
	return nil
}

// Discard discards the transaction.
//
// Can be called even if already committed, in that case it does nothing.
func (c *storeClientTx) Discard(ctx context.Context) error {
	if c.brokenErr != nil {
		return ErrStoreTxBroken
	}
	if c.discarded || c.committed {
		return nil
	}
	_, err := c.client.client.storeClient.DiscardTx(ctx, &proto.StoreDiscardTxRequest{
		Token: c.client.client.token,
		TxId:  c.txID,
	})
	if err != nil {
		c.brokenErr = err
		return err
	}
	c.discarded = true
	return nil
}
