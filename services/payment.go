package services

import (
	"os"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/checkout/session"
	"github.com/stripe/stripe-go/v74/paymentintent"
)

type PaymentService struct{}

func NewPaymentService() *PaymentService {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	return &PaymentService{}
}

type CreateCheckoutSessionRequest struct {
	Amount      int64  `json:"amount"`
	Currency    string `json:"currency"`
	SuccessURL  string `json:"success_url"`
	CancelURL   string `json:"cancel_url"`
	OrderID     string `json:"order_id"`
	Description string `json:"description"`
}

type CreateCheckoutSessionResponse struct {
	SessionID string `json:"session_id"`
	URL       string `json:"url"`
}

func (s *PaymentService) CreateCheckoutSession(req CreateCheckoutSessionRequest) (*CreateCheckoutSessionResponse, error) {
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(req.Currency),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(req.Description),
					},
					UnitAmount: stripe.Int64(req.Amount),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode:              stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL:        stripe.String(req.SuccessURL),
		CancelURL:         stripe.String(req.CancelURL),
		ClientReferenceID: stripe.String(req.OrderID),
	}

	sess, err := session.New(params)
	if err != nil {
		return nil, err
	}

	return &CreateCheckoutSessionResponse{
		SessionID: sess.ID,
		URL:       sess.URL,
	}, nil
}

func (s *PaymentService) ConfirmPayment(paymentIntentID string) error {
	_, err := paymentintent.Confirm(paymentIntentID, nil)
	return err
}

func (s *PaymentService) GetPaymentIntent(paymentIntentID string) (*stripe.PaymentIntent, error) {
	return paymentintent.Get(paymentIntentID, nil)
}

func (s *PaymentService) GetCheckoutSession(sessionID string) (*stripe.CheckoutSession, error) {
	return session.Get(sessionID, nil)
}
