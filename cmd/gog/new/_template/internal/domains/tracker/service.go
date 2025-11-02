package tracker

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	"github.com/nayla-finance/go-nayla/logger"
	"github.com/nayla-finance/go-nayla/nats"
)

var _ Service = &svc{}

type (
	Service interface {
		PublishCallCompleted(ctx context.Context, dto SaveCallDto)
		saveCall(ctx context.Context, dto SaveCallDto) error
	}

	ServiceProvider interface {
		TrackerService() Service
	}

	svcDependencies interface {
		RepositoryProvider
		nats.ServiceProvider
		logger.Provider
	}

	svc struct {
		d svcDependencies
	}
)

func NewService(d svcDependencies) Service {
	return &svc{d: d}
}

// This method should not return an error to avoid interrupting the flow of the request.
// it'll log it and continue.
func (s *svc) PublishCallCompleted(ctx context.Context, dto SaveCallDto) {
	if err := s.d.NatsService().Publish(ctx, SubjectCallCompleted, dto); err != nil {
		s.d.Logger().Errorw(ctx, "failed to publish request to nats", "error", err, "dto", dto)
		sentry.CaptureException(err)
	}
}

func (s *svc) saveCall(ctx context.Context, dto SaveCallDto) error {
	s.d.Logger().Debugw(ctx, "🔍 Saving call", "dto", dto)

	id, err := uuid.NewV7()
	if err != nil {
		s.d.Logger().Errorw(ctx, "failed to generate request ID", "error", err)
		return err
	}

	if dto.RequestID != nil {
		id = *dto.RequestID
	}

	tracker := &Tracker{
		ID:             id,
		IsSuccess:      dto.IsSuccess,
		Path:           dto.Path,
		Method:         dto.Method,
		RequestBody:    dto.ReqBody,
		ResponseBody:   string(dto.RespBody),
		ResponseTimeMs: dto.ResponseTime.Milliseconds(),
	}

	if err := s.d.TrackerRepository().Create(ctx, tracker); err != nil {
		s.d.Logger().Errorw(ctx, "failed to save tracker", "error", err, "tracker", tracker)
		return err
	}

	tracker, err = s.d.TrackerRepository().GetByID(ctx, id)
	if err != nil {
		s.d.Logger().Errorw(ctx, "failed to get tracker", "error", err, "id", id)
		return err
	}

	if err := s.d.NatsService().Publish(ctx, SubjectCallTracked, tracker); err != nil {
		s.d.Logger().Errorw(ctx, "failed to publish tracker to nats", "error", err, "tracker", tracker)
		return err
	}

	return nil
}
