// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package client

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"sync"
	"time"

	"google.golang.org/grpc"

	"github.com/elastic/elastic-agent-client/v7/pkg/proto"
	"github.com/elastic/elastic-agent-client/v7/pkg/utils"
)

// UnitChangedType defines types for when units are adjusted.
type UnitChangedType int

const (
	// UnitChangedAdded is when a new unit is added.
	UnitChangedAdded UnitChangedType = 1
	// UnitChangedModified is when an existing unit is modified.
	UnitChangedModified UnitChangedType = 2
	// UnitChangedRemoved is when an existing unit is removed.
	UnitChangedRemoved UnitChangedType = 3
)

// UnitChanged is what is sent over the UnitChanges channel any time a unit is added, modified, or removed.
type UnitChanged struct {
	Type UnitChangedType
	Unit *Unit
}

// VersionInfo is the version information for the connecting client.
type VersionInfo struct {
	// Name is the name of the program.
	Name string
	// Version is the current version of the program.
	Version string
	// Meta is any extra metadata information about the version.
	Meta map[string]string
}

// V2 manages the state and communication to the Elastic Agent over the V2 control protocol.
type V2 interface {
	// Start starts the connection to Elastic Agent.
	Start(ctx context.Context) error
	// Stop stops the connection to Elastic Agent.
	Stop()
	// UnitChanges returns channel client send unit change notifications to.
	//
	// User of this client must read from this channel, or it will block the client.
	UnitChanges() <-chan UnitChanged
	// Errors returns channel of errors that occurred during communication.
	//
	// User of this client must read from this channel, or it will block the client.
	Errors() <-chan error
	// Artifacts returns the artifacts client.
	Artifacts() ArtifactsClient
}

// clientV2 manages the state and communication to the Elastic Agent over the V2 control protocol.
type clientV2 struct {
	target string
	opts   []grpc.DialOption
	token  string

	versionInfo     VersionInfo
	versionInfoSent bool

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	client proto.ElasticAgentClient
	cfgMu  sync.RWMutex
	obsMu  sync.RWMutex

	kickCh  chan struct{}
	errCh   chan error
	unitsCh chan UnitChanged
	unitsMu sync.RWMutex
	units   []*Unit

	storeClient    proto.ElasticAgentStoreClient
	artifactClient proto.ElasticAgentArtifactClient
	logClient      proto.ElasticAgentLogClient

	// overridden in tests to make fast
	minCheckTimeout time.Duration
}

// NewV2 creates a client connection to Elastic Agent over the V2 control protocol.
func NewV2(target string, token string, versionInfo VersionInfo, opts ...grpc.DialOption) V2 {
	return &clientV2{
		target:          target,
		opts:            opts,
		token:           token,
		versionInfo:     versionInfo,
		kickCh:          make(chan struct{}, 1),
		errCh:           make(chan error),
		unitsCh:         make(chan UnitChanged),
		minCheckTimeout: CheckinMinimumTimeout,
	}
}

// Start starts the connection to Elastic Agent.
func (c *clientV2) Start(ctx context.Context) error {
	c.ctx, c.cancel = context.WithCancel(ctx)
	conn, err := grpc.DialContext(ctx, c.target, c.opts...)
	if err != nil {
		return err
	}
	c.client = proto.NewElasticAgentClient(conn)
	c.storeClient = proto.NewElasticAgentStoreClient(conn)
	c.artifactClient = proto.NewElasticAgentArtifactClient(conn)
	c.logClient = proto.NewElasticAgentLogClient(conn)
	c.startCheckin()
	c.startActions()
	return nil
}

// Stop stops the connection to Elastic Agent.
func (c *clientV2) Stop() {
	if c.cancel != nil {
		c.cancel()
		c.wg.Wait()
		c.ctx = nil
		c.cancel = nil
	}
}

// UnitChanges returns channel client send unit change notifications to.
func (c *clientV2) UnitChanges() <-chan UnitChanged {
	return c.unitsCh
}

// Errors returns channel of errors that occurred during communication.
func (c *clientV2) Errors() <-chan error {
	return c.errCh
}

// Artifacts returns the artifacts client.
func (c *clientV2) Artifacts() ArtifactsClient {
	return &artifactsClient{c}
}

// startCheckin starts the go routines to send and receive check-ins
//
// This starts 3 go routines to manage the check-in bi-directional stream. The first
// go routine starts the stream then starts one go routine to receive messages and
// another go routine to send messages. The first go routine then blocks waiting on
// the receive and send to finish, then restarts the stream or exits if the context
// has been cancelled.
func (c *clientV2) startCheckin() {
	c.wg.Add(1)

	go func() {
		defer c.wg.Done()
		for {
			select {
			case <-c.ctx.Done():
				// stopped
				return
			default:
			}

			c.checkinRoundTrip()
		}
	}()
}

func (c *clientV2) checkinRoundTrip() {
	checkinCtx, checkinCancel := context.WithCancel(c.ctx)
	defer checkinCancel()

	checkinClient, err := c.client.CheckinV2(checkinCtx)
	if err != nil {
		c.errCh <- err
		return
	}

	// ensure first checkin include version information
	c.versionInfoSent = false

	var checkinRead sync.WaitGroup
	var checkinWrite sync.WaitGroup
	done := make(chan bool)

	// expected state check-ins
	checkinRead.Add(1)
	go func() {
		defer checkinRead.Done()
		for {
			expected, err := checkinClient.Recv()
			if err != nil {
				if !errors.Is(err, io.EOF) {
					c.errCh <- err
				}
				close(done)
				return
			}
			c.syncUnits(expected)
		}
	}()

	// observed state check-ins
	checkinWrite.Add(1)
	go func() {
		defer checkinWrite.Done()

		if err := c.sendObserved(checkinClient); err != nil {
			if !errors.Is(err, io.EOF) {
				c.errCh <- err
			}
			return
		}

		t := time.NewTicker(c.minCheckTimeout)
		defer t.Stop()

		for {
			select {
			case <-done:
				return
			case <-c.kickCh:
				if err := c.sendObserved(checkinClient); err != nil {
					if !errors.Is(err, io.EOF) {
						c.errCh <- err
					}
					return
				}
				t.Reset(c.minCheckTimeout)
			case <-t.C:
				if err := c.sendObserved(checkinClient); err != nil {
					if !errors.Is(err, io.EOF) {
						c.errCh <- err
					}
					return
				}
			}
		}
	}()

	// wait for write goroutine to quit then close the client
	checkinWrite.Wait()
	checkinClient.CloseSend()

	// then wait for the read to quit
	checkinRead.Wait()
}

// sendObserved sends the observed state of all the units.
func (c *clientV2) sendObserved(client proto.ElasticAgent_CheckinV2Client) error {
	c.unitsMu.RLock()
	observed := make([]*proto.UnitObserved, 0, len(c.units))
	for _, unit := range c.units {
		observed = append(observed, unit.toObserved())
	}
	c.unitsMu.RUnlock()
	msg := &proto.CheckinObserved{
		Token:       c.token,
		Units:       observed,
		VersionInfo: nil,
	}
	if !c.versionInfoSent {
		c.versionInfoSent = true
		msg.VersionInfo = &proto.CheckinObservedVersionInfo{
			Name:    c.versionInfo.Name,
			Version: c.versionInfo.Version,
			Meta:    c.versionInfo.Meta,
		}
	}
	return client.Send(msg)
}

// syncUnits syncs the expected units with the current state.
func (c *clientV2) syncUnits(expected *proto.CheckinExpected) {
	c.unitsMu.Lock()
	defer c.unitsMu.Unlock()
	i := 0
	for _, unit := range c.units {
		if inExpected(unit, expected.Units) {
			c.units[i] = unit
			i++
		} else {
			c.unitsCh <- UnitChanged{
				Type: UnitChangedRemoved,
				Unit: unit,
			}
		}
	}
	// resize so units that no longer exist are removed from the slice
	c.units = c.units[:i]
	for _, agentUnit := range expected.Units {
		unit := c.findUnit(agentUnit.Id, UnitType(agentUnit.Type))
		if unit == nil {
			// new unit
			unit = newUnit(agentUnit.Id, UnitType(agentUnit.Type), UnitState(agentUnit.State), agentUnit.Config, agentUnit.ConfigStateIdx, c)
			c.units = append(c.units, unit)
			c.unitsCh <- UnitChanged{
				Type: UnitChangedAdded,
				Unit: unit,
			}
		} else {
			// existing unit
			if unit.updateState(UnitState(agentUnit.State), agentUnit.Config, agentUnit.ConfigStateIdx) {
				c.unitsCh <- UnitChanged{
					Type: UnitChangedModified,
					Unit: unit,
				}
			}
		}
	}
}

// findUnit finds an existing unit.
func (c *clientV2) findUnit(id string, unitType UnitType) *Unit {
	for _, unit := range c.units {
		if unit.id == id && unit.unitType == unitType {
			return unit
		}
	}
	return nil
}

// unitChanged triggers the send goroutine to send a new observed state
func (c *clientV2) unitChanged() {
	if len(c.kickCh) <= 0 {
		c.kickCh <- struct{}{}
	}
}

// startActions starts the go routines to send and receive actions
//
// This starts 3 go routines to manage the actions bi-directional stream. The first
// go routine starts the stream then starts one go routine to receive messages and
// another go routine to send messages. The first go routine then blocks waiting on
// the receive and send to finish, then restarts the stream or exits if the context
// has been cancelled.
func (c *clientV2) startActions() {
	c.wg.Add(1)

	// results are held outside of the retry loop, because on re-connect
	// we still want to send the responses that either failed or haven't been
	// sent back to the agent.
	actionResults := make(chan *proto.ActionResponse, 100)
	go func() {
		defer c.wg.Done()
		for {
			select {
			case <-c.ctx.Done():
				// stopped
				return
			default:
			}

			c.actionRoundTrip(actionResults)
		}
	}()
}

func (c *clientV2) actionRoundTrip(actionResults chan *proto.ActionResponse) {
	actionsCtx, actionsCancel := context.WithCancel(c.ctx)
	defer actionsCancel()
	actionsClient, err := c.client.Actions(actionsCtx)
	if err != nil {
		c.errCh <- err
		return
	}

	var actionsRead sync.WaitGroup
	var actionsWrite sync.WaitGroup
	done := make(chan bool)

	// action requests
	actionsRead.Add(1)
	go func() {
		defer actionsRead.Done()
		for {
			action, err := actionsClient.Recv()
			if err != nil {
				if !errors.Is(err, io.EOF) {
					c.errCh <- err
				}
				close(done)
				return
			}

			switch action.Type {
			case proto.ActionRequest_CUSTOM:
				c.tryPerformAction(actionResults, action)
			case proto.ActionRequest_DIAGNOSTICS:
				// TODO: Implement the diagnostics action.
				// At the moment it just returns action type unknown until implemented.
				actionResults <- &proto.ActionResponse{
					Token:  c.token,
					Id:     action.Id,
					Status: proto.ActionResponse_FAILED,
					Result: ActionTypeUnknown,
				}
			default:
				actionResults <- &proto.ActionResponse{
					Token:  c.token,
					Id:     action.Id,
					Status: proto.ActionResponse_FAILED,
					Result: ActionTypeUnknown,
				}
			}
		}
	}()

	// action responses
	actionsWrite.Add(1)
	go func() {
		defer actionsWrite.Done()

		// initial connection of stream must send the token so
		// the Elastic Agent knows this clients token.
		err := actionsClient.Send(&proto.ActionResponse{
			Token:  c.token,
			Id:     ActionResponseInitID,
			Status: proto.ActionResponse_SUCCESS,
			Result: []byte("{}"),
		})
		if err != nil {
			c.errCh <- err
			return
		}

		for {
			select {
			case <-done:
				return
			case res := <-actionResults:
				err := actionsClient.Send(res)
				if err != nil {
					// failed to send, add back to response to try again
					actionResults <- res
					c.errCh <- err
					return
				}
			}
		}
	}()

	// wait for write goroutine to quit then close the client
	actionsWrite.Wait()
	actionsClient.CloseSend()

	// then wait for the read to quit
	actionsRead.Wait()
}

func (c *clientV2) tryPerformAction(actionResults chan *proto.ActionResponse, action *proto.ActionRequest) {
	// find the unit
	c.unitsMu.RLock()
	unit := c.findUnit(action.UnitId, UnitType(action.UnitType))
	c.unitsMu.RUnlock()
	if unit == nil {
		actionResults <- &proto.ActionResponse{
			Token:  c.token,
			Id:     action.Id,
			Status: proto.ActionResponse_FAILED,
			Result: ActionErrUnitNotFound,
		}
		return
	}

	// find the action registered with the unit
	unit.amx.RLock()
	actionImpl, ok := unit.actions[action.Name]
	unit.amx.RUnlock()
	if !ok {
		actionResults <- &proto.ActionResponse{
			Token:  c.token,
			Id:     action.Id,
			Status: proto.ActionResponse_FAILED,
			Result: ActionErrUndefined,
		}
		return
	}

	// ensure that the parameters can be unmarshalled
	var params map[string]interface{}
	err := json.Unmarshal(action.Params, &params)
	if err != nil {
		actionResults <- &proto.ActionResponse{
			Token:  c.token,
			Id:     action.Id,
			Status: proto.ActionResponse_FAILED,
			Result: ActionErrUnmarshableParams,
		}
		return
	}

	// perform the action (in goroutine)
	go func() {
		res, err := actionImpl.Execute(c.ctx, params)
		if err != nil {
			actionResults <- &proto.ActionResponse{
				Token:  c.token,
				Id:     action.Id,
				Status: proto.ActionResponse_FAILED,
				Result: utils.JSONMustMarshal(map[string]string{
					"error": err.Error(),
				}),
			}
			return
		}
		resBytes, err := json.Marshal(res)
		if err != nil {
			// client-side error, should have been marshal-able
			c.errCh <- err
			actionResults <- &proto.ActionResponse{
				Token:  c.token,
				Id:     action.Id,
				Status: proto.ActionResponse_FAILED,
				Result: ActionErrUnmarshableResult,
			}
			return
		}
		actionResults <- &proto.ActionResponse{
			Token:  c.token,
			Id:     action.Id,
			Status: proto.ActionResponse_SUCCESS,
			Result: resBytes,
		}
	}()
}

func inExpected(unit *Unit, expected []*proto.UnitExpected) bool {
	for _, agentUnit := range expected {
		if unit.id == agentUnit.Id && unit.unitType == UnitType(agentUnit.Type) {
			return true
		}
	}
	return false
}
