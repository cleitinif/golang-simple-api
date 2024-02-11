package controllers

import (
	"context"
	er "errors"
	"participantes/cleitinif/errors"
	"participantes/cleitinif/models"
	"participantes/cleitinif/services"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type CustomerController struct {
	customerService *services.CustomerService
}

func NewCustomerController(customerService *services.CustomerService) *CustomerController {
	return &CustomerController{
		customerService: customerService,
	}
}

func (cc *CustomerController) GetStatement(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.Status(400)
		return
	}

	parsedID, err := strconv.Atoi(id)
	if err != nil {
		c.Status(400)
		return
	}

	ctx := c.Request.Context()

	statement, err := cc.customerService.GetStatement(ctx, parsedID)
	if err != nil {
		if er.As(err, &errors.NotFoundError) {
			c.Status(404)
			return
		}

		c.Status(500)
		return
	}

	c.JSON(200, statement)
}

func (cc *CustomerController) InsertTransaction(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.Status(400)
		return
	}

	parsedID, err := strconv.Atoi(id)
	if err != nil {
		c.Status(400)
		return
	}

	var transaction models.Transaction
	if err := c.ShouldBindJSON(&transaction); err != nil {
		c.Status(422)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), time.Millisecond*5000)
	defer cancel()

	r, err := cc.customerService.InsertTransaction(ctx, transaction, parsedID)
	if err != nil {
		if er.As(err, &errors.NotEnoughBalanceError) {
			c.Status(422)
			return
		}

		if er.As(err, &errors.TransactionConflictError) {
			c.Status(503)
			return
		}

		c.Status(500)
		return
	}

	c.JSON(200, r)
}
