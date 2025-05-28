package main

import "fmt"

type Departamento struct {
	Nome          string
	Colaboradores []Colaborador
}

func (d *Departamento) AdicionarColaborador(c Colaborador) {
	d.Colaboradores = append(d.Colaboradores, c)
}

func (d *Departamento) DemitirColaborador(id int)error {
	indiceParaRemover := -1
	for i := 0; i < len(d.Colaboradores); i++ {
		if d.Colaboradores[i].GetId() == id {
			indiceParaRemover = i
			break
		}
	}

	if indiceParaRemover == -1 {
		return fmt.Errorf("demissão falhou: colaborador com ID %d não encontrado no departamento", id)
	}
	d.Colaboradores = append(d.Colaboradores[:indiceParaRemover], d.Colaboradores[indiceParaRemover+1:]...)
	return nil
}

func (d *Departamento) CalcularFolhaSalarial() float64 {
	total := 0.0

	for _, c := range d.Colaboradores {
		total += c.CalcularSalario()
	}
	return total
}

func(d *Departamento)ListarColaboradores(){
	fmt.Printf("--- Colaboradores do Depto. de %s ---\n", d.Nome)
	for _, colab := range d.Colaboradores {
		fmt.Printf(" - ID: %d, Nome: %s\n", colab.GetId(), colab.GetNome())
	}
	fmt.Println("---------------------------------------")
}