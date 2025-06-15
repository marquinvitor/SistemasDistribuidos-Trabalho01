package Models

import "encoding/gob"

func init() {
	gob.Register(Efetivo{})
	gob.Register(Autonomo{})
	gob.Register(Estagiario{})
	gob.Register(StreamedColaborador{})
}