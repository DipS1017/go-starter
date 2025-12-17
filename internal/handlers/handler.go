package handlers

import (
	"log/slog"

	"github.com/webpoint-solutions-llc/dba/internal/interfaces"
	"github.com/webpoint-solutions-llc/dba/internal/responders"
	"github.com/webpoint-solutions-llc/dba/internal/services"
)

type Handler struct {
	svc *services.Service
	log *slog.Logger
	res interfaces.Responders
}

func NewHandler() *Handler {
	logger := slog.Default().With("component", "handler")
	return &Handler{
		log: logger,
		svc: services.NewService(),
		res: responders.NewResponder(logger),
	}
}
