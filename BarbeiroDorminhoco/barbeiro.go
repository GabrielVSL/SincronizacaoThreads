package main

import (
	"fmt"
	"time"
)

// ============================================================================
// 3. barbeiro.go — Linha de Execução da Thread do Barbeiro
// ============================================================================
// Implementa exatamente o método 'barber' descrito por Tanenbaum (pág. 6 do PDF):
//   1. Faz Down() em 'customers' (dorme se for 0);
//   2. Faz Down() em 'mutex' (entra na Região Crítica);
//   3. Decrementa 'waiting--' (retira 1 cliente da espera);
//   4. Faz Up() em 'barbers' (avisa que o barbeiro está pronto);
//   5. Faz Up() em 'mutex' (sai da Região Crítica);
//   6. Executa o corte de cabelo (fora da Região Crítica).

func (b *Barbearia) Barbeiro() {
	b.mutex.Down()
	b.log("[BARBEIRO | ABRIU   ]",
		fmt.Sprintf("Abriu a barbearia (%d cadeiras de espera)", b.totalCadeiras),
		b.EstadoCadeiras())
	b.mutex.Up()

	for {
		// Se não há clientes esperando, avisa no log que o barbeiro vai cair no sono
		b.mutex.Down()
		if b.waiting == 0 {
			b.log("[BARBEIRO | DORMINDO]", "Sem clientes! Sentou na cadeira e dormiu...", b.EstadoCadeiras())
		}
		b.mutex.Up()

		// 1. Bloqueia sobre o semáforo 'customers' (dorme se customers == 0)
		b.customers.Down()

		// 2. Entra na REGIÃO CRÍTICA para atualizar 'waiting' com exclusão mútua
		b.mutex.Down()

		clienteAtual := b.filaEspera[0]
		b.filaEspera = b.filaEspera[1:]
		b.waiting--
		b.Atendidos++
		b.validarEstado()

		// 3. Sinaliza que o barbeiro está pronto para cortar (barbers.Up)
		b.barbers.Up()

		b.log("[BARBEIRO | CORTANDO]",
			fmt.Sprintf("Acordou/chamou Cliente #%02d para cortar!", clienteAtual),
			b.EstadoCadeiras())

		// 4. Sai da REGIÃO CRÍTICA antes de cortar o cabelo (evita deadlock!)
		b.mutex.Up()

		// 5. Corta o cabelo fora da Região Crítica
		b.cortarCabelo(clienteAtual)
	}
}

// cortarCabelo simula o tempo de corte de cabelo fora da Região Crítica.
func (b *Barbearia) cortarCabelo(idCliente int) {
	time.Sleep(b.tempoCorte)

	b.mutex.Down()
	b.log("[BARBEIRO | PRONTO  ]",
		fmt.Sprintf("Terminou o corte do Cliente #%02d", idCliente),
		b.EstadoCadeiras())
	b.mutex.Up()

	// Pequena pausa para o cliente levantar da cadeira antes de chamar o próximo
	time.Sleep(15 * time.Millisecond)
}
