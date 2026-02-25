package services

import (
	"context"
	"net/url"

	"github.com/stripe/stripe-go/v82"
	billingsession "github.com/stripe/stripe-go/v82/billingportal/session"
	"github.com/stripe/stripe-go/v82/checkout/session"
	"github.com/stripe/stripe-go/v82/customer"
	"github.com/webpoint-solutions-llc/go-starter/internal/config"
	"github.com/webpoint-solutions-llc/go-starter/internal/dto"
)

func (s *Service) CreatePortalUpgradeSession(customerID, subscriptionID, newPriceID, returnURL string) (*stripe.BillingPortalSession, error) {
	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(customerID),
		ReturnURL: stripe.String(returnURL),
		FlowData: &stripe.BillingPortalSessionFlowDataParams{
			Type: stripe.String("subscription_update_confirm"),
			SubscriptionUpdateConfirm: &stripe.BillingPortalSessionFlowDataSubscriptionUpdateConfirmParams{
				Subscription: stripe.String(subscriptionID),
				Items: []*stripe.BillingPortalSessionFlowDataSubscriptionUpdateConfirmItemParams{
					{
						Price:    stripe.String(newPriceID),
						Quantity: stripe.Int64(1),
					},
				},
			},
		},
	}
	return billingsession.New(params)
}

func (s *Service) CreateCheckoutSession(params dto.CheckoutSessionParams) (*stripe.CheckoutSession, error) {
	stripe.Key = config.Cfg.StripeAPIKey

	callbackURL, err := url.JoinPath(config.Cfg.FrontendURL, config.Cfg.StripeCheckoutCallback)
	// /profile/portfolio?payment_success=<id>
	if err != nil {
		return nil, err
	}
	// Parse the resulting URL
	parsedURL, err := url.Parse(callbackURL)
	if err != nil {
		return nil, err
	}

	successURLObj := *parsedURL
	qSuccess := successURLObj.Query()
	qSuccess.Set("success", "true")
	qSuccess.Set("id", params.PortfolioID)
	successURLObj.RawQuery = qSuccess.Encode()
	successURL := successURLObj.String()

	cancelURLObj := *parsedURL
	qCancel := cancelURLObj.Query()
	qCancel.Set("success", "false")
	qCancel.Set("id", params.PortfolioID)
	cancelURLObj.RawQuery = qCancel.Encode()
	cancelURL := cancelURLObj.String()

	session_params := &stripe.CheckoutSessionParams{
		Customer:           stripe.String(params.CustomerID),
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		Currency:           stripe.String("usd"),
		Mode:               stripe.String(string(stripe.CheckoutSessionModePayment)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(params.PriceID),
				Quantity: stripe.Int64(params.Quantity),
			},
		},
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		Metadata:   params.Metadata,
	}

	return session.New(session_params)
}

func (s *Service) CreateOrGetStripeCustomer(email, name string, metadata map[string]string) (string, error) {
	stripe.Key = config.Cfg.StripeAPIKey

	stripeInfo, _ := s.q.GetUserStripeInfoByEmail(context.Background(), email)

	if stripeInfo.StripeCustomerID != nil && *stripeInfo.StripeCustomerID != "" {
		return *stripeInfo.StripeCustomerID, nil
	}

	// Search for customer by email
	params := &stripe.CustomerListParams{}
	params.Filters.AddFilter("email", "", email)
	i := customer.List(params)

	for i.Next() {
		c := i.Customer()
		if c.Email == email {
			return c.ID, nil
		}
	}

	newParams := &stripe.CustomerParams{
		Email: stripe.String(email),
		Name:  stripe.String(name),

		Metadata: metadata,
	}
	c, err := customer.New(newParams)
	if err != nil {
		return "", err
	}
	return c.ID, nil
}
