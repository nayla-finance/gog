package registry

import "github.com/PROJECT_NAME/internal/domains/model"

func (r *Registry) SendSignal(signal model.SignalPayload) {
	if r.signal == nil {
		r.Logger().Debug("⚠️ Signal channel is not initialized, skipping signal")
		return
	}

	select {
	case r.signal <- signal:
	default:
		r.Logger().Warn("⚠️  Dropping signal – listener not ready")
	}
}

func (r *Registry) RegisterSignalListener() {
	go func() {
		for signal := range r.signal {
			r.Logger().Debugw("Received signal", "type", signal.Type)
			switch signal.Type {
			case model.SignalTypeNatsConsumerRestart:
				if !r.NatsService().HealthCheck() {
					r.Logger().Debug("❌ NATS connection is not healthy, reconnecting")
					if err := r.NatsService().Reconnect(); err != nil {
						r.Logger().Errorw("❌ Failed to reconnect to NATS", "error", err)
						continue
					}

					r.Logger().Debug("✅ NATS connection reestablished")
				}

				r.Logger().Debug("✅ NATS consumer restart requested")
				if err := r.RegisterConsumers(); err != nil {
					r.Logger().Errorw("❌ Failed to register consumers", "error", err)
					continue
				}

				r.Logger().Debug("✅ NATS consumer restart successful")
			}
		}

		r.Logger().Debug("✅ Signal listener stopped")
	}()
}
