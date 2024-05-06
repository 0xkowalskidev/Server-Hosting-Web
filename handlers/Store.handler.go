package handlers

import (
	"0xKowalski1/server-hosting-web/config"
	"0xKowalski1/server-hosting-web/models"
	"0xKowalski1/server-hosting-web/services"
	"0xKowalski1/server-hosting-web/templates"

	"log"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/checkout/session"
)

type StoreHandler struct {
	CurrencyService *services.CurrencyService
	PriceService    *services.PriceService
}

func NewStoreHandler(currencyService *services.CurrencyService, priceService *services.PriceService) *StoreHandler {
	return &StoreHandler{
		CurrencyService: currencyService,
		PriceService:    priceService,
	}
}

func (sh *StoreHandler) GetStore(c echo.Context) error {
	priceMap, err := sh.getPrices(c)
	if err != nil {
		//500
	}

	return Render(c, 200, templates.StorePage(priceMap["memory"], priceMap["storage"], priceMap["archive"]))
}

func (sh *StoreHandler) SubmitStoreForm(c echo.Context) error {
	memory, memErr := strconv.Atoi(c.FormValue("memory"))
	storage, stoErr := strconv.Atoi(c.FormValue("storage"))
	archive, arcErr := strconv.Atoi(c.FormValue("archive"))

	prices, err := sh.getPrices(c)
	if err != nil {
		//500
	}

	// Validate
	if memErr != nil || stoErr != nil || arcErr != nil {
		// 400
	}

	// Init Stripe
	log.Println(config.Envs.StripeSecretKey)
	stripe.Key = config.Envs.StripeSecretKey

	domain := "http://localhost:3000"
	params := &stripe.CheckoutSessionParams{
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(prices["memory"].StripeID),
				Quantity: stripe.Int64(int64(memory)),
			},
			{
				Price:    stripe.String(prices["storage"].StripeID),
				Quantity: stripe.Int64(int64(storage)),
			},
			{
				Price:    stripe.String(prices["archive"].StripeID),
				Quantity: stripe.Int64(int64(archive)),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL: stripe.String(domain + "/profile/gameservers"),
		CancelURL:  stripe.String(domain + "/store"),
	}

	stripeCheckoutSession, err := session.New(params)

	if err != nil {
		log.Printf("session.New: %v", err)
		// render error
	}

	c.Response().Header().Set("HX-Redirect", stripeCheckoutSession.URL)
	return c.NoContent(302)
}

func (sh *StoreHandler) GetGuidedStoreFlow(c echo.Context) error {
	return Render(c, 200, templates.GuidedStoreFlow())
}

func (sh *StoreHandler) GetAdvancedStoreFlow(c echo.Context) error {
	priceMap, err := sh.getPrices(c)
	if err != nil {
		//500
	}

	return Render(c, 200, templates.AdvancedStoreFlow(priceMap["memory"], priceMap["storage"], priceMap["archive"]))
}

func (sh *StoreHandler) getPrices(c echo.Context) (map[string]models.Price, error) {
	var user *models.User
	userInterface := c.Get("user")
	if userInterface != nil {
		userConversion, ok := userInterface.(*models.User)
		if ok {
			user = userConversion
		}
	}

	var currency models.Currency
	var err error // Avoid variable shadowing
	if user != nil {
		currency, err = sh.CurrencyService.GetCurrencyById(user.CurrencyID)

		if err != nil {
			return nil, err
		}
	} else {
		// Default to USD
		currency, err = sh.CurrencyService.GetDefaultCurrency()
	}

	// Get prices for currency
	prices, err := sh.PriceService.GetPrices(currency)
	if err != nil {
		return nil, err
	}
	priceMap := make(map[string]models.Price)
	for _, price := range prices {
		priceMap[price.Type] = price
	}

	return priceMap, nil
}
