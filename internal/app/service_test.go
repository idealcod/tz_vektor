package app

import (
	"context"
	"testing"

	"tz/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type shipmentRepoStub struct {
	createFunc                func(context.Context, *domain.Shipment, *domain.Event) error
	getByIDFunc               func(context.Context, string) (*domain.Shipment, error)
	updateStatusWithEventFunc func(context.Context, string, domain.Status, *domain.Event) error
	getHistoryFunc            func(context.Context, string) ([]*domain.Event, error)
}

func (s *shipmentRepoStub) Create(ctx context.Context, shipment *domain.Shipment, initialEvent *domain.Event) error {
	if s.createFunc != nil {
		return s.createFunc(ctx, shipment, initialEvent)
	}
	return nil
}

func (s *shipmentRepoStub) GetByID(ctx context.Context, id string) (*domain.Shipment, error) {
	if s.getByIDFunc != nil {
		return s.getByIDFunc(ctx, id)
	}
	return nil, domain.ErrShipmentNotFound
}

func (s *shipmentRepoStub) UpdateStatusWithEvent(ctx context.Context, id string, status domain.Status, event *domain.Event) error {
	if s.updateStatusWithEventFunc != nil {
		return s.updateStatusWithEventFunc(ctx, id, status, event)
	}
	return nil
}

func (s *shipmentRepoStub) GetHistory(ctx context.Context, shipmentID string) ([]*domain.Event, error) {
	if s.getHistoryFunc != nil {
		return s.getHistoryFunc(ctx, shipmentID)
	}
	return nil, nil
}

func TestCreateShipment_PersistsShipmentWithInitialEvent(t *testing.T) {
	var savedShipment *domain.Shipment
	var savedEvent *domain.Event

	repo := &shipmentRepoStub{
		createFunc: func(_ context.Context, shipment *domain.Shipment, initialEvent *domain.Event) error {
			savedShipment = shipment
			savedEvent = initialEvent
			return nil
		},
	}

	service := NewShipmentService(repo)

	shipment, err := service.CreateShipment(context.Background(), "REF-001", "Almaty", "Astana", "John / Truck ABC-123", 1000, 500)

	require.NoError(t, err)
	require.NotNil(t, shipment)
	require.NotNil(t, savedShipment)
	require.NotNil(t, savedEvent)
	assert.Equal(t, shipment.ID, savedShipment.ID)
	assert.Equal(t, shipment.ID, savedEvent.ShipmentID)
	assert.Equal(t, domain.StatusPending, savedEvent.Status)
	assert.Equal(t, "shipment created", savedEvent.Message)
}

func TestCreateShipment_RejectsInvalidInput(t *testing.T) {
	repo := &shipmentRepoStub{}
	service := NewShipmentService(repo)

	_, err := service.CreateShipment(context.Background(), "", "Almaty", "Astana", "John / Truck ABC-123", 1000, 500)

	assert.ErrorIs(t, err, domain.ErrInvalidShipment)
}

func TestAddStatusEvent_AppendsEventAndUpdatesStatus(t *testing.T) {
	shipment := newServiceTestShipment(t)
	var persistedEvent *domain.Event

	repo := &shipmentRepoStub{
		getByIDFunc: func(_ context.Context, id string) (*domain.Shipment, error) {
			assert.Equal(t, shipment.ID, id)
			return shipment, nil
		},
		updateStatusWithEventFunc: func(_ context.Context, id string, status domain.Status, event *domain.Event) error {
			assert.Equal(t, shipment.ID, id)
			assert.Equal(t, domain.StatusPickedUp, status)
			persistedEvent = event
			return nil
		},
	}

	service := NewShipmentService(repo)

	event, err := service.AddStatusEvent(context.Background(), shipment.ID, domain.StatusPickedUp, "driver arrived at pickup point")

	require.NoError(t, err)
	require.NotNil(t, event)
	require.NotNil(t, persistedEvent)
	assert.Equal(t, event.ID, persistedEvent.ID)
	assert.Equal(t, domain.StatusPickedUp, event.Status)
	assert.Equal(t, "driver arrived at pickup point", event.Message)
}

func TestAddStatusEvent_ReturnsNotFoundForUnknownShipment(t *testing.T) {
	repo := &shipmentRepoStub{}
	service := NewShipmentService(repo)

	_, err := service.AddStatusEvent(context.Background(), "missing", domain.StatusPickedUp, "driver arrived")

	assert.ErrorIs(t, err, domain.ErrShipmentNotFound)
}

func newServiceTestShipment(t *testing.T) *domain.Shipment {
	t.Helper()

	shipment, err := domain.NewShipment("REF-001", "Almaty", "Astana", "John / Truck ABC-123", 1000, 500)
	require.NoError(t, err)

	return shipment
}
