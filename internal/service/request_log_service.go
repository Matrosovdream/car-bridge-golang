package service

import (
	"context"

	"github.com/sirupsen/logrus"

	"car-bridge/internal/model"
	"car-bridge/internal/model/converter"
	"car-bridge/internal/repository"
)

// RequestLogService records the outcome (and cost) of each upstream API call.
type RequestLogService struct {
	Log  *logrus.Logger
	Repo *repository.RequestLogRepository
}

// NewRequestLogService wires the service with its repository.
func NewRequestLogService(log *logrus.Logger, repo *repository.RequestLogRepository) *RequestLogService {
	return &RequestLogService{Log: log, Repo: repo}
}

// Record persists one API-call entry. Best-effort: a logging failure is logged
func (s *RequestLogService) Record(ctx context.Context, entry model.APILogEntry) {
	ent := converter.LogEntryToEntity(entry)
	if err := s.Repo.Insert(ctx, ent); err != nil {
		s.Log.WithError(err).WithFields(logrus.Fields{
			"service_type": entry.ServiceType,
			"service_code": entry.ServiceCode,
			"operation":    entry.Operation,
		}).Warn("failed to record API request log")
	}
}
