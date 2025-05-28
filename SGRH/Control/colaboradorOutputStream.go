package Control

import(
	"io"
	"meu_rh/Models"
	"fmt"
)

type ColaboradorOuputStream struct{
	Destino io.Writer
	Colaboradores []Models.Colaborador
}

func NewColaboradorOutputStream(destino io.Writer,dados []Models.Colaborador) *ColaboradorOuputStream{
	return &ColaboradorOuputStream{
		Destino: destino,
		Colaboradores: dados,
	}
}

func(cos *ColaboradorOuputStream)EscreverTodosOsDados()error{
	numObjetos := len(cos.Colaboradores)
	cabecalho := fmt.Sprintf("Enviando dados de %d objetos...\n---\n", numObjetos)
	_, err := cos.Destino.Write([]byte(cabecalho))

	if err != nil{
		return fmt.Errorf("falha ao escrever cabecalho: %w",err)
	}

	//inacabado
}
