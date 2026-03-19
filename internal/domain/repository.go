package domain

import "context"

type ShipmentRepository interface {
	Create(ctx context.Context, s *Shipment, initialEvent *Event) error
	GetByID(ctx context.Context, id string) (*Shipment, error)
	UpdateStatusWithEvent(ctx context.Context, id string, status Status, event *Event) error
	GetHistory(ctx context.Context, shipmentID string) ([]*Event, error)
}
