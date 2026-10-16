package main

import (
	"fmt"
	"strings"
	"time"
)

// ============================================================================
// 2. barbearia.go — Modelagem dos Recursos Compartilhados e dos 3 Semáforos
// ============================================================================
// Segue fielmente a solução de Tanenbaum (Figura 1.20 - Página 6 do PDF):
//   - Variável compartilhada: 'waiting' (clientes sentados esperando)
//   - 3 Semáforos:
//       1. customers (inicia em 0): conta clientes esperando atendimento.
//       2. barbers   (inicia em 0): número de barbeiros (0 ou 1) prontos.
//       3. mutex     (inicia em 1): exclusão mútua para proteger 'waiting'.

type Barbearia struct {
	// Recursos compartilhados
	waiting       int   // Clientes esperando nas cadeiras (0 <= waiting <= totalCadeiras)
	totalCadeiras int   // Número total (N) de cadeiras de espera na barbearia
	filaEspera    []int // Guarda os IDs dos clientes sentados apenas para mostrar no terminal

	// Os 3 Semáforos obrigatórios do PDF (implementados na unha em semaforo.go)
	customers *Semaforo
	barbers   *Semaforo
	mutex     *Semaforo

	// Configuração de tempo e exibição
	tempoCorte time.Duration
	exibirLogs bool

	// Contadores de validação (sempre atualizados dentro da região crítica do 'mutex')
	Atendidos     int
	Desistiram    int
	MaximoWaiting int
	Violacoes     int
}

// NovaBarbearia cria a barbearia com N cadeiras e inicializa os 3 semáforos do PDF.
func NovaBarbearia(numCadeiras int, tempoCorte time.Duration, exibirLogs bool) *Barbearia {
	return &Barbearia{
		waiting:       0,
		totalCadeiras: numCadeiras,
		filaEspera:    make([]int, 0, numCadeiras),

		// Valores iniciais definidos na página 6 do PDF:
		customers: NovoSemaforo(0), // Inicia em 0 (nenhum cliente esperando)
		barbers:   NovoSemaforo(0), // Inicia em 0 (barbeiro ainda não chamou ninguém)
		mutex:     NovoSemaforo(1), // Inicia em 1 (região crítica livre)

		tempoCorte: tempoCorte,
		exibirLogs: exibirLogs,
	}
}

// validarEstado verifica dentro da Região Crítica se 0 <= waiting <= totalCadeiras.
func (b *Barbearia) validarEstado() {
	if b.waiting < 0 || b.waiting > b.totalCadeiras {
		b.Violacoes++
	}
	if b.waiting > b.MaximoWaiting {
		b.MaximoWaiting = b.waiting
	}
}

// EstadoCadeiras desenha as N cadeiras de espera em ASCII puro (ex: [ #01 ][ #02 ][ livre ]).
// Sempre chamado dentro da Região Crítica (protegido pelo mutex).
func (b *Barbearia) EstadoCadeiras() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Espera (%d/%d): ", b.waiting, b.totalCadeiras))
	for i := 0; i < b.totalCadeiras; i++ {
		if i < len(b.filaEspera) {
			sb.WriteString(fmt.Sprintf("[ #%02d ]", b.filaEspera[i]))
		} else {
			sb.WriteString("[ livre ]")
		}
	}
	return sb.String()
}

// log imprime os eventos das threads alinhados no terminal quando exibirLogs == true.
func (b *Barbearia) log(ator string, acao string, estadoSala string) {
	if !b.exibirLogs {
		return
	}
	horario := time.Now().Format("15:04:05.000")
	if estadoSala != "" {
		fmt.Printf("[%s] %-22s | %-44s | %s\n", horario, ator, acao, estadoSala)
	} else {
		fmt.Printf("[%s] %-22s | %s\n", horario, ator, acao)
	}
}
