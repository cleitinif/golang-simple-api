package repositories

import (
	"context"
	"participantes/cleitinif/config"
	cc "participantes/cleitinif/context"
	"participantes/cleitinif/dto"
	"participantes/cleitinif/models"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type CustomerRepository struct {
	manager *pgxpool.Pool
	sugar   *zap.SugaredLogger
	tracer  trace.Tracer
}

type DbManagerKey struct {
	Conn *pgx.Conn
}

func NewCustomerRepository(manager *pgxpool.Pool, sugar *zap.SugaredLogger, tracer trace.Tracer) *CustomerRepository {
	return &CustomerRepository{
		manager: manager,
		sugar:   sugar,
		tracer:  tracer,
	}
}

func (cr *CustomerRepository) InsertTransaction(ctx context.Context, transaction *models.Transaction, id int) (*dto.InsertTransactionResponse, error) {
	ctx, span := cr.tracer.Start(ctx, "CustomerRepository.InsertTransaction")
	defer span.End()

	manager := ctx.Value(DbManagerKey{}).(*DbManagerKey).Conn

	logger := ctx.Value(cc.LoggerKey{}).(*zap.SugaredLogger)

	_, err := manager.Exec(ctx, "INSERT INTO transacoes (cliente_id, valor, tipo, descricao) VALUES ($1, $2, $3, $4)", id, transaction.Value, transaction.Type, transaction.Description)

	if err != nil {
		logger.Warnw("Failed to insert transaction", "transaction", transaction, "customerId", id, "error", err)
		return nil, config.HandleDatabaseError(err)
	}

	return nil, nil
}

func (cr *CustomerRepository) GetTransactions(ctx context.Context, id int) ([]models.Transaction, error) {
	ctx, span := cr.tracer.Start(ctx, "CustomerRepository.GetTransactions")
	defer span.End()

	logger := ctx.Value(cc.LoggerKey{}).(*zap.SugaredLogger)
	logger.Debugw("Getting transactions", "customerId", id)

	manager := ctx.Value(DbManagerKey{}).(*DbManagerKey).Conn

	rows, err := manager.Query(ctx, "SELECT valor, tipo, descricao, realizada_em FROM transacoes WHERE cliente_id = $1 ORDER BY id DESC LIMIT 10", id)
	if err != nil {
		logger.Warnw("Failed to get transactions", "customerId", id, "error", err)
		return nil, config.HandleDatabaseError(err)
	}

	transactions := make([]models.Transaction, 0)
	for rows.Next() {
		var transaction models.Transaction
		var date time.Time

		err = rows.Scan(&transaction.Value, &transaction.Type, &transaction.Description, &date)
		transaction.Date = date.Format(time.RFC3339)

		if err != nil {
			logger.Warnw("Failed to scan transaction", "error", err)
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

func (cr *CustomerRepository) GetCustomer(ctx context.Context, id int) (*models.Customer, error) {
	ctx, span := cr.tracer.Start(ctx, "CustomerRepository.GetCustomer")
	defer span.End()

	logger := ctx.Value(cc.LoggerKey{}).(*zap.SugaredLogger)
	logger.Debugw("Getting customer", "customerId", id)

	manager := ctx.Value(DbManagerKey{}).(*DbManagerKey).Conn

	row := manager.QueryRow(ctx, "SELECT id, limite, saldo FROM clientes WHERE id = $1", id)

	var customer models.Customer
	err := row.Scan(&customer.ID, &customer.Limit, &customer.Amout)
	if err != nil {
		logger.Warnw("Failed to get customer", "customerId", id, "error", err)
		return nil, config.HandleDatabaseError(err)
	}

	return &customer, nil
}

func (cr *CustomerRepository) UpdateCustomerBalance(ctx context.Context, customer *models.Customer, value int) (*models.Customer, error) {
	ctx, span := cr.tracer.Start(ctx, "CustomerRepository.UpdateCustomerBalance")
	defer span.End()

	logger := ctx.Value(cc.LoggerKey{}).(*zap.SugaredLogger)
	logger.Debugw("Updating customer", "customerId", customer.ID)

	manager := ctx.Value(DbManagerKey{}).(*DbManagerKey).Conn

	err := manager.QueryRow(ctx, "UPDATE clientes SET saldo = saldo + $1 WHERE id = $2 returning id, limite, saldo", value, customer.ID).Scan(&customer.ID, &customer.Limit, &customer.Amout)
	if err != nil {
		logger.Warnw("Failed to update customer balance", "customerId", customer.ID, "error", err)
		return nil, config.HandleDatabaseError(err)
	}

	return customer, nil
}
