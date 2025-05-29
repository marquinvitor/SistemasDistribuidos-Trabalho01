package Control

import (
	"fmt"
	"io"
	"strconv"
	"meu_rh/Models"
)

type ColaboradorOuputStream struct {
	Destino       io.Writer
	Colaboradores []Models.Colaborador
}

func NewColaboradorOutputStream(destino io.Writer, dados []Models.Colaborador) *ColaboradorOuputStream {
	return &ColaboradorOuputStream{
		Destino:       destino,
		Colaboradores: dados,
	}
}

func (cos *ColaboradorOuputStream) EscreverTodosOsDados() error {
	numObjetos := len(cos.Colaboradores)
	cabecalho := fmt.Sprintf("Enviando dados de %d objetos...\n---\n", numObjetos)
	_, err := cos.Destino.Write([]byte(cabecalho))

	if err != nil {
		return fmt.Errorf("falha ao escrever cabecalho: %w", err)
	}

	for _, colab := range cos.Colaboradores {
		idStr := strconv.Itoa(colab.GetId())
		nome := colab.GetNome()
		salarioStr := fmt.Sprintf("%.2f", colab.CalcularSalario())

		linha := fmt.Sprintf("|ID: %s, Nome: %s Salario: %s|", idStr, nome, salarioStr)

		dadosBytes := []byte(linha)

		bytesEscritos, err := cos.Destino.Write(dadosBytes)

		if err != nil {
			return fmt.Errorf("falha ao escrever os dados: %w", err)
		}

		fmt.Printf(" Foram escritos %d bytes", bytesEscritos)

		final := "---\nEnvio concluído.\n"
		_, err = cos.Destino.Write([]byte(final))

		if err != nil {
			return fmt.Errorf("falha ao escrever finalização: %w", err)
		}

	}
	return nil
}
