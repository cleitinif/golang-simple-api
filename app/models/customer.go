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
