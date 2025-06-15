package main

import (
	"context"
	"log"
	"time"

	 pb "rmi/proto" 

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Conecta ao servidor RMI sem usar sockets diretamente
	conn, err := grpc.NewClient("localhost:8088", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Não foi possível conectar: %v", err)
	}
	defer conn.Close()

	// Cria um cliente (stub) para o serviço SGRH
	client := pb.NewSGRHClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// --- Exemplo de uso ---

	// 1. Adicionar Colaboradores
	log.Println("--- Adicionando Colaboradores ---")
	colabEfetivo := &pb.Colaborador{
		Id:   201,
		Nome: "Roberto Carlos",
		Tipo: pb.TipoColaborador_EFETIVO,
		DetalhesSalario: &pb.Colaborador_SalarioMensal{
			SalarioMensal: 7500.00,
		},
	}
	respAdd, err := client.AdicionarColaborador(ctx, &pb.AddColaboradorRequest{NomeDepartamento: "TI", Colaborador: colabEfetivo})
	if err != nil {
		log.Fatalf("Erro ao adicionar: %v", err)
	}
	log.Printf("Resposta do servidor: %s", respAdd.Message)

	// 2. Listar Colaboradores
	log.Println("\n--- Listando Colaboradores em TI ---")
	respList, err := client.ListarColaboradores(ctx, &pb.ListarColaboradoresRequest{NomeDepartamento: "TI"})
	if err != nil {
		log.Fatalf("Erro ao listar: %v", err)
	}
	for _, c := range respList.Colaboradores {
		log.Printf("  - ID: %d, Nome: %s, Tipo: %s", c.Id, c.Nome, c.Tipo)
	}

	// 3. Calcular Folha Salarial
	log.Println("\n--- Calculando Folha Salarial de TI ---")
	respFolha, err := client.CalcularFolhaSalarial(ctx, &pb.CalcularFolhaSalarialRequest{NomeDepartamento: "TI"})
	if err != nil {
		log.Fatalf("Erro ao calcular folha: %v", err)
	}
	log.Printf("Total da folha: R$ %.2f", respFolha.TotalFolha)

	// 4. Demitir Colaborador
	log.Println("\n--- Demitindo Colaborador ID 201 ---")
	respDemitir, err := client.DemitirColaborador(ctx, &pb.DemitirColaboradorRequest{NomeDepartamento: "TI", ColaboradorId: 201})
	if err != nil {
		log.Fatalf("Erro ao demitir: %v", err)
	}
	log.Printf("Resposta do servidor: %s", respDemitir.Message)

	// 5. Listar novamente para confirmar
	log.Println("\n--- Listando Colaboradores em TI (após demissão) ---")
	respList2, err := client.ListarColaboradores(ctx, &pb.ListarColaboradoresRequest{NomeDepartamento: "TI"})
	if err != nil {
		log.Fatalf("Erro ao listar: %v", err)
	}
	if len(respList2.Colaboradores) == 0 {
		log.Println("Nenhum colaborador no departamento.")
	} else {
		for _, c := range respList2.Colaboradores {
			log.Printf("  - ID: %d, Nome: %s, Tipo: %s", c.Id, c.Nome, c.Tipo)
		}
	}
}