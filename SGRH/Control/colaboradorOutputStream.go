package Control

import(
	"io"
	"meu_rh/Models"
)

type ColaboradorOuputStream struct{
	Destino io.Writer
	colaboradores []Models.Colaborador
}