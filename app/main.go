package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"participantes/cleitinif/config"
	cc "participantes/cleitinif/context"
	"participantes/cleitinif/controllers"
	"participantes/cleitinif/repositories"
	"participantes/cleitinif/services"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()
	/// Observability
	shutdown := config.InitProviderWithJaegerExporter(ctx)
	shutdownMetrics := config.InitMetricsExporter(ctx)
	defer func() {
		shutdown(ctx)
		shutdownMetrics(ctx)
	}()

	tracer := otel.Tracer("app")

	r := gin.New()

	dbConfig, err := config.NewDatabaseConfig()
	if err != nil {
		panic(err)
	}

	loggerConfig := zap.NewProductionConfig()
	loggerConfig.Development = false
	// config.OutputPaths = []string{""}
	// config.ErrorOutputPaths = []string{"stderr", "./errors.log"}
	loggerConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	loggerConfig.Encoding = "json"

	logger, err := loggerConfig.Build()

	if err != nil {
		panic(err)
	}

	defer logger.Sync()
	sugar := logger.Sugar()
	sugar.Infow("Starting the application")

	if os.Getenv("OPEN_TELEMETRY_ENABLED") == "true" {
		r.Use(otelgin.Middleware("app",
			otelgin.WithFilter(func(c *http.Request) bool {
				// Exclude health check
				return c.URL.Path != "/health"
			})),
		)
	}

	r.Use(func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
				sugar.Errorw("Panic recovered", "error", r)
			}
		}()

		customMetrics := config.GetCustomMetrics()
		span := trace.SpanFromContext(c.Request.Context())

		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		customMetrics.TotalRequests.Inc()

		ctx := cc.NewApplicationContext(c.Request.Context())
		reqId := uuid.NewString()
		perRequestLogger := sugar.With(zap.String("request_id", reqId))
		ctx.SetRequestID(reqId)
		ctx.SetLogger(perRequestLogger)

		c.Request = c.Request.WithContext(ctx)

		span.SetAttributes(attribute.String("http.request_id", reqId))

		timeStart := time.Now()
		perRequestLogger.Infow("Starting the request", "path", c.Request.URL.Path, "method", c.Request.Method)
		c.Next()
		status := c.Writer.Status()
		timeElapsed := time.Since(timeStart)
		perRequestLogger.Infow("Finishing the request", "path", c.Request.URL.Path, "method", c.Request.Method, "elapsed", timeElapsed, "status", status)

		if status >= 400 {
			perRequestLogger.Errorw("Request with error", "status", status, "error", c.Errors.String())
			span.SetStatus(codes.Error, fmt.Sprintf("%d", status))
			customMetrics.ErrorRequests.Inc()
		}

	})

	var dbUrl = fmt.Sprintf("postgres://%s:%s@%s:%d/%s", dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Database)
	conf, err := pgxpool.ParseConfig(dbUrl)
	if err != nil {
		sugar.Fatalw("Failed to create a config, error: ", err)
	}

	pgxpool, err := pgxpool.NewWithConfig(ctx, conf)

	if err != nil {
		panic(err)
	}

	customerRepository := repositories.NewCustomerRepository(pgxpool, sugar, tracer)
	customerController := controllers.NewCustomerController(services.NewCustomerService(customerRepository, pgxpool))

	r.GET("/clientes/:id/extrato", customerController.GetStatement)
	r.POST("/clientes/:id/transacoes", customerController.InsertTransaction)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	r.Run("0.0.0.0:8080")
}
