package stopper

import (
	"context"
	"github.com/dipdup-io/workerpool"
	"github.com/dipdup-net/indexer-sdk/pkg/modules"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	InputName = "signal"
)

// Module - cancels context of all application if get signal.
//
//	                |----------------|
//	                |                |
//	-- struct{} ->  |     MODULE     |
//	                |                |
//	                |----------------|
type Module struct {
	input *modules.Input
	stop  context.CancelFunc
	log   zerolog.Logger
	g     workerpool.Group
}

func NewModule(cancelFunc context.CancelFunc) Module {
	m := Module{
		input: modules.NewInput(InputName),
		stop:  cancelFunc,
		g:     workerpool.NewGroup(),
	}
	m.log = log.With().Str("module", m.Name()).Logger()

	return m
}

func (*Module) Name() string {
	return "stopper"
}

// Start -
func (s *Module) Start(ctx context.Context) {
	s.g.GoCtx(ctx, s.listen)
}

func (s *Module) listen(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.input.Listen():
			log.Info().Msg("stop signal received")
			if s.stop != nil {
				log.Info().Msg("cancelling global indexer context...")
				s.stop()
			}
		}
	}
}

// Close -
func (s *Module) Close() error {
	s.g.Wait()
	return s.input.Close()
}

// Output -
func (*Module) Output(name string) (*modules.Output, error) {
	return nil, errors.Wrap(modules.ErrUnknownOutput, name)
}

// Input -
func (s *Module) Input(name string) (*modules.Input, error) {
	if name != InputName {
		return nil, errors.Wrap(modules.ErrUnknownInput, name)
	}
	return s.input, nil
}

// AttachTo -
func (s *Module) AttachTo(name string, input *modules.Input) error {
	output, err := s.Output(name)
	if err != nil {
		return err
	}

	output.Attach(input)
	return nil
}