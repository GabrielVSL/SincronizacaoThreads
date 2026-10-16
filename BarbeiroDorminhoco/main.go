package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ============================================================================
// 6. main.go — Menu Principal Estilizado e Resumo do Problema
// ============================================================================
// Estrutura de 6 Arquivos (100% fiel ao PDF):
//   1. semaforo.go  -> Semáforo na unha (Down e Up).
//   2. barbearia.go -> Variável 'waiting', os 3 semáforos ('customers', 'barbers', 'mutex') e Snapshots.
//   3. barbeiro.go  -> Thread do Barbeiro (Barbeiro e cortarCabelo).
//   4. cliente.go   -> Thread do Cliente (Cliente e receberCorte).
//   5. validacao.go -> Simulações em 3 etapas (Ao Vivo + Histórico + Replay) e Testes.
//   6. main.go      -> Menu interativo e inicialização do programa.

func main() {
	leitor := bufio.NewReader(os.Stdin)

	for {
		limparTela()

		// Cabeçalho (Largura = 114 colunas)
		cabecalho := lipgloss.NewStyle().
			Bold(true).
			Foreground(corBranco).
			Background(lipgloss.Color("24")).
			Padding(0, 2).
			Width(larguraTotal).
			Render("PUC MINAS - SISTEMAS OPERACIONAIS | TRABALHO PRATICO 1 (PARTE II - ITEM B)\nPROBLEMA DO BARBEIRO DORMINHOCO EM GO (SEMAFOROS IMPLEMENTADOS NA UNHA)")
		fmt.Println(cabecalho)

		// Caixa da Esquerda (54 + 2 de borda = 56 colunas, 8 linhas sem quebras órfãs)
		textoHistoria := lipgloss.NewStyle().Bold(true).Foreground(corTitulo).Render("HISTORIA DO PROBLEMA (Tanenbaum, Fig. 1.20)") + "\n\n" +
			"A barbearia tem 1 barbeiro e N cadeiras:\n" +
			"* Sem clientes: barbeiro DORME na cadeira.\n" +
			"* Primeiro cliente: ACORDA o barbeiro.\n" +
			"* Barbeiro ocupado e ha vaga (waiting < N):\n" +
			"  o cliente SENTA na espera e aguarda a vez.\n" +
			"* Sala lotada (waiting == N): VAI EMBORA."

		caixaHistoria := lipgloss.NewStyle().
			Border(bordaPadrao).
			BorderForeground(corTitulo).
			Padding(0, 2).
			Width(larguraMeiaCaixa).
			Height(8).
			Render(textoHistoria)

		// Caixa da Direita (54 + 2 de borda = 56 colunas, 8 linhas sem quebras órfãs)
		textoSemaforos := lipgloss.NewStyle().Bold(true).Foreground(corDestaque).Render("MODELAGEM (3 SEMAFOROS + 1 VARIAVEL)") + "\n\n" +
			"* " + lipgloss.NewStyle().Bold(true).Foreground(corAmarelo).Render("waiting (0..N)") + " : Clientes na sala de espera\n" +
			"* " + lipgloss.NewStyle().Bold(true).Foreground(corAzul).Render("customers = 0 ") + " : Conta clientes esperando\n" +
			"* " + lipgloss.NewStyle().Bold(true).Foreground(corVerde).Render("barbers   = 0 ") + " : Indica barbeiro disponivel\n" +
			"* " + lipgloss.NewStyle().Bold(true).Foreground(corDestaque).Render("mutex     = 1 ") + " : Exclusao mutua (waiting)\n\n" +
			lipgloss.NewStyle().Foreground(corCinza).Render("Anti-Deadlock: mutex.Up() ANTES de esperar!")

		caixaSemaforos := lipgloss.NewStyle().
			Border(bordaPadrao).
			BorderForeground(corDestaque).
			Padding(0, 2).
			Width(larguraMeiaCaixa).
			Height(8).
			Render(textoSemaforos)

		// 56 + 2 ("  ") + 56 = 114 colunas exatas!
		fmt.Println(lipgloss.JoinHorizontal(lipgloss.Top, caixaHistoria, "  ", caixaSemaforos))

		// Menu Inferior (112 + 2 de borda = 114 colunas exatas!)
		op1 := lipgloss.NewStyle().Bold(true).Foreground(corVerde).Render("[1]") +
			" Executar Simulacao Visual Padrao  (Ao Vivo -> Historico Completo -> Replay Passo a Passo)"
		op2 := lipgloss.NewStyle().Bold(true).Foreground(corAmarelo).Render("[2]") +
			" Executar Simulacao Customizada    (Escolher N cadeiras, clientes e velocidades + Replay)"
		op3 := lipgloss.NewStyle().Bold(true).Foreground(corAzul).Render("[3]") +
			" Executar Testes de Validacao      (3 baterias de testes contra Race Condition e Deadlock)"
		op0 := lipgloss.NewStyle().Bold(true).Foreground(corVermelho).Render("[0]") +
			" Sair do Programa"

		caixaMenu := lipgloss.NewStyle().
			Border(bordaPadrao).
			BorderForeground(corCinza).
			Padding(0, 2).
			Width(larguraCaixaInt).
			Render(fmt.Sprintf("%s\n\n  %s\n  %s\n  %s\n  %s",
				lipgloss.NewStyle().Bold(true).Foreground(corBranco).Render("MENU DE EXECUCAO INTERATIVO:"),
				op1, op2, op3, op0))

		fmt.Println(caixaMenu)
		fmt.Print("\n  Escolha uma opcao (0 a 3): ")

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
			fmt.Printf("\n  [!] Opcao '%s' invalida. Pressione ENTER e tente novamente...", opcao)
			_, _ = leitor.ReadString('\n')
		}
	}
}

func aguardarEnter(leitor *bufio.Reader) {
	fmt.Print("\n  Pressione [ENTER] para voltar ao Menu Principal...")
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
