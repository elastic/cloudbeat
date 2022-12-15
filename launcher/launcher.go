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
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/management"
	cloudbeat_config "github.com/elastic/cloudbeat/config"
	"github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/elastic/go-ucfg"
)

const (
	reconfigureWaitTimeout = 10 * time.Minute
)

type launcher struct {
	wg        sync.WaitGroup // WaitGroup used to wait for active beaters
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
		beaterErr: make(chan error, 1),
		wg:        sync.WaitGroup{},
		log:       log,
		reloader:  reloader,
		validator: validator,
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

		if err := l.configUpdate(update); err != nil {
			return fmt.Errorf("failed to update with initial reconfiguration from Fleet server: %w", err)
		}
	}

	err := l.run()
	return err
}

func (l *launcher) run() error {
	err := l.runLoop()
	if err != nil {
		l.log.Errorf("Launcher has stopped: %v", err)
	} else {
		l.log.Info("Launcher was shutted down gracefully")
	}

	l.reloader.Stop()
	return err
}

// runLoop is the loop that keeps the launcher alive
func (l *launcher) runLoop() error {
	l.log.Info("Launcher is running")
	for {
		// Run a new beater
		err := l.runBeater()
		if err != nil {
			return fmt.Errorf("launcher could not run Beater: %v", err)
		}

		// Wait for something to happen:
		// config update produces val, nil
		// signal produces nil, nil
		// beater error produces nil, err
		cfg, err := l.waitForUpdates()

		// If it's not a beater error, should stop the beater
		if err == nil {
			l.stopBeater()
		}

		// If it's a config update let's merge the new config and continue with the next iteration
		if cfg != nil {
			err = l.configUpdate(cfg)
			if err != nil {
				return fmt.Errorf("failed to update Beater config: %w", err)
			}
			l.log.Infof("Restart the Beater with the new configuration of %d keys", len(l.latest.FlattenedKeys()))
			continue
		}

		// If the beater produced an error, should bubble the error up
		return err
	}
}

func (l *launcher) Stop() {
	l.log.Info("Launcher is about to shut down gracefully")
	close(l.beaterErr)
}

// runBeater creates a new beater and starts a goroutine for running it
// It is protected from panics and ship errors back to beaterErr
func (l *launcher) runBeater() error {
	l.log.Info("Launcher is creating a new Beater")
	var err error
	l.beater, err = l.creator(l.beat, l.latest)
	if err != nil {
		return fmt.Errorf("could not create beater: %w", err)
	}

	l.wg.Add(1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				l.beaterErr <- fmt.Errorf("beater panic recovered: %s", r)
			}
		}()
		defer l.wg.Done()

		l.log.Info("Launcher is running the new created Beater")
		err := l.beater.Run(l.beat)
		if err != nil {
			l.beaterErr <- fmt.Errorf("beater returned an error: %w", err)
		}
		l.log.Info("Beater run has finished")
	}()
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

// waitForUpdates is the function that keeps Launcher runLoop busy
// It will finish for one of following reasons:
// 1. The Stop function got called
// 2. The beater run has returned
// 3. A config update received
func (l *launcher) waitForUpdates() (*config.C, error) {
	select {
	case err := <-l.beaterErr:
		return nil, err

	case update, ok := <-l.reloader.Channel():
		if !ok {
			return nil, fmt.Errorf("reloader channel unexpectedly closed")
		}

		return update, nil
	}
}

// configUpdate applies incoming reconfiguration from the Fleet server to the beater config
func (l *launcher) configUpdate(update *config.C) error {
	l.log.Infof("Got config update from fleet with %d keys", len(update.FlattenedKeys()))

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
		case <-l.beaterErr:
			return nil, fmt.Errorf("error channel closed")

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
					if errors.Is(err, cloudbeat_config.ErrBenchmarkNotSupported) {
						l.beat.Manager.UpdateStatus(management.Degraded, cloudbeat_config.ErrBenchmarkNotSupported.Error())
					}
					continue
				}
			}

			l.log.Infof("Received valid reconfiguration after waiting for %s", time.Since(start))
			return update, nil
		}
	}
}
