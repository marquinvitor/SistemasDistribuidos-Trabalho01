package Control

import (
	"bufio"
	"fmt"
	"io"
	"meu_rh/Models"
	"strconv"
	"strings"
)

type ColaboradorInputStream struct {
	Origem  io.Reader
	scanner *bufio.Scanner
}

func NewColaboradorInputStream(origem io.Reader) *ColaboradorInputStream {
	return &ColaboradorInputStream{
		Origem:  origem,
		scanner: bufio.NewScanner(origem),
	}
}

func (cis *ColaboradorInputStream) LerTodosOsDados() ([]Models.Colaborador, error) {
	var colaboradores []Models.Colaborador

	if !cis.scanner.Scan() {
		if err := cis.scanner.Err(); err != nil {
			return nil, fmt.Errorf("falha ao ler cabecalho do stream: %w", err)
		}
		return nil, fmt.Errorf("stream vazio ou formato de cabecalho inesperado")
	}
	headerLine := cis.scanner.Text()

	var numObjetos int
	_, err := fmt.Sscanf(headerLine, "Enviando dados de %d objetos...", &numObjetos)
	if err != nil {
		return nil, fmt.Errorf("formato de cabecalho invalido '%s': %w", headerLine, err)
	}

	if !cis.scanner.Scan() || cis.scanner.Text() != "---" {
		if errSc := cis.scanner.Err(); errSc != nil {
			return nil, fmt.Errorf("falha ao ler separador '---' pos-cabecalho: %w", errSc)
		}

		if numObjetos > 0 || (numObjetos == 0 && cis.scanner.Text() != "") { 
		    return nil, fmt.Errorf("esperado '---' apos cabecalho, obteve: '%s'", cis.scanner.Text())
		}
	}


	if numObjetos == 0 {
		return colaboradores, nil
	}

	for i := 0; i < numObjetos; i++ {
		if !cis.scanner.Scan() {
			if err := cis.scanner.Err(); err != nil {
				return nil, fmt.Errorf("falha ao ler dados do colaborador %d: %w", i+1, err)
			}
			return nil, fmt.Errorf("fim inesperado do stream ao ler colaborador %d de %d", i+1, numObjetos)
		}
		line := cis.scanner.Text()

		if !strings.HasPrefix(line, "|ID:") || !strings.HasSuffix(line, "|") {
			return nil, fmt.Errorf("formato de linha de colaborador invalido para obj %d: '%s'", i+1, line)
		}
		content := strings.Trim(line, "|") 

		idAndRest := strings.SplitN(content, ", Nome: ", 2)
		if len(idAndRest) != 2 {
			return nil, fmt.Errorf("formato de dados de colaborador (esperado 'ID: ..., Nome: ...') invalido na linha '%s' para obj %d", line, i+1)
		}
		idStrPart := strings.TrimPrefix(idAndRest[0], "ID: ")

		nomeAndSalario := idAndRest[1] 
		
		nomeSalarioParts := strings.SplitN(nomeAndSalario, " Salario: ", 2)
		if len(nomeSalarioParts) != 2 {
			return nil, fmt.Errorf("formato de dados de colaborador (esperado 'Nome Salario: ...') invalido na linha '%s' para obj %d", line, i+1)
		}
		nomePart := nomeSalarioParts[0]
		salarioStrPart := nomeSalarioParts[1]
		// FIM DA NOVA LÓGICA DE PARSING

		id, err := strconv.Atoi(idStrPart)
		if err != nil {
			return nil, fmt.Errorf("falha ao converter ID '%s' para obj %d: %w", idStrPart, i+1, err)
		}

		salario, err := strconv.ParseFloat(salarioStrPart, 64)
		if err != nil {
			return nil, fmt.Errorf("falha ao converter Salario '%s' para obj %d: %w", salarioStrPart, i+1, err)
		}

		colab := Models.StreamedColaborador{
			ColaboradorBase:  Models.ColaboradorBase{Id: id, Nome: nomePart},
			SalarioCalculado: salario,
		}
		colaboradores = append(colaboradores, colab)

		// Ler o "---" do footer repetitivo
		if !cis.scanner.Scan() || cis.scanner.Text() != "---" {
			if err := cis.scanner.Err(); err != nil {
				return nil, fmt.Errorf("falha ao ler '---' do footer do colaborador %d: %w", i+1, err)
			}
			return nil, fmt.Errorf("esperado '---' no footer do colaborador %d de %d, obteve: '%s'", i+1, numObjetos, cis.scanner.Text())
		}
		// Ler o "Envio concluído." do footer repetitivo
		if !cis.scanner.Scan() || !strings.Contains(cis.scanner.Text(), "Envio concluído.") {
			if err := cis.scanner.Err(); err != nil {
				return nil, fmt.Errorf("falha ao ler 'Envio concluído.' do footer do colaborador %d: %w", i+1, err)
			}
			return nil, fmt.Errorf("esperado 'Envio concluído.' no footer do colaborador %d de %d, obteve: '%s'", i+1, numObjetos, cis.scanner.Text())
		}
	}

	if err := cis.scanner.Err(); err != nil {
		return colaboradores, fmt.Errorf("erro de scanner ao final da leitura: %w", err)
	}

	return colaboradores, nil
}