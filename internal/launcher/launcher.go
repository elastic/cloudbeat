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
	"github.com/elastic/beats/v7/libbeat/management/status"
	"github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/go-ucfg"

	"github.com/elastic/cloudbeat/internal/infra/clog"
)

const (
	reconfigureWaitTimeout = 10 * time.Minute

	// Time to wait for the beater to stop before ignoring it
	shutdownGracePeriod = 20 * time.Second
)

// ErrStopSignal is used to indicate we got a stop signal
var ErrStopSignal = errors.New("stop beater")

// ErrGracefulExit is used when the launcher stops before shutdownGracePeriod, after waiting for the beater to stop
var ErrGracefulExit = beat.GracefulExit

// ErrTimeoutExit is used when the launcher stops after shutdownGracePeriod, without waiting for the beater to stop
var ErrTimeoutExit = errors.New("exit after timeout")

type launcher struct {
	wg        sync.WaitGroup // WaitGroup used to wait for active beaters
	beater    beat.Beater
	beaterErr chan error
	reloader  Reloader
	log       *clog.Logger
	latest    *config.C
	beat      *beat.Beat
	creator   beat.Creator
	validator Validator
	name      string
}

type Reloader interface {
	Channel() <-chan *config.C
	Stop()
}

type Validator interface {
	Validate(*config.C) error
}

func New(log *clog.Logger, name string, reloader Reloader, validator Validator, creator beat.Creator, cfg *config.C) beat.Beater {
	return &launcher{
		beaterErr: make(chan error, 1),
		wg:        sync.WaitGroup{},
		log:       log,
		name:      name,
		reloader:  reloader,
		validator: validator,
		creator:   creator,
		latest:    cfg,
	}
}

func (l *launcher) Run(b *beat.Beat) error {
	// Configure the beats Manager to start after all the reloadable hooks are initialized
	// and shutdown when the function returns.
	l.beat = b
	if err := b.Manager.Start(); err != nil {
		return err
	}

	// Wait for Fleet-side reconfiguration only if beater is running in Agent-managed mode.
	if b.Manager.Enabled() {
		defer b.Manager.Stop()
		l.log.Infof("Waiting for initial reconfiguration from Fleet server...")
		update, err := l.reconfigureWait(reconfigureWaitTimeout)
		if err != nil {
			l.log.Errorf("Failed while waiting for the initial reconfiguration from Fleet server: %v", err)
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

	switch {
	case errors.Is(err, ErrGracefulExit):
		l.log.Info("Launcher stopped successfully")
	case errors.Is(err, ErrTimeoutExit):
		l.log.Info("Launcher stopped after timeout")
	case err == nil: // unexpected
	default:
		l.log.Errorf("Launcher stopped by error: %v", err)
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
		// config update		(val, nil)
		// stop signal			(nil, ErrStopSignal)
		// beater error			(nil, err)
		cfg, err := l.waitForUpdates()

		if isConfigUpdate(cfg, err) {
			l.stopBeater()

			err = l.configUpdate(cfg)
			if err != nil {
				return fmt.Errorf("failed to update Beater config: %w", err)
			}
			l.log.Infof("Restart %s with the new configuration of %d keys", l.name, len(l.latest.FlattenedKeys()))
			continue
		}

		if isStopSignal(cfg, err) {
			return l.stopBeaterWithTimeout(shutdownGracePeriod)
		}

		if isBeaterError(cfg, err) {
			return err
		}
	}
}

func (l *launcher) Stop() {
	l.log.Info("Launcher is about to shut down gracefully")
	close(l.beaterErr)
}

// runBeater creates a new beater and starts a goroutine for running it.
// It is protected from panics and ship errors back to beaterErr
func (l *launcher) runBeater() error {
	l.log.Infof("Launcher is creating a new %s", l.name)
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

		l.log.Infof("Launcher is running the new created %s", l.name)
		err := l.beater.Run(l.beat)
		if err != nil {
			l.beaterErr <- fmt.Errorf("beater returned an error: %w", err)
		}
		l.log.Infof("%s run has finished", l.name)
	}()
	return nil
}

func (l *launcher) stopBeater() {
	l.log.Infof("Launcher is shutting %s down gracefully", l.name)
	l.beater.Stop()

	// By waiting to the wait group, we make sure that the old beater has really stopped
	l.wg.Wait()
	l.log.Infof("Launcher shut %s down gracefully", l.name)
}

// Returns an error indicating if the beater was stopped gracefully or not
func (l *launcher) stopBeaterWithTimeout(duration time.Duration) error {
	l.log.Infof("Launcher is shutting %s down gracefully", l.name)
	l.beater.Stop()

	wgCh := make(chan struct{})

	go func() {
		// By waiting to the wait group, we make sure that the old beater has really stopped
		l.wg.Wait()
		close(wgCh)
	}()

	select {
	case <-time.After(duration):
		l.log.Infof("Grace period for %s ended", l.name)
		return ErrTimeoutExit
	case <-wgCh:
		l.log.Infof("Launcher shut %s down gracefully", l.name)
		return ErrGracefulExit
	}
}

// waitForUpdates is the function that keeps Launcher runLoop busy.
// It will finish for one of following reasons:
//  1. The Stop function got called 	(nil, ErrStopSignal)
//  2. The beater run has returned 		(nil, err)
//  3. A config update received 		(val, nil)
func (l *launcher) waitForUpdates() (*config.C, error) {
	select {
	case err, ok := <-l.beaterErr:
		if !ok {
			l.log.Infof("Launcher received a stop signal")
			return nil, ErrStopSignal
		}
		return nil, err

	case update, ok := <-l.reloader.Channel():
		if !ok {
			return nil, errors.New("reloader channel unexpectedly closed")
		}

		l.log.Infof("Launcher will restart %s to apply the configuration update", l.name)
		return update, nil
	}
}

func isConfigUpdate(cfg *config.C, err error) bool {
	return cfg != nil && err == nil
}

func isStopSignal(cfg *config.C, err error) bool {
	return cfg == nil && errors.Is(err, ErrStopSignal)
}

func isBeaterError(cfg *config.C, err error) bool {
	return cfg == nil && err != nil
}

// configUpdate applies incoming reconfiguration from the Fleet server to the beater config
func (l *launcher) configUpdate(update *config.C) error {
	l.log.Infof("Merging config update from fleet with %d keys", len(update.FlattenedKeys()))

	return l.latest.MergeWithOpts(update, ucfg.ReplaceArrValues)
}

// reconfigureWait will wait for and consume incoming reconfiguration from the Fleet server, and keep
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
					healthErr := &BeaterUnhealthyError{}
					if errors.As(err, healthErr) {
						l.beat.Manager.UpdateStatus(status.Degraded, healthErr.Error())
					}
					continue
				}
			}

			l.log.Infof("Received valid reconfiguration after waiting for %s", time.Since(start))
			return update, nil
		}
	}
}
