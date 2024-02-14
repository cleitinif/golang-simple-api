package repositories

import (
	"context"
	"participantes/cleitinif/config"
	"participantes/cleitinif/dto"
	"participantes/cleitinif/models"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CustomerRepository struct {
	manager *pgxpool.Pool
}

type DbManagerKey struct {
	Conn *pgx.Conn
}

func NewCustomerRepository(manager *pgxpool.Pool) *CustomerRepository {
	return &CustomerRepository{
		manager: manager,
	}
}

func (cr *CustomerRepository) InsertTransaction(ctx context.Context, transaction *models.Transaction, id int) (*dto.InsertTransactionResponse, error) {
	manager := ctx.Value(DbManagerKey{}).(*DbManagerKey).Conn

	_, err := manager.Exec(ctx, "INSERT INTO transacoes (cliente_id, valor, tipo, descricao) VALUES ($1, $2, $3, $4)", id, transaction.Value, transaction.Type, transaction.Description)

	if err != nil {
		return nil, config.HandleDatabaseError(err)
	}

	return nil, nil
}

func (cr *CustomerRepository) GetTransactions(ctx context.Context, id int) ([]models.Transaction, error) {
	manager := ctx.Value(DbManagerKey{}).(*DbManagerKey).Conn

	rows, err := manager.Query(ctx, "SELECT valor, tipo, descricao, realizada_em FROM transacoes WHERE cliente_id = $1 ORDER BY id DESC LIMIT 10", id)
	if err != nil {
		return nil, config.HandleDatabaseError(err)
	}

	transactions := make([]models.Transaction, 0)
	for rows.Next() {
		var transaction models.Transaction
		var date time.Time

		err = rows.Scan(&transaction.Value, &transaction.Type, &transaction.Description, &date)
		transaction.Date = date.Format(time.RFC3339)

		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

func (cr *CustomerRepository) GetCustomer(ctx context.Context, id int) (*models.Customer, error) {
	manager := ctx.Value(DbManagerKey{}).(*DbManagerKey).Conn

	row := manager.QueryRow(ctx, "SELECT id, limite, saldo FROM clientes WHERE id = $1", id)

	var customer models.Customer
	err := row.Scan(&customer.ID, &customer.Limit, &customer.Amout)
	if err != nil {
		return nil, config.HandleDatabaseError(err)
	}

	return &customer, nil
}

func (cr *CustomerRepository) UpdateCustomerBalance(ctx context.Context, customer *models.Customer, value int) (*models.Customer, error) {
	manager := ctx.Value(DbManagerKey{}).(*DbManagerKey).Conn

	err := manager.QueryRow(ctx, "UPDATE clientes SET saldo = saldo + $1 WHERE id = $2 returning id, limite, saldo", value, customer.ID).Scan(&customer.ID, &customer.Limit, &customer.Amout)
	if err != nil {
		return nil, config.HandleDatabaseError(err)
	}

	return customer, nil
}
