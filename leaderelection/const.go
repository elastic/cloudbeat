package leaderelection

import (
	"time"
)

const (
	// LeaseDuration is the duration that non-leader candidates will
	// wait to force acquire leadership. This is measured against time of
	// last observed ack.
	//
	LeaseDuration = 15 * time.Second
	// RenewDeadline is the duration that the acting manager will retry
	// refreshing leadership before giving up.
	//
	RenewDeadline = 10 * time.Second

	// RetryPeriod is the duration the LeaderElector clients should wait
	// between tries of actions.
	//
	RetryPeriod = 2 * time.Second

	PodNameEnvar = "POD_NAME"

	LeaderLeaseName = "cloudbeat-cluster-leader"
)
