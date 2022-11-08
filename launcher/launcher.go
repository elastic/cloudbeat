// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package launcher

import (
	"fmt"
	"sync"
	"time"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/elastic/go-ucfg"
)

const (
	reconfigureWaitTimeout = 10 * time.Minute
)

type launcher struct {
	done      chan struct{}  // Channel used to initiate shutdown.
	wg        sync.WaitGroup // WaitGroup used to wait for active beaters
	stopped   bool
	cfgmtx    sync.Mutex
	runmtx    sync.Mutex
	beater    beat.Beater
	beaterErr chan error
	reloader  Reloader
	log       *logp.Logger
	latest    *config.C
	beat      *beat.Beat
	creator   beat.Creator
	validator Validator
}

type Reloader interface {
	Channel() <-chan *config.C
	Stop()
}

type Validator interface {
	Validate(*config.C) error
}

func New(log *logp.Logger, reloader Reloader, validator Validator, creator beat.Creator, cfg *config.C) (*launcher, error) {
	s := &launcher{
		done:      make(chan struct{}),
		wg:        sync.WaitGroup{},
		stopped:   false,
		cfgmtx:    sync.Mutex{},
		runmtx:    sync.Mutex{},
		log:       log,
		reloader:  reloader,
		validator: validator,
		beaterErr: make(chan error, 1),
		creator:   creator,
		latest:    cfg,
	}

	return s, nil
}

func (l *launcher) Run(b *beat.Beat) error {
	// Configure the beats Manager to start after all the reloadable hooks are initialized
	// and shutdown when the function returns.
	l.beat = b
	if err := b.Manager.Start(); err != nil {
		return err
	}
	defer b.Manager.Stop()

	// Wait for Fleet-side reconfiguration only if beater is running in Agent-managed mode.
	if b.Manager.Enabled() {
		l.log.Infof("Waiting for initial reconfiguration from Fleet server...")
		update, err := l.reconfigureWait(reconfigureWaitTimeout)
		if err != nil {
			l.log.Errorf("Failed while waiting for initial reconfiguraiton from Fleet server: %v", err)
			return err
		}

		if err := l.mergeConfig(update); err != nil {
			return fmt.Errorf("failed to update with initial reconfiguration from Fleet server: %w", err)
		}
	}

	err := l.run()
	return err
}

func (l *launcher) run() error {
	l.log.Info("Launcher is running")
	err := l.runBeater()
	if err != nil {
		l.log.Errorf("Launcher could not run Beater: %v", err)
		return err
	}

	go l.waitForUpdates()
	err = l.waitForFinish()
	if err != nil {
		l.log.Errorf("Launcher has stopped: %v", err)
	} else {
		l.log.Info("Launcher was shutted down gracefully")
	}
	return err
}

func (l *launcher) Stop() {
	l.log.Info("Launcher is about to shut down gracefully")

	// Make sure not to interrupt to an update
	l.cfgmtx.Lock()
	defer l.cfgmtx.Unlock()

	l.runmtx.Lock()
	defer l.runmtx.Unlock()

	// Stop listening for updates
	l.reloader.Stop()

	// Stop the beater
	l.stopBeater()

	// Trigger waitForFinish to stop
	close(l.done)
	l.stopped = true
}

func (l *launcher) runBeater() error {
	l.runmtx.Lock()
	defer l.runmtx.Unlock()
	if l.stopped {
		return nil
	}

	l.log.Info("Launcher is creating a new Beater")
	beater, err := l.creator(l.beat, l.latest)
	if err != nil {
		return fmt.Errorf("could not create beater: %w", err)
	}

	l.wg.Add(1)
	go func(beater beat.Beater) {
		l.log.Info("Launcher is running the new created Beater")
		defer func() {
			if r := recover(); r != nil {
				l.beaterErr <- fmt.Errorf("beater panic recovered: %s", r)
			}
		}()
		defer l.wg.Done()
		l.beaterErr <- beater.Run(l.beat)
		l.log.Info("Beater run has finished")
	}(beater)

	l.beater = beater
	return nil
}

// stopBeater only returns after the beater truely stopped running
func (l *launcher) stopBeater() {
	l.log.Info("Launcher is shutting the Beater down gracefully")
	l.beater.Stop()

	// By waiting to the wait group, it make sure that the old beater has really stopped
	l.wg.Wait()
	l.log.Info("Launcher shutted the Beater down gracefully")
}

// waitForFinish is the function that keeps Launcher up and running
// It will finish for one of two reasons:
// 1. The context is done, that probably means that the Stop function got called
// 2. The beater run has failed and returned an error
// Note that if the beater returned nil the launcher will keep running (and a new beater should be up again)
func (l *launcher) waitForFinish() error {
	for {
		select {
		case <-l.done:
			return nil

		case err := <-l.beaterErr:
			if err != nil {
				l.reloader.Stop()
				return fmt.Errorf("beater returned an error:  %w", err)
			}
		}
	}
}

func (l *launcher) waitForUpdates() {
	for {
		update, ok := <-l.reloader.Channel()
		if !ok {
			l.log.Info("Launcher has stopped waiting for updates")
			return
		}

		if err := l.configUpdate(update); err != nil {
			l.beaterErr <- fmt.Errorf("failed to update Beater config: %w", err)
		}
	}
}

// configUpdate applies incoming reconfiguration from the Fleet server to the beater config,
// and recreate the beater with the new values.
func (l *launcher) configUpdate(update *config.C) error {
	l.cfgmtx.Lock()
	defer l.cfgmtx.Unlock()
	if l.stopped {
		return nil
	}

	l.log.Infof("Got config update from fleet with %d keys", len(update.FlattenedKeys()))

	err := l.mergeConfig(update)
	if err != nil {
		return err
	}

	l.log.Infof("Restart the Beater with the new configuration of %d keys", len(l.latest.FlattenedKeys()))
	l.stopBeater()
	return l.runBeater()
}

func (l *launcher) mergeConfig(update *config.C) error {
	return l.latest.MergeWithOpts(update, ucfg.ReplaceArrValues)
}

// reconfigureWait will wait for and consume incoming reconfuration from the Fleet server, and keep
// discarding them until the incoming config contains the necessary information to start beater
// properly, thereafter returning the valid config.
func (l *launcher) reconfigureWait(timeout time.Duration) (*config.C, error) {
	start := time.Now()
	timer := time.After(timeout)

	for {
		select {
		case <-l.done:
			return nil, fmt.Errorf("cancelled via context")

		case <-timer:
			return nil, fmt.Errorf("timed out waiting for reconfiguration after %s", time.Since(start))

		case update, ok := <-l.reloader.Channel():
			if !ok {
				return nil, fmt.Errorf("reconfiguration channel is closed")
			}

			if l.validator != nil {
				err := l.validator.Validate(update)
				if err != nil {
					l.log.Errorf("Config update validation failed: %v", err)
					continue
				}
			}

			l.log.Infof("Received valid reconfiguration after waiting for %s", time.Since(start))
			return update, nil
		}
	}
}
