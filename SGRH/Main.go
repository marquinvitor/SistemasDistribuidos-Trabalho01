package main

import (
	"fmt"
	"io"
	"meu_rh/Control"
	"meu_rh/Models"
	"net"
	"os"
	"time"	
)

func main(){

	colab1 := Models.Efetivo{
		ColaboradorBase: Models.ColaboradorBase{Id: 101, Nome: "Ana Silva"},
		SalarioMensal:   5000.00,
	}
	colab2 := Models.Estagiario{
		ColaboradorBase: Models.ColaboradorBase{Id: 102, Nome: "Bruno Costa"},
		AuxilioEstagio:  1200.00,
	}
	colab3 := Models.Autonomo{
		ColaboradorBase:  Models.ColaboradorBase{Id: 103, Nome: "Carla Dias"},
		ValorHora:        100.00,
		HorasTrabalhadas: 40,
	}
	
	allColabs := []Models.Colaborador{colab1,colab2,colab3}

	
						//	|----AQUI PASSOU-----|

	fmt.Println("\n>>>> INICIANDO TESTE 1: SAÍDA PADRÃO (os.Stdout) <<<<")
	streamParaConsole := Control.NewColaboradorOutputStream(os.Stdout,allColabs)
	err := streamParaConsole.EscreverTodosOsDados()
	if err != nil {
		fmt.Printf("Erro no teste 1: %v\n", err)
	}
	fmt.Println(">>>> FIM DO TESTE 1 <<<<")
	
						//	|----AQUI PASSOU-----|

	fmt.Println("\n>>>> INICIANDO TESTE 2: ARQUIVO (colaboradores.txt) <<<<")

	nomeArquivo := "colaboradores.txt"
	arquivo, err := os.Create(nomeArquivo)
	if err != nil {
		fmt.Printf("Erro ao criar arquivo para o teste 2: %v\n", err)
	} else {
		defer arquivo.Close()

		streamToArquivo := Control.NewColaboradorOutputStream(arquivo, allColabs)
		err = streamToArquivo.EscreverTodosOsDados()
		if err != nil {
			fmt.Printf("Erro no teste 2: %v\n", err)
		} else {
			fmt.Printf("Dados escritos com sucesso no arquivo '%s'\n", nomeArquivo)
		}
	}
	fmt.Println(">>>> FIM DO TESTE 2 <<<<")
	
						//	|----AQUI PASSOU-----|
	fmt.Println("\n>>>> INICIANDO TESTE 3: SERVIDOR TCP <<<<")
	enderecoServidor := "localhost:8080"	

	go initTcpServer(enderecoServidor)

	time.Sleep(1 * time.Second)
	conn, err := net.Dial("tcp", enderecoServidor)
	if err != nil {
		fmt.Printf("Erro ao conectar ao servidor para o teste 3: %v\n", err)
	} else {
		defer conn.Close()
		streamToRede := Control.NewColaboradorOutputStream(conn, allColabs)
		err = streamToRede.EscreverTodosOsDados()
		if err != nil {
			fmt.Printf("Erro no teste 3: %v\n", err)
		} else {
			fmt.Println("Dados enviados com sucesso para o servidor TCP.")
		}
	}
	fmt.Println(">>>> FIM DO TESTE 3 <<<<")	
	
}

func initTcpServer(endereco string){
	listener, err := net.Listen("tcp", endereco)
	if err != nil {
		fmt.Printf("[SERVIDOR] Erro ao iniciar: %v\n", err)
		return
	}

	defer listener.Close()

	fmt.Printf("[SERVIDOR] Escutando em %s...\n", endereco)
	conn, err := listener.Accept()
	if err != nil {
		fmt.Printf("[SERVIDOR] Erro ao aceitar conexão: %v\n", err)
		return
	}
	defer conn.Close()

	fmt.Println("[SERVIDOR] Cliente conectado. Lendo dados recebidos:")
	io.Copy(os.Stdout, conn)
	fmt.Println("\n[SERVIDOR] Conexão com cliente finalizada.")

	
}


