import threading
import time
from typing import Callable, Dict

class RobotSimulation:
    """
    Núcleo de concorrência do problema dos 'Três Robôs'.
    Controla os semáforos, as threads e as regras de sincronização.
    """

    def __init__(self, bpm: int = 120):
        # Configuração de tempo baseada em BPM (Batidas por Minuto)
        self.sleep_time = 60.0 / bpm
        
        # Estado de sincronização (Chaos Mode)
        self.sync_enabled = True
        self.running = False
        
        # Callbacks para notificar a GUI e tocar o som
        self.on_robot_action: Callable[[str], None] = lambda robot_id: None
        
        # Estado atual para o renderizador
        self.active_robot = None
        self.state_lock = threading.Lock()

        # Semáforos (Sincronização Baton Passing)
        # Ordem exigida: Bart -> Lisa -> Maggie -> Lisa -> Bart
        self.sem_bart = threading.Semaphore(1)  # Bart começa
        self.sem_lisa_from_bart = threading.Semaphore(0)
        self.sem_maggie = threading.Semaphore(0)
        self.sem_lisa_from_maggie = threading.Semaphore(0)

        # Referências das threads
        self.threads = []

    def set_action_callback(self, callback: Callable[[str], None]):
        """Define o callback disparado quando um robô executa sua ação."""
        self.on_robot_action = callback

    def _update_active_robot(self, robot_id: str):
        """Atualiza o robô ativo de forma thread-safe."""
        with self.state_lock:
            self.active_robot = robot_id
            
        # Notifica o listener (geralmente a interface gráfica/áudio)
        if self.on_robot_action and robot_id:
            self.on_robot_action(robot_id)

    def _bart_routine(self):
        while self.running:
            is_sync = self.sync_enabled
            if is_sync:
                self.sem_bart.acquire()
            
            if not self.running: break
                
            self._update_active_robot("bart")
            time.sleep(self.sleep_time)
            self._update_active_robot(None) # Ocioso após tocar
            
            if is_sync:
                self.sem_lisa_from_bart.release()

    def _lisa_from_bart_routine(self):
        while self.running:
            is_sync = self.sync_enabled
            if is_sync:
                self.sem_lisa_from_bart.acquire()
                
            if not self.running: break
                
            self._update_active_robot("lisa")
            time.sleep(self.sleep_time)
            self._update_active_robot(None)
            
            if is_sync:
                self.sem_maggie.release()

    def _lisa_from_maggie_routine(self):
        while self.running:
            is_sync = self.sync_enabled
            if is_sync:
                self.sem_lisa_from_maggie.acquire()
                
            if not self.running: break
                
            self._update_active_robot("lisa")
            time.sleep(self.sleep_time)
            self._update_active_robot(None)
            
            if is_sync:
                self.sem_bart.release()

    def _maggie_routine(self):
        while self.running:
            is_sync = self.sync_enabled
            if is_sync:
                self.sem_maggie.acquire()
                
            if not self.running: break
                
            self._update_active_robot("maggie")
            time.sleep(self.sleep_time)
            self._update_active_robot(None)
            
            if is_sync:
                self.sem_lisa_from_maggie.release()

    def start(self):
        """Inicia todas as threads."""
        self.running = True
        self.threads = [
            threading.Thread(target=self._bart_routine, name="Thread-Bart"),
            threading.Thread(target=self._lisa_from_bart_routine, name="Thread-Lisa-Phase1"),
            threading.Thread(target=self._maggie_routine, name="Thread-Maggie"),
            threading.Thread(target=self._lisa_from_maggie_routine, name="Thread-Lisa-Phase2")
        ]
        
        for t in self.threads:
            t.daemon = True # Garante que as threads morrem ao fechar o programa
            t.start()

    def stop(self):
        """Sinaliza para as threads pararem."""
        self.running = False
        # Libera possíveis deadlocks no desligamento
        self.sem_bart.release()
        self.sem_lisa_from_bart.release()
        self.sem_maggie.release()
        self.sem_lisa_from_maggie.release()
        
    def toggle_sync(self):
        """Ativa/desativa a sincronização (Chaos Mode)."""
        self.sync_enabled = not self.sync_enabled
        
        if not self.sync_enabled:
            # Desligou a sincronização: libera todos que estão bloqueados
            # para que as threads possam rodar livremente no modo Caos.
            self.sem_bart.release()
            self.sem_lisa_from_bart.release()
            self.sem_maggie.release()
            self.sem_lisa_from_maggie.release()
        else:
            # Ligou a sincronização novamente: precisamos restaurar o estado inicial perfeito.
            # 1. Drenar todos os semáforos até zero (remover os 'bastões fantasma' do sistema)
            while self.sem_bart.acquire(blocking=False): pass
            while self.sem_lisa_from_bart.acquire(blocking=False): pass
            while self.sem_maggie.acquire(blocking=False): pass
            while self.sem_lisa_from_maggie.acquire(blocking=False): pass
            
            # 2. Entregar um único bastão (token) para o Bart recomeçar o ciclo
            self.sem_bart.release()
            
        return self.sync_enabled


class ConcertSimulation:
    """
    Núcleo de concorrência para o 'Modo Concerto'.
    Usa um padrão Master-Worker (Maestro-Músicos) com semáforos.
    """
    def __init__(self, bpm: int = 160):
        self.sleep_time = 60.0 / bpm
        self.running = False
        self.on_robot_action: Callable[[str, str], None] = lambda robot_id, note: None
        
        self.active_robot = None
        self.state_lock = threading.Lock()
        
        # A partitura (Brilha Brilha Estrelinha) mapeada nos robôs e suas notas
        self.partitura = [
            ("bart", "bart"), ("bart", "bart"), ("maggie", "maggie"), ("maggie", "maggie"), 
            ("lisa", "A4"), ("lisa", "A4"), ("maggie", "maggie"), None,
            
            ("maggie", "F4"), ("maggie", "F4"), ("lisa", "lisa"), ("lisa", "lisa"),
            ("bart", "D4"), ("bart", "D4"), ("bart", "bart"), None,
            
            ("maggie", "maggie"), ("maggie", "maggie"), ("maggie", "F4"), ("maggie", "F4"),
            ("lisa", "lisa"), ("lisa", "lisa"), ("bart", "D4"), None,
            
            ("maggie", "maggie"), ("maggie", "maggie"), ("maggie", "F4"), ("maggie", "F4"),
            ("lisa", "lisa"), ("lisa", "lisa"), ("bart", "D4"), None,
            
            ("bart", "bart"), ("bart", "bart"), ("maggie", "maggie"), ("maggie", "maggie"), 
            ("lisa", "A4"), ("lisa", "A4"), ("maggie", "maggie"), None,
            
            ("maggie", "F4"), ("maggie", "F4"), ("lisa", "lisa"), ("lisa", "lisa"),
            ("bart", "D4"), ("bart", "D4"), ("bart", "bart"), None,
        ]
        
        self.current_note_idx = 0
        self.current_note_name = ""
        
        # Semáforos para comunicação Maestro <-> Músicos
        self.sem_maestro = threading.Semaphore(1)
        self.sem_robots = {
            "bart": threading.Semaphore(0),
            "lisa": threading.Semaphore(0),
            "maggie": threading.Semaphore(0)
        }
        self.threads = []

    def set_action_callback(self, callback: Callable[[str, str], None]):
        self.on_robot_action = callback

    def _update_active_robot(self, robot_id: str, note: str):
        with self.state_lock:
            self.active_robot = robot_id
        if self.on_robot_action and robot_id:
            self.on_robot_action(robot_id, note)

    def _maestro_routine(self):
        """O Maestro lê a partitura e libera o robô correto (Padrão Dispatcher)."""
        while self.running:
            self.sem_maestro.acquire()
            if not self.running: break
            
            item = self.partitura[self.current_note_idx]
            
            if item is None:
                # Pausa
                time.sleep(self.sleep_time)
                self.current_note_idx = (self.current_note_idx + 1) % len(self.partitura)
                self.sem_maestro.release()
            else:
                robot_id, note = item
                self.current_note_name = note
                self.current_note_idx = (self.current_note_idx + 1) % len(self.partitura)
                # Acorda o robô específico
                self.sem_robots[robot_id].release()

    def _robot_routine(self, robot_id: str):
        """Worker genérico para os robôs."""
        while self.running:
            self.sem_robots[robot_id].acquire()
            if not self.running: break
            
            note_to_play = self.current_note_name
            self._update_active_robot(robot_id, note_to_play)
            time.sleep(self.sleep_time)
            self._update_active_robot(None, "")
            
            # Devolve o controle para o maestro
            self.sem_maestro.release()

    def start(self):
        self.running = True
        self.current_note_idx = 0
        
        # Drenar semáforos preventivamente
        while self.sem_maestro.acquire(blocking=False): pass
        for s in self.sem_robots.values():
            while s.acquire(blocking=False): pass
            
        self.sem_maestro.release() # Maestro começa
        
        self.threads = [
            threading.Thread(target=self._maestro_routine, name="Thread-Maestro"),
            threading.Thread(target=self._robot_routine, args=("bart",), name="Thread-Musico-Bart"),
            threading.Thread(target=self._robot_routine, args=("lisa",), name="Thread-Musico-Lisa"),
            threading.Thread(target=self._robot_routine, args=("maggie",), name="Thread-Musico-Maggie")
        ]
        
        for t in self.threads:
            t.daemon = True
            t.start()

    def stop(self):
        self.running = False
        self.sem_maestro.release()
        for s in self.sem_robots.values():
            s.release()
