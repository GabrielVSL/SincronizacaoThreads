package main

import (
	"bufio"
	"fmt"
	"strings"
	"sync"
	"time"
)

// ============================================================================
// 5. validacao.go — Simulações e Testes de Validação da Sincronização
// ============================================================================
// Atende ao requisito do PDF (página 4):
// "O funcionamento do programa deverá ser validado por meio de testes,
// demonstrando que a solução apresenta o comportamento esperado e que o
// mecanismo de sincronização está funcionando corretamente."

// ExecutarSimulacaoPadrao demonstra todos os cenários do problema passo a passo:
// barbeiro dormindo, acordando, fila enchendo, clientes indo embora e voltando a dormir.
func ExecutarSimulacaoPadrao(leitor *bufio.Reader) {
	fmt.Println()
	fmt.Println(strings.Repeat("=", 95))
	fmt.Println("  SIMULACAO VISUAL PADRAO (N = 3 cadeiras de espera | Tempo de corte = 300ms)")
	fmt.Println(strings.Repeat("=", 95))
	fmt.Println()

	b := NovaBarbearia(3, 300*time.Millisecond, true)
	go b.Barbeiro()
	time.Sleep(80 * time.Millisecond)

	var wg sync.WaitGroup

	// Onda 1: Chegam 6 clientes rapidamente (1 corta, 3 sentam nas cadeiras e 2 desistem por lotação)
	fmt.Println(">>> [ONDA 1] Chegam 6 clientes em sequencia (Clientes #01 a #06):")
	for id := 1; id <= 6; id++ {
		wg.Add(1)
		go b.Cliente(id, &wg)
		time.Sleep(45 * time.Millisecond)
	}

	wg.Wait()
	time.Sleep(150 * time.Millisecond)

	// Onda 2: Barbeiro já voltou a dormir; chegam mais 2 clientes que o acordam novamente
	fmt.Println("\n>>> [ONDA 2] Barbeiro esta dormindo; chegam mais 2 clientes (#07 e #08):")
	for id := 7; id <= 8; id++ {
		wg.Add(1)
		go b.Cliente(id, &wg)
		time.Sleep(70 * time.Millisecond)
	}

	wg.Wait()
	time.Sleep(30 * time.Millisecond)

	// Desativa os logs desta instância ao finalizar a simulação
	b.mutex.Down()
	b.exibirLogs = false
	b.mutex.Up()

	fmt.Println("\n" + strings.Repeat("-", 95))
	fmt.Printf("  RESUMO: Total = 8 clientes | Atendidos = %d | Desistiram (lotado) = %d | Max. na Espera = %d/3\n",
		b.Atendidos, b.Desistiram, b.MaximoWaiting)
	fmt.Println(strings.Repeat("-", 95))

	aguardarEnter(leitor)
}

// ExecutarSimulacaoCustomizada permite ao usuário definir N cadeiras, clientes e tempos.
func ExecutarSimulacaoCustomizada(leitor *bufio.Reader) {
	fmt.Println()
	fmt.Println(strings.Repeat("=", 95))
	fmt.Println("  SIMULACAO CUSTOMIZADA (Defina os parametros da barbearia)")
	fmt.Println(strings.Repeat("=", 95))

	cadeiras := lerInteiro(leitor, "  1. Quantas cadeiras de espera (N)?          [Padrao: 4]: ", 4, 1, 20)
	clientes := lerInteiro(leitor, "  2. Quantos clientes vao chegar?             [Padrao: 10]: ", 10, 1, 100)
	corteMs := lerInteiro(leitor, "  3. Tempo de cada corte (em ms)?             [Padrao: 250]: ", 250, 20, 5000)
	chegadaMs := lerInteiro(leitor, "  4. Intervalo entre chegadas (em ms)?        [Padrao: 60]: ", 60, 0, 5000)
	fmt.Println()

	b := NovaBarbearia(cadeiras, time.Duration(corteMs)*time.Millisecond, true)
	go b.Barbeiro()
	time.Sleep(80 * time.Millisecond)

	var wg sync.WaitGroup
	for id := 1; id <= clientes; id++ {
		wg.Add(1)
		go b.Cliente(id, &wg)
		if chegadaMs > 0 {
			time.Sleep(time.Duration(chegadaMs) * time.Millisecond)
		}
	}

	wg.Wait()
	time.Sleep(30 * time.Millisecond)

	b.mutex.Down()
	b.exibirLogs = false
	b.mutex.Up()

	fmt.Println("\n" + strings.Repeat("-", 95))
	fmt.Printf("  RESUMO: Total = %d clientes | Atendidos = %d | Desistiram = %d | Max. na Espera = %d/%d\n",
		clientes, b.Atendidos, b.Desistiram, b.MaximoWaiting, cadeiras)
	fmt.Println(strings.Repeat("-", 95))

	aguardarEnter(leitor)
}

// ExecutarTestesValidacao roda os 3 testes automatizados para comprovar a sincronização.
func ExecutarTestesValidacao(leitor *bufio.Reader) {
	fmt.Println()
	fmt.Println(strings.Repeat("=", 95))
	fmt.Println("  TESTES DE VALIDACAO DO MECANISMO DE SINCRONIZACAO")
	fmt.Println(strings.Repeat("=", 95))
	fmt.Println()

	t1 := testeDormirAcordar()
	fmt.Println()
	t2 := testeLimiteCadeiras()
	fmt.Println()
	t3 := testeEstresseSemDeadlock()

	fmt.Println("\n" + strings.Repeat("=", 95))
	if t1 && t2 && t3 {
		fmt.Println("  [RESULTADO FINAL: APROVADO] Todos os 3 testes passaram sem Race Condition e sem Deadlock!")
	} else {
		fmt.Println("  [RESULTADO FINAL: FALHOU] Algum teste apresentou falha.")
	}
	fmt.Println(strings.Repeat("=", 95))

	aguardarEnter(leitor)
}

// Teste 1: Chegadas espaçadas (Barbeiro dorme e acorda a cada cliente, 0 desistências).
func testeDormirAcordar() bool {
	fmt.Println("  [TESTE 1] Fluxo Espacado (Barbeiro dorme e acorda sem fila)")
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
	fmt.Printf("    -> Resultado: [PASSOU = %v] Atendidos: %d/%d | Desistiram: %d | Violacoes: %d\n",
		passou, b.Atendidos, total, b.Desistiram, b.Violacoes)
	return passou
}

// Teste 2: Rajada simultânea maior que N cadeiras (Valida waiting <= N e desistências).
func testeLimiteCadeiras() bool {
	fmt.Println("  [TESTE 2] Lotacao Maxima das Cadeiras (waiting <= N)")
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
	fmt.Printf("    -> Resultado: [PASSOU = %v] Atendidos: %d | Desistiram: %d | Max. na Espera: %d/%d | Violacoes: %d\n",
		passou, b.Atendidos, b.Desistiram, b.MaximoWaiting, cadeiras, b.Violacoes)
	return passou
}

// Teste 3: Estresse com 50 threads concorrentes e verificação de Timeout contra Deadlock.
func testeEstresseSemDeadlock() bool {
	fmt.Println("  [TESTE 3] Estresse com 50 Threads Concorrentes (Anti-Deadlock e Race Condition)")
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
		fmt.Printf("    -> Resultado: [PASSOU = %v] Total: %d (Atendidos: %d + Desistiram: %d) | Sem Deadlock!\n",
			passou, total, b.Atendidos, b.Desistiram)
		return passou
	case <-time.After(5 * time.Second):
		fmt.Println("    -> Resultado: [FALHOU] Deadlock detectado (timeout excedido)!")
		return false
	}
}
