package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// ============================================================================
// 2. barbearia.go — Recursos Compartilhados, os 3 Semáforos e Visualização
// ============================================================================
// Mantém estritamente a modelagem do PDF (Tanenbaum, pág. 6):
//   - 1 variável compartilhada: 'waiting' (clientes sentados nas N cadeiras)
//   - 3 semáforos feitos na unha: 'customers' (0), 'barbers' (0) e 'mutex' (1)

const (
	larguraTotal       = 114 // Largura total padronizada de todas as telas
	larguraCaixaInt    = 112 // Largura interna de caixas que ocupam a tela toda (112 + 2 de borda = 114)
	larguraMeiaCaixa   = 54  // Largura interna de duas caixas lado a lado: (54+2) + 2 + (54+2) = 114
)

var (
	corTitulo   = lipgloss.Color("39")  // Azul brilhante
	corVerde    = lipgloss.Color("42")  // Verde (Corte / Sucesso)
	corAmarelo  = lipgloss.Color("214") // Amarelo/Laranja (Espera)
	corVermelho = lipgloss.Color("196") // Vermelho (Lotado / Desistiu)
	corAzul     = lipgloss.Color("45")  // Ciano (Dormindo)
	corCinza    = lipgloss.Color("244") // Cinza (Vazio / Borda suave)
	corBranco   = lipgloss.Color("255")
	corDestaque = lipgloss.Color("213") // Rosa/Roxo claro para semáforos

	bordaPadrao = lipgloss.Border{
		Top:         "-",
		Bottom:      "-",
		Left:        "|",
		Right:       "|",
		TopLeft:     "+",
		TopRight:    "+",
		BottomLeft:  "+",
		BottomRight: "+",
	}
)

// EventoHistorico armazena uma "Fotografia Completa" (Snapshot) da barbearia naquele passo,
// permitindo reproduzir e navegar pelo histórico passo a passo depois da simulação.
type EventoHistorico struct {
	Passo          int
	Horario        string
	Ator           string
	TagStatus      string
	CorTag         lipgloss.Color
	Acao           string
	SemaforoOp     string
	Explicacao     string
	// Fotografia do estado da barbearia no exato instante deste passo:
	ClienteEmCorte int
	Waiting        int
	TotalCadeiras  int
	FilaEspera     []int
	Atendidos      int
	Desistiram     int
}

type Barbearia struct {
	// Recursos compartilhados (Modelagem Tanenbaum)
	waiting        int   // Quantidade de clientes nas cadeiras (0 <= waiting <= totalCadeiras)
	totalCadeiras  int   // Capacidade máxima (N) de cadeiras de espera
	filaEspera     []int // IDs dos clientes sentados (para desenhar nas cadeiras)
	clienteEmCorte int   // ID do cliente na cadeira de barbeiro (0 = ninguém)

	// Os 3 Semáforos obrigatórios do PDF (implementados na unha em semaforo.go)
	customers *Semaforo
	barbers   *Semaforo
	mutex     *Semaforo

	// Configurações e histórico de snapshots
	tempoCorte time.Duration
	exibirLogs bool
	historico  []EventoHistorico

	// Contadores de validação (sempre atualizados dentro da região crítica do 'mutex')
	Atendidos     int
	Desistiram    int
	MaximoWaiting int
	Violacoes     int
}

func init() {
	lipgloss.SetColorProfile(termenv.ANSI256)
	if runtime.GOOS == "windows" {
		_ = exec.Command("cmd", "/c", "chcp 65001").Run()
	}
}

func NovaBarbearia(numCadeiras int, tempoCorte time.Duration, exibirLogs bool) *Barbearia {
	return &Barbearia{
		waiting:       0,
		totalCadeiras: numCadeiras,
		filaEspera:    make([]int, 0, numCadeiras),

		// Valores iniciais definidos na página 6 do PDF:
		customers: NovoSemaforo(0),
		barbers:   NovoSemaforo(0),
		mutex:     NovoSemaforo(1),

		tempoCorte: tempoCorte,
		exibirLogs: exibirLogs,
		historico:  make([]EventoHistorico, 0, 32),
	}
}

func (b *Barbearia) validarEstado() {
	if b.waiting < 0 || b.waiting > b.totalCadeiras {
		b.Violacoes++
	}
	if b.waiting > b.MaximoWaiting {
		b.MaximoWaiting = b.waiting
	}
}

// registrarEvento salva atomicamente a fotografia completa deste passo dentro da Região Crítica.
func (b *Barbearia) registrarEvento(ator, tagStatus string, cor lipgloss.Color, acao, semaforoOp, explicacao string) {
	if !b.exibirLogs {
		return
	}

	copiaFila := make([]int, len(b.filaEspera))
	copy(copiaFila, b.filaEspera)

	ev := EventoHistorico{
		Passo:          len(b.historico) + 1,
		Horario:        time.Now().Format("15:04:05.000"),
		Ator:           ator,
		TagStatus:      tagStatus,
		CorTag:         cor,
		Acao:           acao,
		SemaforoOp:     semaforoOp,
		Explicacao:     explicacao,
		ClienteEmCorte: b.clienteEmCorte,
		Waiting:        b.waiting,
		TotalCadeiras:  b.totalCadeiras,
		FilaEspera:     copiaFila,
		Atendidos:      b.Atendidos,
		Desistiram:     b.Desistiram,
	}
	b.historico = append(b.historico, ev)
}

// ReproduzirAoVivo exibe na tela cada passo capturado pelas threads em ritmo pausado (2s por passo)
// para que o usuário consiga ler e acompanhar visualmente toda a execução.
func (b *Barbearia) ReproduzirAoVivo(pausaPorPasso time.Duration) {
	total := len(b.historico)
	for _, ev := range b.historico {
		titulo := fmt.Sprintf(
			"ETAPA 1/3: VISUALIZACAO AO VIVO DA BARBEARIA (Passo %02d de %02d — Ritmo: %.1fs por evento)",
			ev.Passo, total, pausaPorPasso.Seconds(),
		)
		renderizarSnapshotBarbearia(titulo, ev, total)
		time.Sleep(pausaPorPasso)
	}
}

// renderizarSnapshotBarbearia desenha a Planta da Barbearia e a explicação de um passo específico
// (usado tanto na Etapa 1 ao vivo quanto na Etapa 3 no Navegador de Histórico / Replay).
func renderizarSnapshotBarbearia(tituloTopo string, ev EventoHistorico, totalPassos int) {
	limparTela()

	estiloCabecalho := lipgloss.NewStyle().
		Bold(true).
		Foreground(corBranco).
		Background(lipgloss.Color("24")).
		Padding(0, 2).
		Width(larguraTotal)

	fmt.Println(estiloCabecalho.Render(tituloTopo))

	// 1. BLOCO ESQUERDO: CADEIRA DO BARBEIRO (Largura externa = 54 + 2 = 56)
	var statusBarbeiro, detalheBarbeiro string
	var corBarbeiro lipgloss.Color
	if ev.ClienteEmCorte > 0 {
		corBarbeiro = corVerde
		statusBarbeiro = fmt.Sprintf(" CORTANDO CLIENTE #%02d ", ev.ClienteEmCorte)
		detalheBarbeiro = "Barbeiro ocupado no corte (mutex livre)"
	} else {
		corBarbeiro = corAzul
		statusBarbeiro = " DORMINDO NA CADEIRA (zZz) "
		detalheBarbeiro = "Barbeiro bloqueado em customers.Down()"
	}

	badgeBarbeiro := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("16")).
		Background(corBarbeiro).
		Padding(0, 1).
		Render(statusBarbeiro)

	caixaBarbeiro := lipgloss.NewStyle().
		Border(bordaPadrao).
		BorderForeground(corBarbeiro).
		Padding(0, 2).
		Width(larguraMeiaCaixa).
		Height(5).
		Render(fmt.Sprintf("%s\n\n%s\n%s",
			lipgloss.NewStyle().Bold(true).Foreground(corBranco).Render("1. CADEIRA DO BARBEIRO"),
			badgeBarbeiro,
			lipgloss.NewStyle().Foreground(corCinza).Render(detalheBarbeiro)))

	// 2. BLOCO DIREITO: SALA DE ESPERA (Largura externa = 54 + 2 = 56)
	var cadeirasRenderizadas []string
	for i := 0; i < ev.TotalCadeiras; i++ {
		if i < len(ev.FilaEspera) {
			idCli := ev.FilaEspera[i]
			c := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("16")).
				Background(corAmarelo).
				Padding(0, 1).
				Render(fmt.Sprintf("Cad.%d: CLI #%02d", i+1, idCli))
			cadeirasRenderizadas = append(cadeirasRenderizadas, c)
		} else {
			c := lipgloss.NewStyle().
				Foreground(corCinza).
				Padding(0, 1).
				Render(fmt.Sprintf("[Cad.%d: LIVRE]", i+1))
			cadeirasRenderizadas = append(cadeirasRenderizadas, c)
		}
	}

	linhaCadeiras := strings.Join(cadeirasRenderizadas, " ")
	subtituloEspera := fmt.Sprintf("Variavel compartilhada: waiting = %d de %d cadeiras", ev.Waiting, ev.TotalCadeiras)
	if ev.Waiting == ev.TotalCadeiras {
		subtituloEspera += " (LOTADA!)"
	}

	caixaEspera := lipgloss.NewStyle().
		Border(bordaPadrao).
		BorderForeground(corAmarelo).
		Padding(0, 2).
		Width(larguraMeiaCaixa).
		Height(5).
		Render(fmt.Sprintf("%s\n\n%s\n%s",
			lipgloss.NewStyle().Bold(true).Foreground(corBranco).Render("2. SALA DE ESPERA (N CADEIRAS)"),
			linhaCadeiras,
			lipgloss.NewStyle().Foreground(corCinza).Render(subtituloEspera)))

	fmt.Println(lipgloss.JoinHorizontal(lipgloss.Top, caixaBarbeiro, "  ", caixaEspera))

	// 3. CAIXA DO PLACAR E ESTADO ATUAL (Largura externa = 112 + 2 = 114)
	infoPasso := fmt.Sprintf("Passo %02d de %02d", ev.Passo, totalPassos)
	linhaPlacar := fmt.Sprintf("%s   |   %s   |   %s   |   %s",
		lipgloss.NewStyle().Bold(true).Foreground(corTitulo).Render(infoPasso+" ("+ev.Horario+")"),
		lipgloss.NewStyle().Bold(true).Foreground(corVerde).Render(fmt.Sprintf("Atendidos: %d", ev.Atendidos)),
		lipgloss.NewStyle().Bold(true).Foreground(corAmarelo).Render(fmt.Sprintf("Esperando (waiting): %d/%d", ev.Waiting, ev.TotalCadeiras)),
		lipgloss.NewStyle().Bold(true).Foreground(corVermelho).Render(fmt.Sprintf("Foram Embora (Lotado): %d", ev.Desistiram)),
	)

	caixaPlacar := lipgloss.NewStyle().
		Border(bordaPadrao).
		BorderForeground(corCinza).
		Padding(0, 2).
		Width(larguraCaixaInt).
		Render(linhaPlacar)
	fmt.Println(caixaPlacar)

	// 4. QUADRO DIDÁTICO: EXPLICAÇÃO DO EVENTO E DOS SEMÁFOROS (Largura externa = 112 + 2 = 114)
	badgeStatus := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("16")).
		Background(ev.CorTag).
		Padding(0, 1).
		Render(ev.TagStatus)

	linhaAtor := fmt.Sprintf("%s  %s — %s",
		badgeStatus,
		lipgloss.NewStyle().Bold(true).Foreground(corBranco).Render(ev.Ator),
		lipgloss.NewStyle().Bold(true).Foreground(ev.CorTag).Render(ev.Acao),
	)

	linhaSemaforo := fmt.Sprintf("Semaforos / Variaveis acionados: %s",
		lipgloss.NewStyle().Bold(true).Foreground(corDestaque).Render(ev.SemaforoOp))

	linhaExplicacao := fmt.Sprintf("Explicacao: %s",
		lipgloss.NewStyle().Foreground(corBranco).Render(ev.Explicacao))

	caixaExplicacao := lipgloss.NewStyle().
		Border(bordaPadrao).
		BorderForeground(corDestaque).
		Padding(0, 2).
		Width(larguraCaixaInt).
		Render(fmt.Sprintf("%s\n\n%s\n%s", linhaAtor, linhaSemaforo, linhaExplicacao))

	fmt.Println(caixaExplicacao)
}

func limparTela() {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		_ = cmd.Run()
	} else {
		fmt.Print("\033[H\033[2J")
	}
}
