package models

type Customer struct {
	ID    int `json:"id"`
	Limit int `json:"limite"`
	Amout int `json:"saldo"`
}

func NewCustomer(id, limit, saldo int) *Customer {
	return &Customer{
		ID:    id,
		Limit: limit,
		Amout: saldo,
	}
}

func (c *Customer) Withdraw(amount int) {
	c.Amout -= amount
}

func (c *Customer) Deposit(amount int) {
	c.Amout += amount
}

func (c *Customer) CanWithdraw(amount int) bool {
	return c.Amout-amount >= c.Limit
}
