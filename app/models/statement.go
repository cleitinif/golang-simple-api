package models

type Balance struct {
	Amount int    `json:"total"`
	Limit  int    `json:"limite"`
	Date   string `json:"data_extrato"`
}

type Statement struct {
	Balance          Balance       `json:"saldo"`
	LastTransactions []Transaction `json:"ultimas_transacoes"`
}
