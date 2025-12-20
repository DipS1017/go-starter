package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
	"github.com/webpoint-solutions-llc/go-starter/internal/config"
	"github.com/webpoint-solutions-llc/go-starter/internal/dto"
)

func (h *Handler) StripeWebhook(c echo.Context) error {
	const MaxBodyBytes = int64(65536)
	c.Request().Body = http.MaxBytesReader(c.Response(), c.Request().Body, MaxBodyBytes)

	payload, err := io.ReadAll(c.Request().Body)
	if err != nil {
		h.log.Error("Error reading request body", "err", err)
		return c.String(http.StatusRequestEntityTooLarge, "Too large")
	}

	sigHeader := c.Request().Header.Get("Stripe-Signature")
	endpointSecret := config.Cfg.StripeWebhookSecret

	event, err := webhook.ConstructEvent(payload, sigHeader, endpointSecret)
	if err != nil {
		h.log.Error("stripe signature verification failed", "err", err)
		return c.String(http.StatusBadRequest, "Invalid signature")
	}

	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			h.log.Error("Error parsing stripe session", "err", err)
			return c.String(http.StatusBadRequest, "Failed to parse session")
		}
		// You can handle the completed checkout session here if needed

		return c.NoContent(http.StatusOK)

	case "customer.created":
		var customer stripe.Customer
		if err := json.Unmarshal(event.Data.Raw, &customer); err != nil {
			h.log.Error("Error parsing stipe session:", " %v", err)
			return c.String(http.StatusBadRequest, "Failed to parse session")
		}

		userID, err := uuid.Parse(customer.Metadata["id"])

		if err != nil {
			h.log.Debug("Invalid user", "Err", err)
		} else {

			user := dto.BaiscUserInfo{
				StripeCustomerID: &customer.ID,
			}

			params := dto.UpdateUserInfoParams{
				User: user,
			}

			updateErr := h.svc.UpdateUser(context.Background(), userID, params)
			if updateErr != nil {
				h.log.Debug("Error on updating user: ", "err", updateErr.Error())
			} else {
				h.log.Debug("Update user")
			}
		}

	default:
		h.log.Debug("Unhandled event type", " ", event.Type)
	}

	return c.NoContent(http.StatusOK)
}
