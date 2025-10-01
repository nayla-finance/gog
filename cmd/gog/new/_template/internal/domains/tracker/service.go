package tracker

import (
	"context"

	"github.com/PROJECT_NAME/internal/logger"
	"github.com/PROJECT_NAME/internal/nats"
	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
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
		logger.LoggerProvider
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
		s.d.Logger().Errorw("failed to publish request to nats", "error", err, "dto", dto)
		sentry.CaptureException(err)
	}
}

func (s *svc) saveCall(ctx context.Context, dto SaveCallDto) error {
	s.d.Logger().Debugw("🔍 Saving call", "dto", dto)

	id, err := uuid.NewV7()
	if err != nil {
		s.d.Logger().Errorw("failed to generate request ID", "error", err)
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
		s.d.Logger().Errorw("failed to save tracker", "error", err, "tracker", tracker)
		return err
	}

	tracker, err = s.d.TrackerRepository().GetByID(ctx, id)
	if err != nil {
		s.d.Logger().Errorw("failed to get tracker", "error", err, "id", id)
		return err
	}

	if err := s.d.NatsService().Publish(ctx, SubjectCallTracked, tracker); err != nil {
		s.d.Logger().Errorw("failed to publish tracker to nats", "error", err, "tracker", tracker)
		return err
	}

	return nil
}
