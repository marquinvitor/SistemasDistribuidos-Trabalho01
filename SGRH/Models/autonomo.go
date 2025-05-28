package Models

type Autonomo struct{
	ColaboradorBase
	HorasTrabalhadas int
	ValorHora float64
}

func(a Autonomo)CalcularSalario()float64{
	return a.ValorHora * float64(a.HorasTrabalhadas)
}