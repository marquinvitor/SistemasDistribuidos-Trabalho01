package main

type Efetivo struct{
	ColaboradorBase
	SalarioMensal float64
}

func (e Efetivo) CalcularSalario() float64{
	return e.SalarioMensal
}

