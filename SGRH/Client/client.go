package main

import (
	"encoding/gob"
	"fmt"
	"meu_rh/Models"
	"meu_rh/Shared"
	"net"
)

type RHServiceClient struct {
	conn    net.Conn
	encoder *gob.Encoder
	decoder *gob.Decoder
}

func NewRHServiceClient(enderecoServidor string) (*RHServiceClient, error) {
	conn, err := net.Dial("tcp", enderecoServidor)
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar: %w", err)
	}
	return &RHServiceClient{
		conn:    conn,
		encoder: gob.NewEncoder(conn),
		decoder: gob.NewDecoder(conn),
	}, nil
}

func (c *RHServiceClient) Close() {
	c.conn.Close()
}

func (c *RHServiceClient) AdicionarColaborador(deptNome string, colab Models.Colaborador) (Shared.Response, error) {
	reqData := Shared.AddColaboradorRequestData{
		DepartamentoNome: deptNome,
		Colaborador:      colab,
	}
	req := Shared.Request{Operation: Shared.OpAdicionarColaborador, Data: reqData}
	if err := c.encoder.Encode(req); err != nil {
		return Shared.Response{}, err
	}
	var resp Shared.Response
	err := c.decoder.Decode(&resp)
	return resp, err
}

func (c *RHServiceClient) DemitirColaborador(deptNome string, colabID int) (Shared.Response, error) {
	reqData := Shared.DemitirColaboradorRequestData{
		DepartamentoNome: deptNome,
		ColaboradorID:    colabID,
	}
	req := Shared.Request{Operation: Shared.OpDemitirColaborador, Data: reqData}
	if err := c.encoder.Encode(req); err != nil {
		return Shared.Response{}, err
	}
	var resp Shared.Response
	err := c.decoder.Decode(&resp)
	return resp, err
}

func (c *RHServiceClient) CalcularFolhaSalarial(deptNome string) (Shared.Response, error) {
	reqData := Shared.DepartamentoRequestData{DepartamentoNome: deptNome}
	req := Shared.Request{Operation: Shared.OpCalcularFolha, Data: reqData}
	if err := c.encoder.Encode(req); err != nil {
		return Shared.Response{}, err
	}
	var resp Shared.Response
	err := c.decoder.Decode(&resp)
	return resp, err
}

func (c *RHServiceClient) ListarColaboradores(deptNome string) (Shared.Response, error) {
	reqData := Shared.DepartamentoRequestData{DepartamentoNome: deptNome}
	req := Shared.Request{Operation: Shared.OpListarColaboradores, Data: reqData}
	if err := c.encoder.Encode(req); err != nil {
		return Shared.Response{}, err
	}
	var resp Shared.Response
	err := c.decoder.Decode(&resp)
	return resp, err
}

func main() {
	client, err := NewRHServiceClient("localhost:8088")
	if err != nil {
		fmt.Printf("Erro do cliente: %v\n", err)
		return
	}
	defer client.Close()

	colabEfetivo := Models.Efetivo{
		ColaboradorBase: Models.ColaboradorBase{Id: 201, Nome: "Roberto Carlos"},
		SalarioMensal:   7500.00,
	}
	colabAutonomo := Models.Autonomo{
		ColaboradorBase:  Models.ColaboradorBase{Id: 202, Nome: "Laura Matos"},
		ValorHora:        120.00,
		HorasTrabalhadas: 50,
	}

	respAdd, err := client.AdicionarColaborador("TI", colabEfetivo)
	fmt.Printf("Adicionar Efetivo (TI): Success: %t, Msg: %s, Err: %v\n", respAdd.Success, respAdd.Message, err)

	respAdd, err = client.AdicionarColaborador("TI", colabAutonomo)
	fmt.Printf("Adicionar Autonomo (TI): Success: %t, Msg: %s, Err: %v\n", respAdd.Success, respAdd.Message, err)

	respListTI, err := client.ListarColaboradores("TI")
	fmt.Printf("Listar (TI): Success: %t, Msg: %s, Err: %v\n", respListTI.Success, respListTI.Message, err)
	if respListTI.Success && respListTI.Data != nil {
		if colabs, ok := respListTI.Data.([]Models.Colaborador); ok {
			fmt.Println("Colaboradores em TI:")
			for _, c := range colabs {
				fmt.Printf("  ID: %d, Nome: %s, Salario: %.2f\n", c.GetId(), c.GetNome(), c.CalcularSalario())
			}
		}
	}

	respFolhaTI, err := client.CalcularFolhaSalarial("TI")
	fmt.Printf("Calcular Folha (TI): Success: %t, Msg: %s, Err: %v\n", respFolhaTI.Success, respFolhaTI.Message, err)
	if respFolhaTI.Success {
		if total, ok := respFolhaTI.Data.(float64); ok {
			fmt.Printf("Total da folha salarial de TI: %.2f\n", total)
		}
	}

	respDemitir, err := client.DemitirColaborador("TI", 201)
	fmt.Printf("Demitir (ID 201 de TI): Success: %t, Msg: %s, Err: %v\n", respDemitir.Success, respDemitir.Message, err)

	respListTI2, err := client.ListarColaboradores("TI")
	fmt.Printf("Listar (TI) apos demissao: Success: %t, Msg: %s, Err: %v\n", respListTI2.Success, respListTI2.Message, err)
	if respListTI2.Success && respListTI2.Data != nil {
		if colabs, ok := respListTI2.Data.([]Models.Colaborador); ok {
			fmt.Println("Colaboradores em TI (apos demissao):")
			for _, c := range colabs {
				fmt.Printf("  ID: %d, Nome: %s, Salario: %.2f\n", c.GetId(), c.GetNome(), c.CalcularSalario())
			}
		}
	}
}