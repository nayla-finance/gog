package registry

import (
	"context"

	"github.com/PROJECT_NAME/internal/domains/model"
)

func (r *Registry) SendSignal(signal model.SignalPayload) {
	if r.signal == nil {
		r.Logger().Debugw(context.Background(), "⚠️ Signal channel is not initialized, skipping signal")
		return
	}

	select {
	case r.signal <- signal:
	default:
		r.Logger().Warnw(context.Background(), "⚠️  Dropping signal – listener not ready")
	}
}

func (r *Registry) RegisterSignalListener() {
	go func() {
		for signal := range r.signal {
			ctx := context.Background()

			r.Logger().Debugw(ctx, "Received signal", "type", signal.Type)
			switch signal.Type {
			case model.SignalTypeNatsConsumerRestart:
				if !r.NatsService().Ping(ctx) {
					r.Logger().Debugw(ctx, "❌ NATS connection is not healthy, reconnecting")
					if err := r.NatsService().Reconnect(ctx); err != nil {
						r.Logger().Errorw(ctx, "❌ Failed to reconnect to NATS", "error", err)
						continue
					}

					r.Logger().Debugw(ctx, "✅ NATS connection reestablished")
				}

				r.Logger().Debugw(ctx, "✅ NATS consumer restart requested")
				if err := r.RegisterConsumers(); err != nil {
					r.Logger().Errorw(ctx, "❌ Failed to register consumers", "error", err)
					continue
				}

				r.Logger().Debugw(ctx, "✅ NATS consumer restart successful")
			}
		}

		r.Logger().Debugw(context.Background(), "✅ Signal listener stopped")
	}()
}
