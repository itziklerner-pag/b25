package api

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yourorg/b25/services/payment/internal/api/handlers"
	"github.com/yourorg/b25/services/payment/internal/api/middleware"
	"github.com/yourorg/b25/services/payment/internal/config"
	"github.com/yourorg/b25/services/payment/internal/logger"
	"github.com/yourorg/b25/services/payment/internal/service"
)

func RegisterRoutes(
	router *gin.Engine,
	paymentService *service.PaymentService,
	subscriptionService *service.SubscriptionService,
	invoiceService *service.InvoiceService,
	refundService *service.RefundService,
	webhookService *service.WebhookService,
	cfg *config.Config,
	log *logger.Logger,
) {
	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// Metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Apply CORS middleware
		v1.Use(middleware.CORS(cfg.Security.AllowedOrigins))

		// Apply rate limiting
		v1.Use(middleware.RateLimit(cfg.Security.RateLimitPerMinute))

		// Webhook endpoint (no auth required)
		webhookHandler := handlers.NewWebhookHandler(webhookService, log)
		v1.POST("/webhooks/stripe", webhookHandler.HandleStripeWebhook)

		// Protected routes (require authentication)
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(cfg.Security.JWTSecret))
		{
			// Payment endpoints
			paymentHandler := handlers.NewPaymentHandler(paymentService, log)
			protected.POST("/payments", paymentHandler.CreatePayment)
			protected.GET("/payments/:id", paymentHandler.GetPayment)
			protected.GET("/users/:user_id/payments", paymentHandler.GetUserPayments)

			// Payment method endpoints
			protected.POST("/payment-methods", paymentHandler.AttachPaymentMethod)
			protected.GET("/users/:user_id/payment-methods", paymentHandler.GetUserPaymentMethods)
			protected.DELETE("/payment-methods/:id", paymentHandler.DetachPaymentMethod)

			// Subscription endpoints
			subscriptionHandler := handlers.NewSubscriptionHandler(subscriptionService, log)
			protected.POST("/subscriptions", subscriptionHandler.CreateSubscription)
			protected.GET("/subscriptions/:id", subscriptionHandler.GetSubscription)
			protected.GET("/users/:user_id/subscriptions", subscriptionHandler.GetUserSubscriptions)
			protected.DELETE("/subscriptions/:id", subscriptionHandler.CancelSubscription)

			// Invoice endpoints
			invoiceHandler := handlers.NewInvoiceHandler(invoiceService, log)
			protected.GET("/invoices/:id", invoiceHandler.GetInvoice)
			protected.GET("/users/:user_id/invoices", invoiceHandler.GetUserInvoices)
			protected.GET("/subscriptions/:subscription_id/invoices", invoiceHandler.GetSubscriptionInvoices)

			// Refund endpoints
			refundHandler := handlers.NewRefundHandler(refundService, log)
			protected.POST("/refunds", refundHandler.CreateRefund)
		}
	}
}
