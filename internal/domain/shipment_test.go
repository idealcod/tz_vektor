package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestShipment() *Shipment {
	s, err := NewShipment("REF-001", "Almaty", "Astana", "John / Truck ABC-123", 1000, 500)
	if err != nil {
		panic(err)
	}
	return s
}

func TestNewShipment_StartsWithPending(t *testing.T) {
	s := newTestShipment()

	assert.Equal(t, StatusPending, s.CurrentStatus)
	assert.NotEmpty(t, s.ID)
}

func TestNewShipment_RequiresValidFields(t *testing.T) {
	_, err := NewShipment("", "Almaty", "Astana", "John / Truck ABC-123", 1000, 500)

	assert.ErrorIs(t, err, ErrInvalidShipment)
}

func TestNewShipment_RejectsRevenueGreaterThanAmount(t *testing.T) {
	_, err := NewShipment("REF-001", "Almaty", "Astana", "John / Truck ABC-123", 1000, 1500)

	assert.ErrorIs(t, err, ErrInvalidShipment)
}

func TestTransition_PendingToPickedUp_Valid(t *testing.T) {
	s := newTestShipment()

	event, err := s.Transition(StatusPickedUp)

	require.NoError(t, err)
	assert.Equal(t, StatusPickedUp, s.CurrentStatus)
	assert.Equal(t, StatusPickedUp, event.Status)
}

func TestTransition_PendingToDelivered_Invalid(t *testing.T) {
	s := newTestShipment()

	_, err := s.Transition(StatusDelivered)

	assert.ErrorIs(t, err, ErrInvalidTransition)
	assert.Equal(t, StatusPending, s.CurrentStatus)
}

func TestTransition_PendingToInTransit_Invalid(t *testing.T) {
	s := newTestShipment()

	_, err := s.Transition(StatusInTransit)

	assert.ErrorIs(t, err, ErrInvalidTransition)
}

func TestTransition_DuplicateStatus_Invalid(t *testing.T) {
	s := newTestShipment()

	_, err := s.Transition(StatusPending)

	assert.ErrorIs(t, err, ErrDuplicateStatus)
}

func TestTransition_FullHappyPath(t *testing.T) {
	s := newTestShipment()

	_, err := s.Transition(StatusPickedUp)
	require.NoError(t, err)

	_, err = s.Transition(StatusInTransit)
	require.NoError(t, err)

	_, err = s.Transition(StatusDelivered)
	require.NoError(t, err)

	assert.Equal(t, StatusDelivered, s.CurrentStatus)
}

func TestTransition_FromDelivered_AlwaysInvalid(t *testing.T) {
	s := newTestShipment()
	s.Transition(StatusPickedUp)
	s.Transition(StatusInTransit)
	s.Transition(StatusDelivered)

	_, err := s.Transition(StatusPickedUp)

	assert.ErrorIs(t, err, ErrInvalidTransition)
}

func TestTransition_DeliveredToPickedUp_Invalid(t *testing.T) {
	s := newTestShipment()
	s.CurrentStatus = StatusDelivered

	_, err := s.Transition(StatusPickedUp)

	assert.ErrorIs(t, err, ErrInvalidTransition)
}
