package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Shipment struct {
	ID              string
	ReferenceNumber string
	Origin          string
	Destination     string
	CurrentStatus   Status
	DriverDetails   string
	Amount          float64
	DriverRevenue   float64
	CreatedAt       time.Time
}

type Event struct {
	ID         string
	ShipmentID string
	Status     Status
	Message    string
	CreatedAt  time.Time
}

func NewShipment(ref, origin, dest, driver string, amount, revenue float64) (*Shipment, error) {
	shipment := &Shipment{
		ID:              uuid.NewString(),
		ReferenceNumber: ref,
		Origin:          origin,
		Destination:     dest,
		CurrentStatus:   StatusPending,
		DriverDetails:   driver,
		Amount:          amount,
		DriverRevenue:   revenue,
		CreatedAt:       time.Now(),
	}

	if err := shipment.Validate(); err != nil {
		return nil, err
	}

	return shipment, nil
}

func (s *Shipment) Validate() error {
	switch {
	case strings.TrimSpace(s.ReferenceNumber) == "":
		return fmt.Errorf("%w: reference number is required", ErrInvalidShipment)
	case strings.TrimSpace(s.Origin) == "":
		return fmt.Errorf("%w: origin is required", ErrInvalidShipment)
	case strings.TrimSpace(s.Destination) == "":
		return fmt.Errorf("%w: destination is required", ErrInvalidShipment)
	case strings.TrimSpace(s.DriverDetails) == "":
		return fmt.Errorf("%w: driver and unit details are required", ErrInvalidShipment)
	case s.Amount <= 0:
		return fmt.Errorf("%w: amount must be greater than zero", ErrInvalidShipment)
	case s.DriverRevenue < 0:
		return fmt.Errorf("%w: driver revenue must not be negative", ErrInvalidShipment)
	case s.DriverRevenue > s.Amount:
		return fmt.Errorf("%w: driver revenue must not exceed shipment amount", ErrInvalidShipment)
	default:
		return nil
	}
}

func (s *Shipment) Transition(newStatus Status) (*Event, error) {
	if s.CurrentStatus == newStatus {
		return nil, ErrDuplicateStatus
	}

	if !canTransition(s.CurrentStatus, newStatus) {
		return nil, ErrInvalidTransition
	}

	s.CurrentStatus = newStatus

	return &Event{
		ID:         uuid.NewString(),
		ShipmentID: s.ID,
		Status:     newStatus,
		CreatedAt:  time.Now(),
	}, nil
}
