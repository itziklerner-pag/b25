module github.com/yourorg/b25/services/payment

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/google/uuid v1.5.0
	github.com/lib/pq v1.10.9
	github.com/prometheus/client_golang v1.18.0
	github.com/stripe/stripe-go/v76 v76.12.0
	github.com/joho/godotenv v1.5.1
	go.uber.org/zap v1.26.0
	github.com/golang-migrate/migrate/v4 v4.17.0
	github.com/go-playground/validator/v10 v10.16.0
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/redis/go-redis/v9 v9.3.1
)
