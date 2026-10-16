# Roteiro de Explicação para o Vídeo — Problema do Barbeiro Dorminhoco (Go)

Este guia mapeia diretamente **todos os tópicos obrigatórios do edital (página 5 do PDF)** para os **6 arquivos enxutos** que criamos em Go.

---

## Estrutura Enxuta dos 6 Arquivos

| Arquivo | Paralelo com a Parte I | O que contém (e quem pode explicar no vídeo) |
| :--- | :--- | :--- |
| **`semaforo.go`** | Primitiva do SO | Implementação na unha do Semáforo (`NovoSemaforo`, `Down` e `Up`) sem *busy wait*. |
| **`barbearia.go`** | `Buffer.java` | Variável compartilhada `waiting` e os **3 semáforos do PDF**: `customers` (0), `barbers` (0) e `mutex` (1). |
| **`barbeiro.go`** | `Consumidor.java` | Linha de execução da Thread `Barbeiro()` e `cortarCabelo()`, idêntica à página 6 do PDF. |
| **`cliente.go`** | `Produtor.java` | Linha de execução da Thread `Cliente()` e `receberCorte()`, idêntica à página 6 do PDF. |
| **`validacao.go`** | Testes | Simulações visuais e os 3 testes automatizados contra *Race Condition* e *Deadlock*. |
| **`main.go`** | `ProdCons.java` | Menu interativo com o resumo da história e opções `[1]`, `[2]`, `[3]` e `[0]`. |

---

## Divisão dos 6 Tópicos Obrigatórios do Vídeo (Página 5 do PDF)

1. **Modelagem do Problema e Semáforo na Unha (`semaforo.go`):**
   - Mostrar que implementamos o semáforo do zero com `Down()` (bloqueia em `cond.Wait()` enquanto `valor == 0` e depois decrementa `valor--`) e `Up()` (incrementa `valor++` e acorda uma thread com `cond.Signal()`).
2. **Recursos Compartilhados e os 3 Semáforos (`barbearia.go`):**
   - Mostrar a struct `Barbearia` com a variável compartilhada `waiting` (que existe porque não se lê o valor interno de um semáforo) e os 3 semáforos de Tanenbaum:
     - `customers = 0` (conta clientes esperando);
     - `barbers = 0` (indica barbeiro pronto);
     - `mutex = 1` (exclusão mútua).
3. **Identificação das Threads e Linhas de Execução (`barbeiro.go` e `cliente.go`):**
   - **Thread do Barbeiro (`barbeiro.go`):** Laço infinito (`for`) onde dorme em `customers.Down()`, entra na região crítica (`mutex.Down()`), faz `waiting--`, acorda o cliente (`barbers.Up()`), libera `mutex.Up()` e corta o cabelo.
   - **Thread do Cliente (`cliente.go`):** Não tem laço (quer apenas 1 corte). Entra em `mutex.Down()`, verifica `if waiting < totalCadeiras`: se houver vaga, faz `waiting++`, acorda o barbeiro (`customers.Up()`), libera `mutex.Up()` e espera em `barbers.Down()`; se estiver lotado, libera `mutex.Up()` e vai embora.
4. **Regiões Críticas e Estratégia para Evitar Deadlocks (`barbeiro.go` e `cliente.go`):**
   - A Região Crítica é todo acesso à variável `waiting`, protegido por `mutex.Down()` e `mutex.Up()`.
   - **Como evitamos Deadlock:** Nenhuma thread dorme segurando o `mutex`! O cliente sempre libera `mutex.Up()` **antes** de esperar em `barbers.Down()`, e o barbeiro sempre libera `mutex.Up()` **antes** de executar `cortarCabelo()`.
5. **Execução e Validação por Meio de Testes (`main.go` e `validacao.go`):**
   - Rodar `go run .` no terminal, executar a **Opção `[1]`** (mostrando o barbeiro dormindo, acordando, a sala lotando em `3/3`, clientes desistindo e o barbeiro voltando a dormir) e em seguida a **Opção `[3]`** (mostrando os 3 testes de validação aprovados).
