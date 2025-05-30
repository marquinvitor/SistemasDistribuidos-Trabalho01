package main

import (
	"encoding/gob"
	"fmt"
	"io"
	"meu_rh/Models"
	"meu_rh/Shared"
	"net"
	"runtime/debug"
	"sync"
)

type DepartamentoManager struct {
	Departamentos map[string]*Models.Departamento
	mu            sync.Mutex
}

func NewDepartamentoManager() *DepartamentoManager {
	return &DepartamentoManager{
		Departamentos: make(map[string]*Models.Departamento),
	}
}

func (dm *DepartamentoManager) GetOrCreateDepartamento(nome string) *Models.Departamento {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	if dep, exists := dm.Departamentos[nome]; exists {
		return dep
	}
	dep := &Models.Departamento{Nome: nome}
	dm.Departamentos[nome] = dep
	return dep
}

func handleConnection(conn net.Conn, manager *DepartamentoManager) {
	remoteAddr := conn.RemoteAddr().String()
	fmt.Printf("[SERVIDOR %s] Cliente conectado.\n", remoteAddr)

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("------------------------------------------------------\n")
			fmt.Printf("[SERVIDOR %s] PÂNICO FATAL EM handleConnection: %v\n", remoteAddr, r)
			fmt.Printf("STACK TRACE:\n%s\n", string(debug.Stack()))
			fmt.Printf("------------------------------------------------------\n")
		}
		conn.Close()
		fmt.Printf("[SERVIDOR %s] Cliente desconectado.\n", remoteAddr)
	}()

	decoder := gob.NewDecoder(conn)
	encoder := gob.NewEncoder(conn)

	for {
		var req Shared.Request
		if err := decoder.Decode(&req); err != nil {
			if err != io.EOF {
				fmt.Printf("[SERVIDOR %s] Erro ao decodificar requisição: %v\n", remoteAddr, err)
			}
			break
		}

		fmt.Printf("[SERVIDOR %s] Processando operação: %s\n", remoteAddr, req.Operation)

		var resp Shared.Response

		switch req.Operation {
		case Shared.OpAdicionarColaborador:
			if addData, ok := req.Data.(Shared.AddColaboradorRequestData); ok {
				dep := manager.GetOrCreateDepartamento(addData.DepartamentoNome)
				dep.AdicionarColaborador(addData.Colaborador)
				resp = Shared.Response{Success: true, Message: "Colaborador adicionado"}
			} else {
				resp = Shared.Response{Success: false, Message: "Tipo de dado invalido para AdicionarColaborador"}
				fmt.Printf("[SERVIDOR %s] ERRO na op '%s': Tipo de dado inválido. Recebido: %T\n", remoteAddr, req.Operation, req.Data)
			}
		case Shared.OpDemitirColaborador:
			if demitirData, ok := req.Data.(Shared.DemitirColaboradorRequestData); ok {
				dep := manager.GetOrCreateDepartamento(demitirData.DepartamentoNome)
				err := dep.DemitirColaborador(demitirData.ColaboradorID)
				if err != nil {
					resp = Shared.Response{Success: false, Message: err.Error()}
				} else {
					resp = Shared.Response{Success: true, Message: "Colaborador demitido"}
				}
			} else {
				resp = Shared.Response{Success: false, Message: "Tipo de dado invalido para DemitirColaborador"}
				fmt.Printf("[SERVIDOR %s] ERRO na op '%s': Tipo de dado inválido. Recebido: %T\n", remoteAddr, req.Operation, req.Data)
			}
		case Shared.OpCalcularFolha:
			if deptData, ok := req.Data.(Shared.DepartamentoRequestData); ok {
				dep := manager.GetOrCreateDepartamento(deptData.DepartamentoNome)
				totalFolha := dep.CalcularFolhaSalarial()
				resp = Shared.Response{Success: true, Data: totalFolha}
			} else {
				resp = Shared.Response{Success: false, Message: "Tipo de dado invalido para CalcularFolha"}
				fmt.Printf("[SERVIDOR %s] ERRO na op '%s': Tipo de dado inválido. Recebido: %T\n", remoteAddr, req.Operation, req.Data)
			}
		case Shared.OpListarColaboradores:
			if deptData, ok := req.Data.(Shared.DepartamentoRequestData); ok {
				dep := manager.GetOrCreateDepartamento(deptData.DepartamentoNome)
				resp = Shared.Response{Success: true, Data: dep.Colaboradores}
			} else {
				resp = Shared.Response{Success: false, Message: "Tipo de dado invalido para ListarColaboradores"}
				fmt.Printf("[SERVIDOR %s] ERRO na op '%s': Tipo de dado inválido. Recebido: %T\n", remoteAddr, req.Operation, req.Data)
			}
		default:
			resp = Shared.Response{Success: false, Message: "Operacao desconhecida"}
			fmt.Printf("[SERVIDOR %s] ERRO: Operação desconhecida '%s'\n", remoteAddr, req.Operation)
		}

		if err := encoder.Encode(resp); err != nil {
			fmt.Printf("[SERVIDOR %s] Erro ao codificar/enviar resposta para op '%s': %v\n", remoteAddr, req.Operation, err)
			break
		}
	}
}

func main() {
	endereco := "localhost:8088"
	listener, err := net.Listen("tcp", endereco)
	if err != nil {
		fmt.Printf("[SERVIDOR] Falha fatal ao iniciar servidor: %v\n", err)
		return
	}
	defer listener.Close()
	fmt.Printf("[SERVIDOR] Servico de RH Remoto escutando em %s\n", endereco)

	manager := NewDepartamentoManager()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("[SERVIDOR] Falha ao aceitar nova conexão: %v\n", err)
			continue
		}
		go handleConnection(conn, manager)
	}
}