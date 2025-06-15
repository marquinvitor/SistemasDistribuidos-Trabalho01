package main

import (
	"context"
	"log"
	"net"
	"sync"

	pb "rmi/proto" 

	"google.golang.org/grpc"
)


type Colaborador struct {
	pb.Colaborador
}

func (c *Colaborador) CalcularSalario() float64 {
	switch c.Tipo {
	case pb.TipoColaborador_EFETIVO:
		return c.GetSalarioMensal()
	case pb.TipoColaborador_AUTONOMO:
		return c.GetAutonomo().ValorHora * float64(c.GetAutonomo().HorasTrabalhadas)
	case pb.TipoColaborador_ESTAGIARIO:
		return c.GetAuxilioEstagio()
	}
	return 0.0
}

type Departamento struct {
	Nome          string
	Colaboradores map[int32]*Colaborador
}

// Estrutura do servidor que implementa a interface SGRH
type sgrhServer struct {
	pb.UnimplementedSGRHServer
	mu            sync.Mutex
	departamentos map[string]*Departamento
}


func (s *sgrhServer) AdicionarColaborador(ctx context.Context, req *pb.AddColaboradorRequest) (*pb.SGRHResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	dep, exists := s.departamentos[req.NomeDepartamento]
	if !exists {
		dep = &Departamento{
			Nome:          req.NomeDepartamento,
			Colaboradores: make(map[int32]*Colaborador),
		}
		s.departamentos[req.NomeDepartamento] = dep
	}

	colabID := req.Colaborador.Id
	if _, exists := dep.Colaboradores[colabID]; exists {
		return &pb.SGRHResponse{Success: false, Message: "Colaborador com este ID já existe no departamento."}, nil
	}

	dep.Colaboradores[colabID] = &Colaborador{*req.Colaborador}
	log.Printf("Colaborador %s (ID: %d) adicionado ao depto %s", req.Colaborador.Nome, colabID, req.NomeDepartamento)
	return &pb.SGRHResponse{Success: true, Message: "Colaborador adicionado com sucesso."}, nil
}


func (s *sgrhServer) DemitirColaborador(ctx context.Context, req *pb.DemitirColaboradorRequest) (*pb.SGRHResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	dep, exists := s.departamentos[req.NomeDepartamento]
	if !exists {
		return &pb.SGRHResponse{Success: false, Message: "Departamento não encontrado."}, nil
	}

	if _, exists := dep.Colaboradores[req.ColaboradorId]; !exists {
		return &pb.SGRHResponse{Success: false, Message: "Colaborador não encontrado no departamento."}, nil
	}

	delete(dep.Colaboradores, req.ColaboradorId)
	log.Printf("Colaborador ID: %d demitido do depto %s", req.ColaboradorId, req.NomeDepartamento)
	return &pb.SGRHResponse{Success: true, Message: "Colaborador demitido com sucesso."}, nil
}


func (s *sgrhServer) ListarColaboradores(ctx context.Context, req *pb.ListarColaboradoresRequest) (*pb.ListarColaboradoresResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	dep, exists := s.departamentos[req.NomeDepartamento]
	if !exists {
		return &pb.ListarColaboradoresResponse{Colaboradores: nil}, nil
	}

	var colabsProto []*pb.Colaborador
	for _, c := range dep.Colaboradores {
		colabsProto = append(colabsProto, &c.Colaborador)
	}

	return &pb.ListarColaboradoresResponse{Colaboradores: colabsProto}, nil
}


func (s *sgrhServer) CalcularFolhaSalarial(ctx context.Context, req *pb.CalcularFolhaSalarialRequest) (*pb.CalcularFolhaSalarialResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	dep, exists := s.departamentos[req.NomeDepartamento]
	if !exists {
		return &pb.CalcularFolhaSalarialResponse{TotalFolha: 0.0}, nil
	}

	total := 0.0
	for _, c := range dep.Colaboradores {
		total += c.CalcularSalario()
	}
	log.Printf("Folha do depto %s calculada: R$ %.2f", req.NomeDepartamento, total)
	return &pb.CalcularFolhaSalarialResponse{TotalFolha: total}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":8088")
	if err != nil {
		log.Fatalf("Falha ao escutar na porta: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterSGRHServer(grpcServer, &sgrhServer{
		departamentos: make(map[string]*Departamento),
	})

	log.Println("Servidor RMI (gRPC) escutando em localhost:8088")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Falha ao servir: %v", err)
	}
}