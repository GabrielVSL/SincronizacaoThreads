package main

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// ============================================================================
// 5. validacao.go — Simulação em 3 Etapas, Navegador de Histórico e Testes
// ============================================================================

// ExecutarSimulacaoPadrao executa a simulação em 3 etapas:
//   Etapa 1: Visualização ao vivo limpa na Planta da Barbearia
//   Etapa 2: Tabela completa do Histórico de Eventos
//   Etapa 3: Navegador Interativo do Histórico (Replay Passo a Passo)
func ExecutarSimulacaoPadrao(leitor *bufio.Reader) {
	b := NovaBarbearia(3, 180*time.Millisecond, true)
	go b.Barbeiro()
	time.Sleep(40 * time.Millisecond)

	var wg sync.WaitGroup

	// Onda 1: Chegam 6 clientes (1 corta, 3 sentam nas cadeiras e 2 desistem por lotação)
	for id := 1; id <= 6; id++ {
		wg.Add(1)
		go b.Cliente(id, &wg)
		time.Sleep(25 * time.Millisecond)
	}

	wg.Wait()
	time.Sleep(60 * time.Millisecond)

	// Onda 2: Barbeiro voltou a dormir; chegam mais 2 clientes que o acordam novamente
	for id := 7; id <= 8; id++ {
		wg.Add(1)
		go b.Cliente(id, &wg)
		time.Sleep(40 * time.Millisecond)
	}

	wg.Wait()
	time.Sleep(30 * time.Millisecond)

	b.mutex.Down()
	b.exibirLogs = false
	b.mutex.Up()

	// Etapa 1: Exibe cada acontecimento pausadamente na tela (2.0 segundos por quadro)
	b.ReproduzirAoVivo(2000 * time.Millisecond)

	fmt.Println()
	avisoFim := lipgloss.NewStyle().
		Bold(true).
		Foreground(corVerde).
		Border(bordaPadrao).
		BorderForeground(corVerde).
		Padding(0, 2).
		Width(larguraCaixaInt).
		Render("ETAPA 1 CONCLUIDA! Todos os 8 clientes foram processados.\nPressione [ENTER] para abrir a ETAPA 2 (Historico Completo) e a ETAPA 3 (Navegador Passo a Passo)...")
	fmt.Println(avisoFim)
	_, _ = leitor.ReadString('\n')

	// Inicia a Etapa 2 (Tabela Completa) e Etapa 3 (Navegador Interativo)
	navegarPeloHistorico(leitor, b, 8, 3)
}

// ExecutarSimulacaoCustomizada permite escolher os parâmetros e também abre o Histórico e Replay ao final.
func ExecutarSimulacaoCustomizada(leitor *bufio.Reader) {
	limparTela()
	titulo := lipgloss.NewStyle().
		Bold(true).
		Foreground(corBranco).
		Background(lipgloss.Color("24")).
		Padding(0, 2).
		Width(larguraTotal).
		Render("SIMULACAO CUSTOMIZADA (DEFINA OS PARAMETROS DA BARBEARIA)")
	fmt.Println(titulo)
	fmt.Println()

	cadeiras := lerInteiro(leitor, "  1. Quantas cadeiras de espera (N)?          [Padrao: 4]: ", 4, 1, 6)
	clientes := lerInteiro(leitor, "  2. Quantos clientes vao chegar?             [Padrao: 8]: ", 8, 1, 30)
	corteMs := lerInteiro(leitor, "  3. Tempo de cada corte (em ms)?             [Padrao: 200]: ", 200, 30, 5000)
	chegadaMs := lerInteiro(leitor, "  4. Intervalo entre chegadas (em ms)?        [Padrao: 40]: ", 40, 10, 5000)
	pausaSeg := lerInteiro(leitor, "  5. Tempo de exibicao de cada passo (em s)?  [Padrao: 2]: ", 2, 1, 10)

	b := NovaBarbearia(cadeiras, time.Duration(corteMs)*time.Millisecond, true)
	go b.Barbeiro()
	time.Sleep(40 * time.Millisecond)

	var wg sync.WaitGroup
	for id := 1; id <= clientes; id++ {
		wg.Add(1)
		go b.Cliente(id, &wg)
		time.Sleep(time.Duration(chegadaMs) * time.Millisecond)
	}

	wg.Wait()
	time.Sleep(30 * time.Millisecond)

	b.mutex.Down()
	b.exibirLogs = false
	b.mutex.Up()

	b.ReproduzirAoVivo(time.Duration(pausaSeg) * time.Second)

	fmt.Println()
	avisoFim := lipgloss.NewStyle().
		Bold(true).
		Foreground(corVerde).
		Border(bordaPadrao).
		BorderForeground(corVerde).
		Padding(0, 2).
		Width(larguraCaixaInt).
		Render("ETAPA 1 CONCLUIDA! Pressione [ENTER] para ver o Historico Completo e o Navegador Passo a Passo...")
	fmt.Println(avisoFim)
	_, _ = leitor.ReadString('\n')

	navegarPeloHistorico(leitor, b, clientes, cadeiras)
}

// exibirTabelaHistoricoCompleto mostra a Etapa 2: todos os eventos numerados do início ao fim.
func exibirTabelaHistoricoCompleto(b *Barbearia, totalClientes int, cadeiras int) {
	limparTela()

	cabecalho := lipgloss.NewStyle().
		Bold(true).
		Foreground(corBranco).
		Background(lipgloss.Color("24")).
		Padding(0, 2).
		Width(larguraTotal).
		Render("ETAPA 2/3: HISTORICO COMPLETO DE EVENTOS DA SIMULACAO")
	fmt.Println(cabecalho)
	fmt.Println()

	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(corTitulo).Render(
		"  PASSO | HORARIO      | THREAD       | STATUS     | O QUE ACONTECEU                | SEMAFOROS ACIONADOS"))
	fmt.Println(lipgloss.NewStyle().Foreground(corCinza).Render(strings.Repeat("-", larguraTotal)))

	for _, ev := range b.historico {
		tag := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("16")).
			Background(ev.CorTag).
			Render(fmt.Sprintf(" %-8s ", ev.TagStatus))

		sem := lipgloss.NewStyle().Foreground(corDestaque).Render(ev.SemaforoOp)
		fmt.Printf("   #%02d  | %s | %-12s | %s | %-30s | %s\n",
			ev.Passo, ev.Horario, ev.Ator, tag, ev.Acao, sem)
	}

	resumo := fmt.Sprintf(
		"RESUMO DA EXECUCAO:   Total de Clientes: %d   |   Atendidos: %d   |   Desistiram (Lotado): %d   |   Pico na Espera: %d/%d",
		totalClientes, b.Atendidos, b.Desistiram, b.MaximoWaiting, cadeiras,
	)
	caixaResumo := lipgloss.NewStyle().
		Bold(true).
		Foreground(corVerde).
		Border(bordaPadrao).
		BorderForeground(corVerde).
		Padding(0, 2).
		Width(larguraCaixaInt).
		Render(resumo)
	fmt.Println()
	fmt.Println(caixaResumo)
}

// navegarPeloHistorico implementa a Etapa 2 (Tabela) + Etapa 3 (Replay Passo a Passo).
func navegarPeloHistorico(leitor *bufio.Reader, b *Barbearia, totalClientes int, cadeiras int) {
	totalPassos := len(b.historico)
	if totalPassos == 0 {
		return
	}

	// Começa exibindo a Etapa 2 (Tabela Completa do Histórico)
	exibirTabelaHistoricoCompleto(b, totalClientes, cadeiras)

	caixaConvite := lipgloss.NewStyle().
		Border(bordaPadrao).
		BorderForeground(corTitulo).
		Padding(0, 2).
		Width(larguraCaixaInt).
		Render(fmt.Sprintf("%s\n  * Digite %s (ou um numero de %s a %s) para abrir o %s\n  * Digite %s para voltar direto ao Menu Principal",
			lipgloss.NewStyle().Bold(true).Foreground(corTitulo).Render("ETAPA 3/3: QUER NAVEGAR PASSO A PASSO PELA BARBEARIA (MODO REPLAY)?"),
			lipgloss.NewStyle().Bold(true).Foreground(corVerde).Render("[ENTER]"),
			lipgloss.NewStyle().Bold(true).Foreground(corAmarelo).Render("1"),
			lipgloss.NewStyle().Bold(true).Foreground(corAmarelo).Render(strconv.Itoa(totalPassos)),
			lipgloss.NewStyle().Bold(true).Foreground(corBranco).Render("Navegador Interativo Passo a Passo"),
			lipgloss.NewStyle().Bold(true).Foreground(corVermelho).Render("[0]"),
		))
	fmt.Println(caixaConvite)
	fmt.Print("\n  Sua escolha ([ENTER] = Iniciar no Passo #01 | [1.." + strconv.Itoa(totalPassos) + "] = Ir p/ Passo | [0] = Menu): ")

	entradaInicial, _ := leitor.ReadString('\n')
	escolha := strings.TrimSpace(strings.ToUpper(entradaInicial))
	if escolha == "0" {
		return
	}

	indiceAtual := 0 // Começa no Passo #01 (índice 0)
	if num, err := strconv.Atoi(escolha); err == nil && num >= 1 && num <= totalPassos {
		indiceAtual = num - 1
	}

	for {
		ev := b.historico[indiceAtual]
		tituloReplay := fmt.Sprintf("ETAPA 3/3: NAVEGADOR DE HISTORICO (REPLAY PASSO A PASSO — PASSO %02d DE %02d)",
			ev.Passo, totalPassos)

		// Desenha o estado congelado da barbearia naquele exato passo
		renderizarSnapshotBarbearia(tituloReplay, ev, totalPassos)

		// Mostra linha do tempo compacta (Passo Anterior | Passo Atual | Próximo Passo)
		antStr := "(Inicio da simulacao)"
		if indiceAtual > 0 {
			ant := b.historico[indiceAtual-1]
			antStr = fmt.Sprintf("#%02d %s: %s", ant.Passo, ant.Ator, ant.Acao)
		}
		proxStr := "(Fim da simulacao)"
		if indiceAtual < totalPassos-1 {
			prox := b.historico[indiceAtual+1]
			proxStr = fmt.Sprintf("#%02d %s: %s", prox.Passo, prox.Ator, prox.Acao)
		}

		linhaDoTempo := fmt.Sprintf("  << Anterior: %-34s  |  Proximo: %-34s >>", antStr, proxStr)
		fmt.Println(lipgloss.NewStyle().Foreground(corCinza).Render(linhaDoTempo))

		// Caixa de Controles de Navegação
		controles := fmt.Sprintf(
			"CONTROLES DO REPLAY:   %s Proximo Passo   |   %s Passo Anterior   |   %s Ir p/ Passo   |   %s Ver Tabela Completa   |   %s Sair p/ Menu",
			lipgloss.NewStyle().Bold(true).Foreground(corVerde).Render("[ENTER / P]"),
			lipgloss.NewStyle().Bold(true).Foreground(corAmarelo).Render("[A]"),
			lipgloss.NewStyle().Bold(true).Foreground(corTitulo).Render("[1.."+strconv.Itoa(totalPassos)+"]"),
			lipgloss.NewStyle().Bold(true).Foreground(corDestaque).Render("[H]"),
			lipgloss.NewStyle().Bold(true).Foreground(corVermelho).Render("[0]"),
		)
		caixaControles := lipgloss.NewStyle().
			Border(bordaPadrao).
			BorderForeground(corTitulo).
			Padding(0, 2).
			Width(larguraCaixaInt).
			Render(controles)
		fmt.Println(caixaControles)
		fmt.Print("  Comando de navegacao: ")

		cmdStr, err := leitor.ReadString('\n')
		if err != nil {
			return
		}
		cmd := strings.TrimSpace(strings.ToUpper(cmdStr))

		switch cmd {
		case "", "P", "N":
			if indiceAtual < totalPassos-1 {
				indiceAtual++
			}
		case "A", "V":
			if indiceAtual > 0 {
				indiceAtual--
			}
		case "H", "T":
			exibirTabelaHistoricoCompleto(b, totalClientes, cadeiras)
			fmt.Print("\n  Pressione [ENTER] para voltar ao Passo #" + fmt.Sprintf("%02d", indiceAtual+1) + " do Replay (ou digite 0 para sair): ")
			voltar, _ := leitor.ReadString('\n')
			if strings.TrimSpace(voltar) == "0" {
				return
			}
		case "0", "S":
			return
		default:
			if num, errConv := strconv.Atoi(cmd); errConv == nil && num >= 1 && num <= totalPassos {
				indiceAtual = num - 1
			}
		}
	}
}

// ============================================================================
// SUÍTE DE TESTES DE VALIDAÇÃO AUTOMATIZADA (Opção 3 do Menu)
// ============================================================================

func ExecutarTestesValidacao(leitor *bufio.Reader) {
	limparTela()
	cabecalho := lipgloss.NewStyle().
		Bold(true).
		Foreground(corBranco).
		Background(lipgloss.Color("24")).
		Padding(0, 2).
		Width(larguraTotal).
		Render("SUITE DE VALIDACAO AUTOMATIZADA DO MECANISMO DE SINCRONIZACAO")
	fmt.Println(cabecalho)
	fmt.Println()

	t1 := testeDormirAcordar()
	t2 := testeLimiteCadeiras()
	t3 := testeEstresseSemDeadlock()

	var msgFinal string
	var corFinal lipgloss.Color
	if t1 && t2 && t3 {
		corFinal = corVerde
		msgFinal = "RESULTADO FINAL: APROVADO!\n" +
			"  * Exclusao Mutua (mutex)     : 0 violacoes de regiao critica\n" +
			"  * Limite do Buffer (waiting) : 0 <= waiting <= N respeitado em 100% dos casos\n" +
			"  * Prevencao de Deadlock      : Todas as threads finalizaram sem travamento"
	} else {
		corFinal = corVermelho
		msgFinal = "RESULTADO FINAL: FALHA EM ALGUM TESTE DE VALIDACAO"
	}

	caixaFinal := lipgloss.NewStyle().
		Bold(true).
		Foreground(corFinal).
		Border(bordaPadrao).
		BorderForeground(corFinal).
		Padding(0, 2).
		Width(larguraCaixaInt).
		Render(msgFinal)

	fmt.Println(caixaFinal)
	aguardarEnter(leitor)
}

func renderizarCardTeste(titulo, cenario, resultado string, passou bool) {
	cor := corVerde
	status := " PASSOU "
	if !passou {
		cor = corVermelho
		status = " FALHOU "
	}
	badge := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("16")).Background(cor).Render(status)
	card := lipgloss.NewStyle().
		Border(bordaPadrao).
		BorderForeground(cor).
		Padding(0, 2).
		Width(larguraCaixaInt).
		Render(fmt.Sprintf("%s  %s\n  * Cenario  : %s\n  * Resultado: %s",
			badge,
			lipgloss.NewStyle().Bold(true).Foreground(corBranco).Render(titulo),
			cenario,
			lipgloss.NewStyle().Foreground(cor).Render(resultado)))
	fmt.Println(card)
}

func testeDormirAcordar() bool {
	const cadeiras, total = 3, 5
	b := NovaBarbearia(cadeiras, 25*time.Millisecond, false)
	go b.Barbeiro()

	var wg sync.WaitGroup
	for id := 1; id <= total; id++ {
		wg.Add(1)
		go b.Cliente(id, &wg)
		time.Sleep(55 * time.Millisecond)
	}
	wg.Wait()

	passou := (b.Atendidos == total) && (b.Desistiram == 0) && (b.Violacoes == 0) && (b.waiting == 0)
	renderizarCardTeste(
		"TESTE 1: Sincronizacao Dormir e Acordar (customers.Down / customers.Up)",
		"5 clientes chegam espacados; o barbeiro dorme entre eles e deve atender todos.",
		fmt.Sprintf("Atendidos = %d/%d | Desistencias = %d | Violacoes de Regiao Critica = %d",
			b.Atendidos, total, b.Desistiram, b.Violacoes),
		passou,
	)
	return passou
}

func testeLimiteCadeiras() bool {
	const cadeiras, total = 4, 12
	b := NovaBarbearia(cadeiras, 100*time.Millisecond, false)
	go b.Barbeiro()
	time.Sleep(20 * time.Millisecond)

	var wg sync.WaitGroup
	for id := 1; id <= total; id++ {
		wg.Add(1)
		go b.Cliente(id, &wg)
	}
	wg.Wait()

	passou := (b.Atendidos+b.Desistiram == total) && (b.Desistiram > 0) &&
		(b.MaximoWaiting <= cadeiras) && (b.Violacoes == 0) && (b.waiting == 0)
	renderizarCardTeste(
		"TESTE 2: Controle de Lotacao das Cadeiras (waiting <= N)",
		"12 clientes chegam simultaneamente disputando N=4 cadeiras de espera.",
		fmt.Sprintf("Atendidos = %d | Desistiram (Lotado) = %d | Pico na Espera = %d/%d | Violacoes = %d",
			b.Atendidos, b.Desistiram, b.MaximoWaiting, cadeiras, b.Violacoes),
		passou,
	)
	return passou
}

func testeEstresseSemDeadlock() bool {
	const cadeiras, total = 5, 50
	b := NovaBarbearia(cadeiras, 10*time.Millisecond, false)
	go b.Barbeiro()

	var wg sync.WaitGroup
	concluido := make(chan struct{})

	go func() {
		for id := 1; id <= total; id++ {
			wg.Add(1)
			go b.Cliente(id, &wg)
			if id%7 == 0 {
				time.Sleep(15 * time.Millisecond)
			}
		}
		wg.Wait()
		close(concluido)
	}()

	select {
	case <-concluido:
		passou := (b.Atendidos+b.Desistiram == total) && (b.Violacoes == 0) &&
			(b.MaximoWaiting <= cadeiras) && (b.waiting == 0)
		renderizarCardTeste(
			"TESTE 3: Estresse com 50 Threads Concorrentes (Anti-Deadlock e Race Condition)",
			"50 threads clientes concorrem simultaneamente com verificacao de timeout.",
			fmt.Sprintf("Total Processado = %d/%d (Atendidos: %d + Desistiram: %d) | Deadlock = Nao",
				b.Atendidos+b.Desistiram, total, b.Atendidos, b.Desistiram),
			passou,
		)
		return passou
	case <-time.After(5 * time.Second):
		renderizarCardTeste(
			"TESTE 3: Estresse com 50 Threads Concorrentes",
			"50 threads clientes concorrem simultaneamente.",
			"DEADLOCK DETECTADO (Timeout excedido!)",
			false,
		)
		return false
	}
}
