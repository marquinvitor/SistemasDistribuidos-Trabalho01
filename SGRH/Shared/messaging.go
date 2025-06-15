package Shared

import (
	"encoding/gob"
	"meu_rh/Models"
)

func init() {
	gob.Register(AddColaboradorRequestData{})
	gob.Register(DemitirColaboradorRequestData{})
	gob.Register(DepartamentoRequestData{})

	gob.Register(([]Models.Colaborador)(nil))
}

const (
	OpAdicionarColaborador = "ADICIONAR_COLABORADOR"
	OpDemitirColaborador   = "DEMITIR_COLABORADOR"
	OpCalcularFolha        = "CALCULAR_FOLHA"
	OpListarColaboradores  = "LISTAR_COLABORADORES"
)

type Request struct {
	Operation string
	Data      interface{}
}

type Response struct {
	Success bool
	Message string
	Data    interface{}
}

type AddColaboradorRequestData struct {
	DepartamentoNome string
	Colaborador      Models.Colaborador
}

type DemitirColaboradorRequestData struct {
	DepartamentoNome string
	ColaboradorID    int
}

type DepartamentoRequestData struct {
	DepartamentoNome string
}