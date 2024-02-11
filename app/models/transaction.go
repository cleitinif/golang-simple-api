package models

type Transaction struct {
	Value       int    `json:"valor" binding:"required,gt=0,number"`
	Type        string `json:"tipo" binding:"required,oneof=c d"`
	Description string `json:"descricao" binding:"required,max=10,alphanum"`
	Date        string `json:"realizada_em"`
}
