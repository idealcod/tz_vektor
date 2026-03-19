package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "tz/gen/shipment/v1"
	"tz/internal/app"
	"tz/internal/domain"
)

type Handler struct {
	pb.UnimplementedShipmentServiceServer
	service *app.ShipmentService
}

func NewHandler(service *app.ShipmentService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateShipment(ctx context.Context, req *pb.CreateShipmentRequest) (*pb.Shipment, error) {
	s, err := h.service.CreateShipment(ctx,
		req.ReferenceNumber, req.Origin, req.Destination,
		req.DriverDetails, req.Amount, req.DriverRevenue,
	)
	if err != nil {
		return nil, toStatusError("failed to create shipment", err)
	}

	return shipmentToProto(s), nil
}

func (h *Handler) GetShipment(ctx context.Context, req *pb.GetShipmentRequest) (*pb.Shipment, error) {
	s, err := h.service.GetShipment(ctx, req.Id)
	if err != nil {
		return nil, toStatusError("failed to get shipment", err)
	}

	return shipmentToProto(s), nil
}

func (h *Handler) AddStatusEvent(ctx context.Context, req *pb.AddStatusEventRequest) (*pb.ShipmentEvent, error) {
	if req.NewStatus == pb.Status_STATUS_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "new_status must be specified")
	}

	newStatus := protoStatusToDomain(req.NewStatus)

	event, err := h.service.AddStatusEvent(ctx, req.ShipmentId, newStatus, req.Message)
	if err != nil {
		return nil, toStatusError("failed to add event", err)
	}

	return &pb.ShipmentEvent{
		Id:         event.ID,
		ShipmentId: event.ShipmentID,
		Status:     domainStatusToProto(event.Status),
		Message:    event.Message,
		OccurredAt: timestamppb.New(event.CreatedAt),
	}, nil
}

func (h *Handler) GetShipmentHistory(ctx context.Context, req *pb.GetShipmentHistoryRequest) (*pb.GetShipmentHistoryResponse, error) {
	if _, err := h.service.GetShipment(ctx, req.ShipmentId); err != nil {
		return nil, toStatusError("failed to get shipment history", err)
	}

	events, err := h.service.GetHistory(ctx, req.ShipmentId)
	if err != nil {
		return nil, toStatusError("failed to get shipment history", err)
	}

	resp := &pb.GetShipmentHistoryResponse{}
	for _, e := range events {
		resp.Events = append(resp.Events, &pb.ShipmentEvent{
			Id:         e.ID,
			ShipmentId: e.ShipmentID,
			Status:     domainStatusToProto(e.Status),
			Message:    e.Message,
			OccurredAt: timestamppb.New(e.CreatedAt),
		})
	}
	return resp, nil
}

func shipmentToProto(s *domain.Shipment) *pb.Shipment {
	return &pb.Shipment{
		Id:              s.ID,
		ReferenceNumber: s.ReferenceNumber,
		Origin:          s.Origin,
		Destination:     s.Destination,
		CurrentStatus:   domainStatusToProto(s.CurrentStatus),
		DriverDetails:   s.DriverDetails,
		Amount:          s.Amount,
		DriverRevenue:   s.DriverRevenue,
		CreatedAt:       timestamppb.New(s.CreatedAt),
	}
}

func protoStatusToDomain(s pb.Status) domain.Status {
	switch s {
	case pb.Status_STATUS_PENDING:
		return domain.StatusPending
	case pb.Status_STATUS_PICKED_UP:
		return domain.StatusPickedUp
	case pb.Status_STATUS_IN_TRANSIT:
		return domain.StatusInTransit
	case pb.Status_STATUS_DELIVERED:
		return domain.StatusDelivered
	default:
		return domain.StatusPending
	}
}

func domainStatusToProto(s domain.Status) pb.Status {
	switch s {
	case domain.StatusPending:
		return pb.Status_STATUS_PENDING
	case domain.StatusPickedUp:
		return pb.Status_STATUS_PICKED_UP
	case domain.StatusInTransit:
		return pb.Status_STATUS_IN_TRANSIT
	case domain.StatusDelivered:
		return pb.Status_STATUS_DELIVERED
	default:
		return pb.Status_STATUS_UNSPECIFIED
	}
}

func toStatusError(prefix string, err error) error {
	switch {
	case errors.Is(err, domain.ErrShipmentNotFound):
		return status.Errorf(codes.NotFound, "%s: %v", prefix, err)
	case errors.Is(err, domain.ErrInvalidShipment),
		errors.Is(err, domain.ErrInvalidTransition),
		errors.Is(err, domain.ErrDuplicateStatus):
		return status.Errorf(codes.InvalidArgument, "%s: %v", prefix, err)
	default:
		return status.Errorf(codes.Internal, "%s: %v", prefix, err)
	}
}
