package main

import (
	"fmt"
	"sync"
	"time"
)

// ============================================================================
// 4. cliente.go — Linha de Execução da Thread do Cliente
// ============================================================================
// Implementa exatamente o método 'customer' descrito por Tanenbaum (pág. 6 do PDF):
//   1. Faz Down() em 'mutex' (entra na Região Crítica);
//   2. Verifica se 'waiting < totalCadeiras':
//      - Se SIM:
//          a) Incrementa 'waiting++';
//          b) Faz Up() em 'customers' (acorda o barbeiro);
//          c) Faz Up() em 'mutex' (libera a Região Crítica ANTES de esperar -> evita Deadlock!);
//          d) Faz Down() em 'barbers' (aguarda o barbeiro chamá-lo);
//          e) Recebe o corte de cabelo e deixa a barbearia.
//      - Se NÃO (lotado):
//          a) Faz Up() em 'mutex' (libera a Região Crítica);
//          b) Sai da barbearia sem cortar o cabelo.

func (b *Barbearia) Cliente(id int, wg *sync.WaitGroup) {
	defer wg.Done()

	// 1. Entra na REGIÃO CRÍTICA obtendo o semáforo 'mutex'
	b.mutex.Down()
	b.validarEstado()

	nomeAtor := fmt.Sprintf("Cliente #%02d", id)

	if b.waiting < b.totalCadeiras {
		// Há cadeira disponível: o cliente senta-se na espera
		b.waiting++
		b.filaEspera = append(b.filaEspera, id)
		b.validarEstado()

		// 2. Faz Up() no semáforo 'customers' (acorda o barbeiro se estiver dormindo)
		b.customers.Up()

		b.registrarEvento(
			nomeAtor,
			"ESPERA",
			corAmarelo,
			fmt.Sprintf("Sentou na espera (%d/%d)", b.waiting, b.totalCadeiras),
			"waiting++ | customers.Up()",
			fmt.Sprintf("%s obteve mutex.Down(), viu vaga (%d/%d), fez waiting++, acordou barbeiro (customers.Up), liberou mutex.Up() e aguarda em barbers.Down().",
				nomeAtor, b.waiting, b.totalCadeiras),
		)

		// 3. Libera a REGIÃO CRÍTICA antes de esperar pelo barbeiro (evita Deadlock!)
		b.mutex.Up()

		// 4. Aguarda o barbeiro chamá-lo (bloqueia em barbers.Down)
		b.barbers.Down()

		// 5. Recebe o corte de cabelo e vai embora
		b.receberCorte(id)
	} else {
		// Barbearia lotada (waiting == totalCadeiras): vai embora sem cortar
		b.Desistiram++

		b.registrarEvento(
			nomeAtor,
			"LOTADO!",
			corVermelho,
			"Sem vagas! Foi embora",
			"mutex.Up() (Desistiu)",
			fmt.Sprintf("%s obteve mutex.Down(), viu que waiting == %d (LOTADO!), liberou mutex.Up() e foi embora sem cortar.",
				nomeAtor, b.totalCadeiras),
		)

		// Libera a REGIÃO CRÍTICA e sai da barbearia
		b.mutex.Up()
	}
}

// receberCorte aguarda o término do corte de cabelo e deixa a barbearia.
func (b *Barbearia) receberCorte(id int) {
	time.Sleep(b.tempoCorte + 5*time.Millisecond)

	b.mutex.Down()
	nomeAtor := fmt.Sprintf("Cliente #%02d", id)
	b.registrarEvento(
		nomeAtor,
		"SAIU",
		corVerde,
		"Cabelo cortado! Saiu feliz",
		"Fim de receberCorte()",
		fmt.Sprintf("%s terminou de receber o corte de cabelo e deixou a barbearia.", nomeAtor),
	)
	b.mutex.Up()
}
