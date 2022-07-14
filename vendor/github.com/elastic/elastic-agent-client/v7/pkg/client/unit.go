// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package client

import (
	"encoding/json"
	"sync"

	"github.com/elastic/elastic-agent-client/v7/pkg/proto"
)

// UnitType is the type of the unit, either input or output
type UnitType proto.UnitType

const (
	// UnitTypeInput is an input unit.
	UnitTypeInput = UnitType(proto.UnitType_INPUT)
	// UnitTypeOutput is an output unit.
	UnitTypeOutput = UnitType(proto.UnitType_OUTPUT)
)

// UnitState is the state for the unit, used both for expected and observed state.
type UnitState proto.State

const (
	// UnitStateStarting is when a unit is starting.
	UnitStateStarting = UnitState(proto.State_STARTING)
	// UnitStateConfiguring is when a unit is currently configuring.
	UnitStateConfiguring = UnitState(proto.State_CONFIGURING)
	// UnitStateHealthy is when the unit is working exactly as it should.
	UnitStateHealthy = UnitState(proto.State_HEALTHY)
	// UnitStateDegraded is when the unit is working but not exactly as its expected.
	UnitStateDegraded = UnitState(proto.State_DEGRADED)
	// UnitStateFailed is when the unit is completely broken and failing to work.
	UnitStateFailed = UnitState(proto.State_FAILED)
	// UnitStateStopping is when the unit is stopping.
	UnitStateStopping = UnitState(proto.State_STOPPING)
	// UnitStateStopped is when the unit is stopped.
	UnitStateStopped = UnitState(proto.State_STOPPED)
)

// Unit represents a distinct item that needs to be operating with-in this process.
//
// This is normally N number of inputs and 1 output (possible for multiple in the future).
type Unit struct {
	id       string
	unitType UnitType

	expMu     sync.RWMutex
	exp       UnitState
	config    string
	configIdx uint64

	stateMu             sync.RWMutex
	state               UnitState
	stateMsg            string
	statePayload        map[string]interface{}
	statePayloadEncoded json.RawMessage

	amx     sync.RWMutex
	actions map[string]Action

	client *clientV2
}

// ID of the unit.
func (u *Unit) ID() string {
	return u.id
}

// Type of the unit.
func (u *Unit) Type() UnitType {
	return u.unitType
}

// Expected returns the expected state and config for the unit.
func (u *Unit) Expected() (UnitState, string) {
	u.expMu.RLock()
	defer u.expMu.RUnlock()
	return u.exp, u.config
}

// State returns the currently reported state for the unit.
func (u *Unit) State() (UnitState, string, map[string]interface{}) {
	u.stateMu.RLock()
	defer u.stateMu.RUnlock()
	return u.state, u.stateMsg, u.statePayload
}

// UpdateState updates the state for the unit.
func (u *Unit) UpdateState(state UnitState, message string, payload map[string]interface{}) error {
	var encoded json.RawMessage
	var err error
	if payload != nil {
		encoded, err = json.Marshal(payload)
		if err != nil {
			return err
		}
	}
	u.stateMu.Lock()
	defer u.stateMu.Unlock()
	changed := false
	if u.state != state {
		u.state = state
		changed = true
	}
	if u.stateMsg != message {
		u.stateMsg = message
		changed = true
	}
	u.statePayload = payload
	if (u.statePayloadEncoded == nil && encoded != nil) || (u.statePayloadEncoded != nil && encoded == nil) || (string(u.statePayloadEncoded) != string(encoded)) {
		u.statePayloadEncoded = encoded
		changed = true
	}
	if changed {
		u.client.unitChanged()
	}
	return nil
}

// RegisterAction registers action handler for this unit.
func (u *Unit) RegisterAction(action Action) {
	u.amx.Lock()
	defer u.amx.Unlock()
	u.actions[action.Name()] = action
}

// UnregisterAction unregisters action handler with the client
func (u *Unit) UnregisterAction(action Action) {
	u.amx.Lock()
	defer u.amx.Unlock()
	delete(u.actions, action.Name())
}

// GetAction finds an action by its name.
func (u *Unit) GetAction(name string) (Action, bool) {
	u.amx.RLock()
	defer u.amx.RUnlock()
	act, ok := u.actions[name]
	return act, ok
}

// Store returns the store client.
func (u *Unit) Store() StoreClient {
	return &storeClient{
		client:   u.client,
		unitID:   u.id,
		unitType: u.unitType,
	}
}

// Artifacts returns the artifacts client.
func (u *Unit) Artifacts() ArtifactsClient {
	return u.client.Artifacts()
}

// Logger returns the log client.
func (u *Unit) Logger() LogClient {
	return &logClient{
		client:   u.client,
		unitID:   u.id,
		unitType: u.unitType,
	}
}

// updateConfig updates the configuration for this unit, triggering the delegate function if set.
func (u *Unit) updateState(exp UnitState, cfg string, cfgIdx uint64) bool {
	u.expMu.Lock()
	defer u.expMu.Unlock()
	changed := false
	if u.exp != exp {
		u.exp = exp
		changed = true
	}
	if u.configIdx != cfgIdx {
		u.configIdx = cfgIdx
		if u.config != cfg {
			u.config = cfg
			changed = true
		}
	}
	return changed
}

// toObserved returns the observed unit protocol to send over the stream.
func (u *Unit) toObserved() *proto.UnitObserved {
	u.expMu.RLock()
	cfgIdx := u.configIdx
	u.expMu.RUnlock()
	u.stateMu.RLock()
	defer u.stateMu.RUnlock()
	return &proto.UnitObserved{
		Id:             u.id,
		Type:           proto.UnitType(u.unitType),
		ConfigStateIdx: cfgIdx,
		State:          proto.State(u.state),
		Message:        u.stateMsg,
		Payload:        u.statePayloadEncoded,
	}
}

// newUnit creates a new unit that needs to be created in this process.
func newUnit(id string, unitType UnitType, exp UnitState, cfg string, cfgIdx uint64, client *clientV2) *Unit {
	return &Unit{
		id:        id,
		unitType:  unitType,
		config:    cfg,
		configIdx: cfgIdx,
		exp:       exp,
		state:     UnitStateStarting,
		stateMsg:  "Starting",
		client:    client,
		actions:   make(map[string]Action),
	}
}
