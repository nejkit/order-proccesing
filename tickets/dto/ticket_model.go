package dto

import "order-processing/external/tickets"

type TicketModel struct {
	TicketId      string
	TicketState   int
	OperationType int
	Data          string
}

func MapToModel(ticket *tickets.Ticket) TicketModel {
	return TicketModel{
		TicketId:      ticket.TicketId,
		TicketState:   int(ticket.State),
		OperationType: int(ticket.OperationType),
		Data:          string(ticket.Data),
	}
}

func MapToProto(ticket TicketModel) *tickets.Ticket {
	return &tickets.Ticket{
		TicketId:      ticket.TicketId,
		State:         tickets.TicketState(ticket.TicketState),
		OperationType: tickets.OperationType(ticket.OperationType),
		Data:          []byte(ticket.Data),
	}
}
