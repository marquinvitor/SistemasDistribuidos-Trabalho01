package main

import (
	"bufio"
	"fmt"
	"io"
	"meu_rh/Control"
	"meu_rh/Models"
	"net"
	"os"
	"strings"
	"time"
)

func main() {
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

	allColabs := []Models.Colaborador{colab1, colab2, colab3}

	fmt.Println("\n>>>> INICIANDO TESTE 1: SAÍDA PADRÃO (os.Stdout) <<<<")
	streamParaConsole := Control.NewColaboradorOutputStream(os.Stdout, allColabs)
	err := streamParaConsole.EscreverTodosOsDados()
	if err != nil {
		fmt.Printf("Erro no teste 1: %v\n", err)
	}
	fmt.Println(">>>> FIM DO TESTE 1 <<<<")

	fmt.Println("\n>>>> INICIANDO TESTE 2: ARQUIVO (colaboradores.txt) <<<<")
	nomeArquivo := "colaboradores.txt"
	arquivo, err := os.Create(nomeArquivo)
	if err != nil {
		fmt.Printf("Erro ao criar arquivo para o teste 2: %v\n", err)
	} else {
		streamToArquivo := Control.NewColaboradorOutputStream(arquivo, allColabs)
		err = streamToArquivo.EscreverTodosOsDados()
		closeErr := arquivo.Close()
		if closeErr != nil {
			fmt.Printf("Erro ao fechar o arquivo apos escrita: %v\n", closeErr)
		}
		if err != nil {
			fmt.Printf("Erro no teste 2: %v\n", err)
		} else {
			fmt.Printf("Dados escritos com sucesso no arquivo '%s'\n", nomeArquivo)
		}
	}
	fmt.Println(">>>> FIM DO TESTE 2 <<<<")

	fmt.Println("\n>>>> INICIANDO TESTE 3: SERVIDOR TCP (Cliente OutputStream envia para Servidor) <<<<")
	enderecoServidorOutput := "localhost:8081"
	go func(addr string) {
		listener, listenErr := net.Listen("tcp", addr)
		if listenErr != nil {
			fmt.Printf("[SERVIDOR %s @ Teste 3] Erro ao iniciar: %v\n", addr, listenErr)
			return
		}
		defer listener.Close()
		fmt.Printf("[SERVIDOR %s @ Teste 3] Escutando...\n", addr)
		conn, acceptErr := listener.Accept()
		if acceptErr != nil {
			fmt.Printf("[SERVIDOR %s @ Teste 3] Erro ao aceitar: %v\n", addr, acceptErr)
			return
		}
		defer conn.Close()
		fmt.Printf("[SERVIDOR %s @ Teste 3] Cliente conectado. Lendo dados:\n", addr)
		io.Copy(os.Stdout, conn)
		fmt.Printf("\n[SERVIDOR %s @ Teste 3] Conexão finalizada.\n", addr)
	}(enderecoServidorOutput)

	time.Sleep(1 * time.Second)
	connOutput, err := net.Dial("tcp", enderecoServidorOutput)
	if err != nil {
		fmt.Printf("Erro ao conectar ao servidor para o teste 3: %v\n", err)
	} else {
		defer connOutput.Close()
		streamToRede := Control.NewColaboradorOutputStream(connOutput, allColabs)
		err = streamToRede.EscreverTodosOsDados()
		if err != nil {
			fmt.Printf("Erro no teste 3 ao enviar dados: %v\n", err)
		} else {
			fmt.Println("Dados enviados com sucesso para o servidor TCP no Teste 3.")
		}
	}
	fmt.Println(">>>> FIM DO TESTE 3 <<<<")
	time.Sleep(1 * time.Second)

	fmt.Println("\n>>>> INICIANDO TESTE 4: ENTRADA PADRÃO (os.Stdin) - InputStream <<<<")
	fmt.Println("Cole/digite os dados formatados (Ctrl+D/Ctrl+Z para EOF):")
	
	var sb strings.Builder
	fmt.Fprintf(&sb, "Enviando dados de %d objetos...\n", len(allColabs))
	fmt.Fprintln(&sb, "---")
	for _, c := range allColabs {
		fmt.Fprintf(&sb, "|ID: %d, Nome: %s Salario: %.2f|\n", c.GetId(), c.GetNome(), c.CalcularSalario())
		fmt.Fprintln(&sb, "---")
		fmt.Fprintln(&sb, "Envio concluído.")
	}
	fmt.Println("Exemplo de entrada formatada (para 3 objetos):")
	fmt.Print(sb.String())

	streamDaConsole := Control.NewColaboradorInputStream(bufio.NewReader(os.Stdin))
	dadosLidosConsole, err := streamDaConsole.LerTodosOsDados()
	if err != nil {
		fmt.Printf("Erro no teste 4 ao ler da entrada padrão: %v\n", err)
	} else {
		fmt.Println("Dados lidos da entrada padrão:")
		for _, colab := range dadosLidosConsole {
			fmt.Printf("  ID: %d, Nome: %s, Salário Lido: %.2f\n",
				colab.GetId(), colab.GetNome(), colab.CalcularSalario())
		}
	}
	fmt.Println(">>>> FIM DO TESTE 4 <<<<")

	fmt.Println("\n>>>> INICIANDO TESTE 5: ARQUIVO (colaboradores.txt) - InputStream <<<<")
	arquivoLeitura, err := os.Open(nomeArquivo)
	if err != nil {
		fmt.Printf("Erro ao abrir arquivo para o teste 5: %v\n", err)
	} else {
		streamDoArquivo := Control.NewColaboradorInputStream(arquivoLeitura)
		dadosLidosArquivo, err := streamDoArquivo.LerTodosOsDados()
		if err != nil {
			fmt.Printf("Erro no teste 5 ao ler do arquivo: %v\n", err)
		} else {
			fmt.Printf("Dados lidos com sucesso do arquivo '%s':\n", nomeArquivo)
			for _, colab := range dadosLidosArquivo {
				if sc, ok := colab.(Models.StreamedColaborador); ok {
					fmt.Printf("  ID: %d, Nome: %s, Salário Lido: %.2f\n", sc.GetId(), sc.GetNome(), sc.SalarioCalculado)
				}
			}
		}
		arquivoLeitura.Close()
	}
	fmt.Println(">>>> FIM DO TESTE 5 <<<<")

	fmt.Println("\n>>>> INICIANDO TESTE 6: CLIENTE TCP (InputStream lendo do Servidor) <<<<")
	enderecoServidorInput := "localhost:8082"
	go func(addr string, dataToSend []Models.Colaborador) {
		listener, listenErr := net.Listen("tcp", addr)
		if listenErr != nil {
			fmt.Printf("[SERVIDOR %s @ Teste 6] Erro ao iniciar: %v\n", addr, listenErr)
			return
		}
		defer listener.Close()
		fmt.Printf("[SERVIDOR %s @ Teste 6] Escutando...\n", addr)
		conn, acceptErr := listener.Accept()
		if acceptErr != nil {
			fmt.Printf("[SERVIDOR %s @ Teste 6] Erro ao aceitar: %v\n", addr, acceptErr)
			return
		}
		defer conn.Close()
		fmt.Printf("[SERVIDOR %s @ Teste 6] Cliente conectado. Enviando dados...\n", addr)
		outputStream := Control.NewColaboradorOutputStream(conn, dataToSend)
		sendErr := outputStream.EscreverTodosOsDados()
		if sendErr != nil {
			fmt.Printf("[SERVIDOR %s @ Teste 6] Erro ao enviar dados: %v\n", addr, sendErr)
		} else {
			fmt.Printf("[SERVIDOR %s @ Teste 6] Dados enviados.\n", addr)
		}
	}(enderecoServidorInput, allColabs)

	time.Sleep(1 * time.Second)
	connInput, err := net.Dial("tcp", enderecoServidorInput)
	if err != nil {
		fmt.Printf("Erro ao conectar ao servidor para o teste 6: %v\n", err)
	} else {
		defer connInput.Close()
		streamDaRede := Control.NewColaboradorInputStream(connInput)
		dadosLidosRede, err := streamDaRede.LerTodosOsDados()
		if err != nil {
			fmt.Printf("Erro no teste 6 ao ler dados da rede: %v\n", err)
		} else {
			fmt.Println("Dados lidos com sucesso do servidor TCP no Teste 6:")
			for _, colab := range dadosLidosRede {
				if sc, ok := colab.(Models.StreamedColaborador); ok {
					fmt.Printf("  ID: %d, Nome: %s, Salário Lido: %.2f\n", sc.GetId(), sc.GetNome(), sc.SalarioCalculado)
				}
			}
		}
	}
	fmt.Println(">>>> FIM DO TESTE 6 <<<<")
	time.Sleep(1 * time.Second)
}