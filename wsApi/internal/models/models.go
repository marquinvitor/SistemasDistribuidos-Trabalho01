package models
import (
	"fmt"
)

type Colaborador interface {
	Identificar() string
	CalcularSalario() float64
	GetId() int
	GetNome() string
}

///////////////////////////////////////////////////////////////

type ColaboradorBase struct {
	Id   int
	Nome string
}

func (c ColaboradorBase) Identificar() string {
    return fmt.Sprintf("ID - %d | Nome - %s", c.Id, c.Nome)
}

func (c ColaboradorBase) GetId() int {
	return c.Id
}

func (c ColaboradorBase) GetNome() string {
	return c.Nome
}

///////////////////////////////////////////////////////////////

type Autonomo struct{
	ColaboradorBase
	HorasTrabalhadas int
	ValorHora float64
}

func(a Autonomo)CalcularSalario()float64{
	return a.ValorHora * float64(a.HorasTrabalhadas)
}

///////////////////////////////////////////////////////////////


type Efetivo struct{
	ColaboradorBase
	SalarioMensal float64
}

func (e Efetivo) CalcularSalario() float64{
	return e.SalarioMensal
}

type Estagiario struct{
	ColaboradorBase
	AuxilioEstagio float64
}

func (e Estagiario) CalcularSalario() float64{
	return e.AuxilioEstagio
}

///////////////////////////////////////////////////////////////

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