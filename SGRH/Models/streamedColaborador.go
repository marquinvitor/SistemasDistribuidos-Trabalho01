package Models

type StreamedColaborador struct {
	ColaboradorBase
	SalarioCalculado float64
}

func (sc StreamedColaborador) CalcularSalario() float64 {
	return sc.SalarioCalculado
}

var _ Colaborador = StreamedColaborador{}