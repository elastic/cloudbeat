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
	"context"
	"errors"
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
	wg        sync.WaitGroup
	beater    beat.Beater
	beaterErr chan error
	reloader  Reloader
	log       *logp.Logger
	ctx       context.Context
	cancel    context.CancelFunc
	latest    *config.C
	beat      *beat.Beat
	creator   beat.Creator
	validator Validator
}

type Reloader interface {
	Channel() <-chan *config.C
}

type Validator interface {
	Validate(*config.C) error
}

func New(ctx context.Context, log *logp.Logger, reloader Reloader, validator Validator, creator beat.Creator, cfg *config.C) (*launcher, error) {
	cctx, cancel := context.WithCancel(ctx)
	s := &launcher{
		wg:        sync.WaitGroup{},
		ctx:       cctx,
		cancel:    cancel,
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
	l.log.Info("Beater launcher is running")
	err := l.runBeater()
	if err != nil {
		l.log.Errorf("Could not run Beater: %v", err)
		return err
	}

	err = l.waitForUpdates()
	if err != nil {
		l.log.Errorf("Beater launcher has stopped: %v", err)
	} else {
		l.log.Info("Beater launcher was shutted down gracefully")
	}
	return err
}

func (l *launcher) Stop() {
	l.log.Info("Beater launcher is about to shut down gracefully")
	l.cancel()
}

func (l *launcher) runBeater() error {
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

func (l *launcher) waitForUpdates() error {
	for {
		select {
		case <-l.ctx.Done():
			l.stopBeater()
			return nil

		case err := <-l.beaterErr:
			if err != nil {
				return fmt.Errorf("beater returned an error:  %w", err)
			}

		case update, ok := <-l.reloader.Channel():
			if !ok {
				return errors.New("Reloader channel closed")
			}

			go func() {
				if err := l.configUpdate(update); err != nil {
					l.beaterErr <- fmt.Errorf("failed to update Beater config: %w", err)
					return
				}
			}()
		}
	}
}

// configUpdate applies incoming reconfiguration from the Fleet server to the beater config,
// and recreate the beater with the new values.
func (l *launcher) configUpdate(update *config.C) error {
	l.log.Infof("Got config update from fleet with %d keys", len(update.FlattenedKeys()))

	err := l.mergeConfig(update)
	if err != nil {
		return err
	}

	l.log.Infof("Creating a new Beater with the new configuration of %d keys", len(l.latest.FlattenedKeys()))
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
		case <-l.ctx.Done():
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
