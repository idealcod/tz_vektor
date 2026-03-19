package domain

import "errors"

type Status string

const (
	StatusPending   Status = "PENDING"
	StatusPickedUp  Status = "PICKED_UP"
	StatusInTransit Status = "IN_TRANSIT"
	StatusDelivered Status = "DELIVERED"
)

var ErrInvalidTransition = errors.New("invalid status transition")
var ErrDuplicateStatus = errors.New("duplicate status update")
var ErrShipmentNotFound = errors.New("shipment not found")
var ErrInvalidShipment = errors.New("invalid shipment")

var allowedTransitions = map[Status][]Status{
	StatusPending:   {StatusPickedUp},
	StatusPickedUp:  {StatusInTransit},
	StatusInTransit: {StatusDelivered},
	StatusDelivered: {},
}

func canTransition(from Status, to Status) bool {
	if from == to {
		return false
	}

	allowed := allowedTransitions[from]
	for _, a := range allowed {
		if a == to {
			return true
		}
	}
	return false
}
