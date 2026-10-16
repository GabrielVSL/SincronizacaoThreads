package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ============================================================================
// 6. main.go — Menu Principal e Resumo do Problema
// ============================================================================
// Estrutura de 6 Arquivos (100% fiel ao PDF):
//   1. semaforo.go  -> Semáforo na unha (Down e Up).
//   2. barbearia.go -> Recurso compartilhado ('waiting') e os 3 semáforos ('customers', 'barbers', 'mutex').
//   3. barbeiro.go  -> Thread do Barbeiro (Barbeiro e cortarCabelo).
//   4. cliente.go   -> Thread do Cliente (Cliente e receberCorte).
//   5. validacao.go -> Simulações visuais e os 3 testes de validação.
//   6. main.go      -> Menu interativo e inicialização do programa.

func main() {
	leitor := bufio.NewReader(os.Stdin)

	for {
		fmt.Println()
		fmt.Println(strings.Repeat("=", 95))
		fmt.Println("   PUC MINAS - SISTEMAS OPERACIONAIS | TRABALHO PRATICO 1 (PARTE II - ITEM B)")
		fmt.Println("   PROBLEMA DO BARBEIRO DORMINHOCO EM GO (SEMAFOROS IMPLEMENTADOS NA UNHA)")
		fmt.Println(strings.Repeat("=", 95))
		fmt.Println(`
  HISTORIA E MODELAGEM DO PROBLEMA (Tanenbaum, Figura 1.20):
  Na barbearia ha 1 barbeiro, 1 cadeira de barbeiro e N cadeiras de espera:
    * Sem clientes, o barbeiro senta na cadeira e DORME -> bloqueia em customers.Down().
    * Quando chega um cliente, ele obtem 'mutex.Down()' para entrar na Regiao Critica:
        - Se 'waiting < N': incrementa 'waiting++', acorda o barbeiro com 'customers.Up()',
          libera 'mutex.Up()' (evitando Deadlock!) e aguarda sua vez em 'barbers.Down()'.
        - Se 'waiting == N' (lotado): libera 'mutex.Up()' e vai embora sem cortar o cabelo.`)
		fmt.Println(strings.Repeat("-", 95))
		fmt.Println("  MENU DE EXECUCAO:")
		fmt.Println("    [1] Executar Simulacao Visual Padrao (N = 3 cadeiras, 8 clientes em 2 ondas)")
		fmt.Println("    [2] Executar Simulacao Customizada   (Escolher N cadeiras, clientes e tempos)")
		fmt.Println("    [3] Executar Testes de Validacao     (3 baterias de testes contra Race e Deadlock)")
		fmt.Println("    [0] Sair")
		fmt.Println(strings.Repeat("-", 95))
		fmt.Print("  Escolha uma opcao (0 a 3): ")

		entrada, err := leitor.ReadString('\n')
		if err != nil {
			return
		}
		opcao := strings.TrimSpace(entrada)

		switch opcao {
		case "1":
			ExecutarSimulacaoPadrao(leitor)
		case "2":
			ExecutarSimulacaoCustomizada(leitor)
		case "3":
			ExecutarTestesValidacao(leitor)
		case "0":
			fmt.Println("\n  Programa encerrado.")
			return
		default:
			fmt.Printf("\n  [!] Opcao '%s' invalida. Digite 1, 2, 3 ou 0.\n", opcao)
		}
	}
}

func aguardarEnter(leitor *bufio.Reader) {
	fmt.Print("\nPressione [ENTER] para voltar ao Menu...")
	_, _ = leitor.ReadString('\n')
}

func lerInteiro(leitor *bufio.Reader, pergunta string, padrao int, min int, max int) int {
	for {
		fmt.Print(pergunta)
		texto, _ := leitor.ReadString('\n')
		texto = strings.TrimSpace(texto)
		if texto == "" {
			return padrao
		}
		valor, err := strconv.Atoi(texto)
		if err != nil || valor < min || valor > max {
			fmt.Printf("    [!] Digite um numero entre %d e %d (ou ENTER para %d).\n", min, max, padrao)
			continue
		}
		return valor
	}
}
