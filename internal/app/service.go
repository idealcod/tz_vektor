package app

import (
	"context"

	"tz/internal/domain"
)

type ShipmentService struct {
	repo domain.ShipmentRepository
}

func NewShipmentService(repo domain.ShipmentRepository) *ShipmentService {
	return &ShipmentService{repo: repo}
}

func (s *ShipmentService) CreateShipment(ctx context.Context, ref, origin, dest, driver string, amount, revenue float64) (*domain.Shipment, error) {
	shipment, err := domain.NewShipment(ref, origin, dest, driver, amount, revenue)
	if err != nil {
		return nil, err
	}

	event := &domain.Event{
		ShipmentID: shipment.ID,
		Status:     domain.StatusPending,
		Message:    "shipment created",
	}

	if err := s.repo.Create(ctx, shipment, event); err != nil {
		return nil, err
	}

	return shipment, nil
}

func (s *ShipmentService) GetShipment(ctx context.Context, id string) (*domain.Shipment, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ShipmentService) AddStatusEvent(ctx context.Context, id string, newStatus domain.Status, msg string) (*domain.Event, error) {
	shipment, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	event, err := shipment.Transition(newStatus)
	if err != nil {
		return nil, err
	}
	event.Message = msg

	if err := s.repo.UpdateStatusWithEvent(ctx, id, newStatus, event); err != nil {
		return nil, err
	}

	return event, nil
}

func (s *ShipmentService) GetHistory(ctx context.Context, id string) ([]*domain.Event, error) {
	return s.repo.GetHistory(ctx, id)
}
