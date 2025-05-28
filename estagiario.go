package main

type Estagiario struct{
	ColaboradorBase
	AuxilioEstagio float64
}

func (e Estagiario) CalcularSalario() float64{
	return e.AuxilioEstagio
}
