package main

import "sync"

// ============================================================================
// 1. semaforo.go — Implementação do Semáforo Clássico "Na Unha" (Dijkstra)
// ============================================================================
// Conforme exigido, não utilizamos bibliotecas prontas de semáforo.
// O semáforo possui apenas as duas operações atômicas clássicas:
//   - Down() (P / Wait): Decrementa o contador ou bloqueia a thread se for 0.
//   - Up()   (V / Signal): Incrementa o contador e acorda uma thread bloqueada.
// Nota do PDF (pág. 6): Não há método para ler o valor atual do semáforo!

type Semaforo struct {
	valor int
	trava sync.Mutex
	cond  *sync.Cond
}

// NovoSemaforo inicializa um semáforo com o valor inicial definido:
//   - customers = 0 (clientes esperando)
//   - barbers   = 0 (barbeiro disponível)
//   - mutex     = 1 (exclusão mútua da região crítica)
func NovoSemaforo(valorInicial int) *Semaforo {
	s := &Semaforo{valor: valorInicial}
	s.cond = sync.NewCond(&s.trava)
	return s
}

// Down bloqueia a thread na fila de espera enquanto valor == 0 (sem busy wait)
// e decrementa o valor quando houver permissão disponível.
func (s *Semaforo) Down() {
	s.trava.Lock()
	defer s.trava.Unlock()

	for s.valor == 0 {
		s.cond.Wait()
	}
	s.valor--
}

// Up incrementa o valor do semáforo e acorda uma thread que estava dormindo no Down().
func (s *Semaforo) Up() {
	s.trava.Lock()
	defer s.trava.Unlock()

	s.valor++
	s.cond.Signal()
}
