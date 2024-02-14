package main

import (
	"context"
	"fmt"
	"participantes/cleitinif/config"
	"participantes/cleitinif/controllers"
	"participantes/cleitinif/repositories"
	"participantes/cleitinif/services"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()

	r := gin.New()

	dbConfig, err := config.NewDatabaseConfig()
	if err != nil {
		panic(err)
	}

	if err != nil {
		panic(err)
	}

	r.Use(gin.Recovery())

	var dbUrl = fmt.Sprintf("postgres://%s:%s@%s:%d/%s", dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Database)
	conf, err := pgxpool.ParseConfig(dbUrl)
	if err != nil {
		panic(err)
	}

	pgxpool, err := pgxpool.NewWithConfig(ctx, conf)
	if err != nil {
		panic(err)
	}

	customerRepository := repositories.NewCustomerRepository(pgxpool)
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
