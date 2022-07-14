// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package client

import (
	"time"

	"github.com/elastic/elastic-agent-client/v7/pkg/utils"
)

// CheckinMinimumTimeout is the amount of time the client must send a new checkin even if the status has not changed.
const CheckinMinimumTimeout = time.Second * 25

// InitialConfigIdx is the initial configuration index the client starts with. 0 represents no config state.
const InitialConfigIdx = 0

// ActionResponseInitID is the initial ID sent to Agent on first connect.
const ActionResponseInitID = "init"

// ActionErrUndefined is returned to Elastic Agent as result to an action request
// when the request action is not registered in the client.
var ActionErrUndefined = utils.JSONMustMarshal(map[string]string{
	"error": "action undefined",
})

// ActionErrUnmarshableParams is returned to Elastic Agent as result to an action request
// when the request params could not be un-marshaled to send to the action.
var ActionErrUnmarshableParams = utils.JSONMustMarshal(map[string]string{
	"error": "action params failed to be un-marshaled",
})

// ActionErrInvalidParams is returned to Elastic Agent as result to an action request
// when the request params are invalid for the action.
var ActionErrInvalidParams = utils.JSONMustMarshal(map[string]string{
	"error": "action params invalid",
})

// ActionErrUnmarshableResult is returned to Elastic Agent as result to an action request
// when the action was performed but the response could not be marshalled to send back to
// the agent.
var ActionErrUnmarshableResult = utils.JSONMustMarshal(map[string]string{
	"error": "action result failed to be marshaled",
})

// ActionErrUnitNotFound is returned to Elastic Agent as result to an action request
// when the request action unit cannot be found.
var ActionErrUnitNotFound = utils.JSONMustMarshal(map[string]string{
	"error": "action unit not found",
})

// ActionTypeUnknown is returned to Elastic Agent as result to an action request
// where the action type is unknown.
var ActionTypeUnknown = utils.JSONMustMarshal(map[string]string{
	"error": "action type unknown",
})
