module github.com/b25/services/notification

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/go-playground/validator/v10 v10.16.0
	github.com/golang-migrate/migrate/v4 v4.17.0
	github.com/google/uuid v1.5.0
	github.com/jmoiron/sqlx v1.3.5
	github.com/lib/pq v1.10.9
	github.com/prometheus/client_golang v1.18.0
	github.com/redis/go-redis/v9 v9.4.0
	github.com/sendgrid/sendgrid-go v3.14.0+incompatible
	github.com/spf13/viper v1.18.2
	github.com/stretchr/testify v1.8.4
	github.com/twilio/twilio-go v1.19.0
	go.uber.org/zap v1.26.0
	google.golang.org/grpc v1.60.1
	google.golang.org/protobuf v1.32.0
)

require (
	firebase.google.com/go/v4 v4.13.0
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/hibiken/asynq v0.24.1
	gopkg.in/yaml.v3 v3.0.1
)
