package nats

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func (s *svc) Setup() error {
	s.d.Logger().Debug("🔄 Setting up NATS connection...")

	nc, err := nats.Connect(s.d.Config().Nats.Servers, nats.UserCredentials(s.d.Config().Nats.CredsPath))
	if err != nil {
		s.d.Logger().Error("❌ Failed to connect to NATS", "error", err)
		return err
	}
	s.nc = nc
	s.d.Logger().Debug("Successfully connected to NATS ✅")

	js, err := jetstream.New(nc)
	if err != nil {
		s.d.Logger().Error("❌ Failed to initialize JetStream", "error", err)
		return err
	}
	s.js = js

	s.d.Logger().Debug("🔄Setting up streams...")
	if err = s.SetupStreams(); err != nil {
		s.d.Logger().Error("❌ Failed to setup streams", "error", err)
		return err
	}
	s.d.Logger().Debug("Successfully setup streams ✅")

	return nil
}

func (s *svc) SetupStreams() error {
	for _, stream := range s.cfg.Streams {
		s.d.Logger().Debug("🔄 Creating Or Updating stream...", " stream ", stream.Name)
		_, err := s.js.CreateOrUpdateStream(context.Background(), stream)
		if err != nil {
			s.d.Logger().Error("❌ Failed to create or update stream", " stream ", stream.Name, " error ", err)
			return err
		}
		s.d.Logger().Debug("✅ Stream ", stream.Name, " Configured Successfully")
	}
	return nil
}

// Cleanup gracefully closes NATS connections
func (s *svc) Cleanup() error {
	s.d.Logger().Info("🔄 Cleaning up NATS resources...")

	if s.nc != nil {
		// Wait for any pending messages to be processed
		if err := s.nc.Drain(); err != nil {
			s.d.Logger().Error("❌ Failed to drain NATS connections", "error", err)
			return err
		}
	}

	s.d.Logger().Info("✅ NATS cleanup completed")
	return nil
}
