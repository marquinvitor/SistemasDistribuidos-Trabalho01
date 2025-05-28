package Models

type Colaborador interface {
	Identificar() string
	CalcularSalario() float64
	GetId() int
	GetNome() string
}
