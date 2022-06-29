package rerun

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

func NewLauncher(ctx context.Context,
	log *logp.Logger,
	reloader Reloader,
	validator Validator,
	bt *beat.Beat,
	creator beat.Creator,
	cfg *config.C) (*launcher, error) {
	s := &launcher{
		wg:        sync.WaitGroup{},
		ctx:       ctx,
		log:       log,
		reloader:  reloader,
		validator: validator,
		beaterErr: make(chan error, 1),
		creator:   creator,
		latest:    cfg,
	}

	return s, nil
}

func (s *launcher) Run(b *beat.Beat) error {
	// Configure the beats Manager to start after all the reloadable hooks are initialized
	// and shutdown when the function returns.
	if err := b.Manager.Start(); err != nil {
		return err
	}
	defer b.Manager.Stop()

	// Wait for Fleet-side reconfiguration only if beater is running in Agent-managed mode.
	if b.Manager.Enabled() {
		s.log.Infof("Waiting for initial reconfiguration from Fleet server...")
		update, err := s.reconfigureWait(reconfigureWaitTimeout)
		if err != nil {
			s.log.Errorf("Failed while waiting for initial reconfiguraiton from Fleet server: %v", err)
			return err
		}

		if err := s.configUpdate(update); err != nil {
			return fmt.Errorf("failed to update with initial reconfiguration from Fleet server: %w", err)
		}
	}

	return s.run()
}

func (s *launcher) run() error {
	err := s.runBeater()
	if err != nil {
		return err
	}

	err = s.waitForUpdates()
	s.log.Error("Beater starter is stopping: %w", err)
	return err
}

func (s *launcher) Stop() {
	s.stopBeater()
}

func (s *launcher) runBeater() error {
	beater, err := s.creator(s.beat, s.latest)
	if err != nil {
		return fmt.Errorf("Could not create beater: %w", err)
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.beaterErr <- beater.Run(s.beat)
	}()

	s.beater = beater
	return nil
}

func (s *launcher) stopBeater() {
	s.beater.Stop()
	s.wg.Wait()
}

func (s *launcher) waitForUpdates() error {
	for {
		select {
		case <-s.ctx.Done():
			s.stopBeater()
			return nil

		case err := <-s.beaterErr:
			if err != nil {
				return fmt.Errorf("Beater returned an error:  %w", err)
			}

		case update, ok := <-s.reloader.Channel():
			if !ok {
				return errors.New("Reloader channel closed")
			}

			go func() {
				if err := s.configUpdate(update); err != nil {
					s.log.Errorf("Failed to update beater config: %v", err)
				}
			}()
		}
	}
}

// configUpdate applies incoming reconfiguration from the Fleet server to the beater config,
// and recreate the beater with the new values.
func (s *launcher) configUpdate(update *config.C) error {
	s.log.Info("Got config update")

	err := s.latest.MergeWithOpts(update, ucfg.ReplaceArrValues)
	if err != nil {
		return err
	}

	s.stopBeater()
	return s.runBeater()
}

// reconfigureWait will wait for and consume incoming reconfuration from the Fleet server, and keep
// discarding them until the incoming config contains the necessary information to start beater
// properly, thereafter returning the valid config.
func (s *launcher) reconfigureWait(timeout time.Duration) (*config.C, error) {
	start := time.Now()
	timer := time.After(timeout)

	for {
		select {
		case <-s.ctx.Done():
			return nil, fmt.Errorf("cancelled via context")

		case <-timer:
			return nil, fmt.Errorf("timed out waiting for reconfiguration after %s", time.Since(start))

		case update, ok := <-s.reloader.Channel():
			if !ok {
				return nil, fmt.Errorf("reconfiguration channel is closed")
			}

			if s.validator != nil {
				err := s.validator.Validate(update)
				if err != nil {
					s.log.Errorf("Config update validation failed: %w", err)
					continue
				}
			}

			s.log.Infof("Received valid reconfiguration after waiting for %s", time.Since(start))
			return update, nil
		}
	}
}
