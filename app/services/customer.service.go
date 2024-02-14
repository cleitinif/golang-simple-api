package services

import (
	"context"
	"math"
	"participantes/cleitinif/config"
	"participantes/cleitinif/dto"
	"participantes/cleitinif/errors"
	"participantes/cleitinif/models"
	"participantes/cleitinif/repositories"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CustomerService struct {
	customerRepository *repositories.CustomerRepository
	pool               *pgxpool.Pool
}

func NewCustomerService(customerRepository *repositories.CustomerRepository, pool *pgxpool.Pool) *CustomerService {
	return &CustomerService{
		customerRepository: customerRepository,
		pool:               pool,
	}
}

func (cs *CustomerService) GetStatement(ctx context.Context, id int) (*models.Statement, error) {
	manager, err := cs.pool.Acquire(ctx)
	if err != nil {
		return nil, errors.NewInternalError()
	}
	manager.Exec(ctx, "SET TIME ZONE 'America/Sao_Paulo'; SET idle_in_transaction_session_timeout=5000; SET statement_timeout=5000; SET lock_timeout=3000;")

	tx, err := manager.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.Serializable,
		AccessMode: pgx.ReadOnly,
	})

	if err != nil {
		return nil, errors.NewInternalError()
	}
	ctx = context.WithValue(ctx, repositories.DbManagerKey{}, &repositories.DbManagerKey{Conn: tx.Conn()})

	defer func() {
		tx.Rollback(ctx)
		manager.Release()
	}()

	customer, err := cs.customerRepository.GetCustomer(ctx, id)
	if err != nil {
		return nil, err
	}

	transactions, err := cs.customerRepository.GetTransactions(ctx, id)
	if err != nil {
		return nil, err
	}

	today := time.Now().Format(time.RFC3339)

	statement := &models.Statement{
		Balance: models.Balance{
			Amount: customer.Amout,
			Limit:  customer.Limit,
			Date:   today,
		},
		LastTransactions: transactions,
	}

	return statement, nil
}

func (cs *CustomerService) InsertTransaction(ctx context.Context, transaction models.Transaction, id int) (*dto.InsertTransactionResponse, error) {
	manager, err := cs.pool.Acquire(ctx)
	if err != nil {
		return nil, errors.NewInternalError()
	}
	// manager.Exec(ctx, "SET TIME ZONE 'America/Sao_Paulo'; SET idle_in_transaction_session_timeout=5000; SET statement_timeout=5000; SET lock_timeout=3000;")

	tx, err := manager.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	})
	ctx = context.WithValue(ctx, repositories.DbManagerKey{}, &repositories.DbManagerKey{Conn: tx.Conn()})
	if err != nil {
		return nil, errors.NewInternalError()
	}

	defer func() {
		tx.Rollback(ctx)
		manager.Release()
	}()

	customer, err := cs.customerRepository.GetCustomer(ctx, id)
	if err != nil {
		return nil, err
	}

	_, err = cs.customerRepository.InsertTransaction(ctx, &transaction, id)
	if err != nil {
		return nil, err
	}

	var balanceToUpdate int
	if transaction.Type == "c" {
		balanceToUpdate = transaction.Value
	} else {
		balanceToUpdate = -transaction.Value
	}
	_, err = cs.customerRepository.UpdateCustomerBalance(ctx, customer, balanceToUpdate)
	if err != nil {
		return nil, err
	}

	if !hasEnoughBalance(customer, &transaction) {
		return nil, errors.NewNotEnoughBalanceError()
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, config.HandleDatabaseError(err)
	}

	return &dto.InsertTransactionResponse{
		Limit:  customer.Limit,
		Amount: customer.Amout,
	}, nil
}

func hasEnoughBalance(customer *models.Customer, transaction *models.Transaction) bool {
	if transaction.Type == "c" {
		return true
	}

	calculatedBalance := customer.Amout - transaction.Value

	if calculatedBalance >= 0 && customer.Limit > 0 {
		return true
	}

	if calculatedBalance < 0 && math.Abs(float64(calculatedBalance)) < float64(customer.Limit) {
		return true
	}

	return false
}

// 1000 - 500 = 500
