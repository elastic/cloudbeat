// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package ecs

import "time"

const (
	OutcomeSuccess = "success"
	OutcomeFailure = "failure"
	OutcomeUnknown = "unknown"

	CategoryAuthentication     = "authentication"
	CategoryConfiguration      = "configuration"
	CategoryDatabase           = "database"
	CategoryDriver             = "driver"
	CategoryEmail              = "email"
	CategoryFile               = "file"
	CategoryHost               = "host"
	CategoryIAM                = "iam"
	CategoryIntrusionDetection = "intrusion_detection"
	CategoryMalware            = "malware"
	CategoryPackage            = "package"
	CategoryProcess            = "process"
	CategoryRegistry           = "registry"
	CategorySession            = "session"
	CategoryThreat             = "threat"
	CategoryWeb                = "web"

	TypeAccess       = "access"
	TypeAdmin        = "admin"
	TypeAllowed      = "allowed"
	TypeChange       = "change"
	TypeConnection   = "connection"
	TypeCreation     = "creation"
	TypeDeletion     = "deletion"
	TypeDenied       = "denied"
	TypeEnd          = "end"
	TypeError        = "error"
	TypeGroup        = "group"
	TypeIndication   = "indicator"
	TypeInfo         = "info"
	TypeInstallation = "installation"
	TypeProtocol     = "protocol"
	TypeStart        = "start"
	TypeUser         = "user"

	KindAlert         = "alert"
	KindEnrichment    = "enrichment"
	KindEvent         = "event"
	KindMetric        = "metric"
	KindState         = "state"
	KindPipelineError = "pipeline_error"
	KindSignal        = "signal"
)

// Event According to https://www.elastic.co/guide/en/ecs/current/ecs-event.html
// event.ingested property is not part of this struct as the fleet server setting it
type Event struct {
	// Represents the "big buckets" of ECS categories.
	// For example, filtering on event.category:process yields all events relating to process activity.
	// This field is closely related to event.type, which is used as a subcategory.
	Category []string `json:"category"`
	// Distinct from @timestamp in that @timestamp typically contain the time extracted from the original event.
	Created time.Time `json:"time"`
	// Unique ID to describe the event.
	ID string `json:"id"`
	// High-level information about what type of information the event contains, without being specific to the contents of the event.
	// For example, values of this field distinguish alert events from metric events.
	Kind string `json:"kind"`
	// The sequence number is a value published by some event sources, to make the exact ordering of events unambiguous,
	// regardless of the timestamp precision.
	Sequence int64 `json:"sequence"`
	// Denotes whether the event represents a success or a failure from the perspective of the entity that produced the event.
	Outcome string `json:"outcome"`
	// A categorization "sub-bucket" that, when used along with the event.category field values,
	// enables filtering events down to a level appropriate for single visualization.
	Type []string `json:"type"`
}
