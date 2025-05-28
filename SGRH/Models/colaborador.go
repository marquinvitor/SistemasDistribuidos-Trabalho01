package Models

import (
	"strconv"
)

type ColaboradorBase struct {
	Id   int
	Nome string
}

func (c ColaboradorBase) Identificar() string {
	return "ID - " + strconv.Itoa(c.Id) + "Nome - " + c.Nome
}

func (c ColaboradorBase) GetId() int {
	return c.Id
}

func (c ColaboradorBase) GetNome() string {
	return c.Nome
}


