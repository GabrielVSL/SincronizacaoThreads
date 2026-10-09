import os
import sys
import pygame
from audio_synth import AudioSynthesizer
from simulation import RobotSimulation, ConcertSimulation

# Definições de Cores e Estados
COLOR_BG = (15, 15, 20)
COLOR_TEXT = (220, 220, 220)
COLOR_BUTTON_ON = (0, 200, 100)
COLOR_BUTTON_OFF = (200, 50, 50)
COLOR_BART_ACTIVE = (255, 50, 50)
COLOR_LISA_ACTIVE = (50, 255, 50)
COLOR_MAGGIE_ACTIVE = (50, 150, 255)

STATE_MENU = 0
STATE_EXPERIMENT = 1
STATE_MELODY = 2

class VisualEngine:
    """Motor Gráfico 2D utilizando Pygame para visualização do estado das Threads."""
    
    def __init__(self):
        pygame.init()
        pygame.mixer.init(frequency=44100, size=-16, channels=1, buffer=512)
        
        self.width, self.height = 800, 600
        self.screen = pygame.display.set_mode((self.width, self.height))
        pygame.display.set_caption("Concorrência: Os Três Robôs")
        
        self.clock = pygame.time.Clock()
        self.font = pygame.font.SysFont("consolas", 22, bold=True)
        self.title_font = pygame.font.SysFont("trebuchetms", 38, bold=True)
        self.small_font = pygame.font.SysFont("consolas", 14)
        
        self.state = STATE_MENU
        
        # Geometria dos Robôs na tela
        self.robots_ui = {
            "bart":   {"rect": pygame.Rect(100, 220, 140, 180), "name": "BART"},
            "lisa":   {"rect": pygame.Rect(330, 220, 140, 180), "name": "LISA"},
            "maggie": {"rect": pygame.Rect(560, 220, 140, 180), "name": "MAGGIE"}
        }
        
        # Botoes de Interface
        self.btn_sync = pygame.Rect(250, 480, 300, 60)
        self.btn_menu_exp = pygame.Rect(200, 250, 400, 60)
        self.btn_menu_mel = pygame.Rect(200, 340, 400, 60)
        self.btn_back = pygame.Rect(20, 20, 120, 40)
        
        # Áudio e Imagens
        self.sounds = {}
        self.images = {}
        self._init_assets()
        
        # Core (Simulação OS Original)
        self.simulation = RobotSimulation(bpm=160)
        self.simulation.set_action_callback(self._on_robot_action)
        
        # Core (Simulação Modo Concerto/Melodia)
        self.concert_sim = ConcertSimulation(bpm=240) # BPM mais alto para a música
        self.concert_sim.set_action_callback(self._on_concert_action)
        
        self.glow_surfaces = {}

    def _init_assets(self):
        """Carrega os arquivos de áudio e as imagens (assets)."""
        # Carrega os sons gerados pelo synth e os extras da escala
        sound_names = ["bart", "lisa", "maggie", "A4", "B4", "C5", "D4", "F4"]
        for s in sound_names:
            path = f"assets/{s}.wav"
            if os.path.exists(path):
                self.sounds[s] = pygame.mixer.Sound(path)
        
        for robot in ["bart", "lisa", "maggie"]:
            self.images[robot] = {}
            for st in ["on", "off"]:
                path = f"assets/{robot}_{st}.png"
                if os.path.exists(path):
                    img = pygame.image.load(path).convert_alpha()
                    self.images[robot][st] = pygame.transform.smoothscale(img, (140, 180))
                else:
                    self.images[robot][st] = None

    def _on_robot_action(self, robot_id: str):
        # Toca som no modo OS normal (toca o nome base do robô)
        if robot_id in self.sounds:
            self.sounds[robot_id].play()

    def _on_concert_action(self, robot_id: str, note: str):
        # Toca a nota correspondente no modo Concerto
        if note in self.sounds:
            self.sounds[note].play()

    def _get_glow_surface(self, color, radius):
        key = (color, radius)
        if key not in self.glow_surfaces:
            surf = pygame.Surface((radius*2, radius*2), pygame.SRCALPHA)
            for r in range(radius, 0, -2):
                alpha = int(80 * (1 - r/radius) ** 2)
                pygame.draw.circle(surf, (*color, alpha), (radius, radius), r)
            self.glow_surfaces[key] = surf
        return self.glow_surfaces[key]
        
    def _draw_bg_grid(self):
        self.screen.fill(COLOR_BG)
        for x in range(0, self.width, 40):
            pygame.draw.line(self.screen, (25, 25, 30), (x, 0), (x, self.height))
        for y in range(0, self.height, 40):
            pygame.draw.line(self.screen, (25, 25, 30), (0, y), (self.width, y))

    def _draw_menu(self):
        self._draw_bg_grid()
        
        title_text = "OS INCREDIBOTS"
        title_surf = self.title_font.render(title_text, True, COLOR_TEXT)
        self.screen.blit(title_surf, (self.width//2 - title_surf.get_width()//2, 100))
        
        subtitle = self.font.render("Sistemas Operacionais - Sincronização", True, (150, 150, 160))
        self.screen.blit(subtitle, (self.width//2 - subtitle.get_width()//2, 150))
        
        for btn, text, color in [
            (self.btn_menu_exp, "MODO EXPERIMENTO (OS)", (40, 100, 200)),
            (self.btn_menu_mel, "MODO CONCERTO (Melodia)", (200, 100, 40))
        ]:
            pygame.draw.rect(self.screen, (0, 0, 0), btn.move(0, 5), border_radius=12)
            pygame.draw.rect(self.screen, color, btn, border_radius=12)
            pygame.draw.rect(self.screen, (255, 255, 255), btn, width=2, border_radius=12)
            
            txt_surf = self.font.render(text, True, (255, 255, 255))
            self.screen.blit(txt_surf, (btn.centerx - txt_surf.get_width()//2, btn.centery - txt_surf.get_height()//2))

    def _draw_robots_state(self, active_robot_id):
        colors = {"bart": COLOR_BART_ACTIVE, "lisa": COLOR_LISA_ACTIVE, "maggie": COLOR_MAGGIE_ACTIVE}
        
        for r_id, data in self.robots_ui.items():
            rect = data["rect"]
            is_active = (active_robot_id == r_id)
            color = colors[r_id]
            
            draw_rect = rect.copy()
            if is_active:
                draw_rect.y -= 15
                glow = self._get_glow_surface(color, 120)
                self.screen.blit(glow, (draw_rect.centerx - 120, draw_rect.centery - 120), special_flags=pygame.BLEND_ALPHA_SDL2)
            
            img_state = "on" if is_active else "off"
            img = self.images[r_id][img_state]
            
            if img:
                self.screen.blit(img, (draw_rect.x, draw_rect.y))
            else:
                pygame.draw.rect(self.screen, (40, 40, 50), draw_rect, border_radius=15)
            
            name_surf = self.font.render(data["name"], True, COLOR_TEXT)
            self.screen.blit(name_surf, (draw_rect.centerx - name_surf.get_width()//2, draw_rect.bottom + 15))

    def _draw_experiment(self):
        self._draw_bg_grid()
        
        pygame.draw.rect(self.screen, (60, 60, 70), self.btn_back, border_radius=5)
        back_surf = self.small_font.render("< VOLTAR", True, (255, 255, 255))
        self.screen.blit(back_surf, (self.btn_back.centerx - back_surf.get_width()//2, self.btn_back.centery - back_surf.get_height()//2))
        
        title_text = "MODO EXPERIMENTO"
        title_shadow = self.title_font.render(title_text, True, (0, 0, 0))
        title_surf = self.title_font.render(title_text, True, COLOR_TEXT)
        self.screen.blit(title_shadow, (self.width//2 - title_shadow.get_width()//2 + 3, 43))
        self.screen.blit(title_surf, (self.width//2 - title_surf.get_width()//2, 40))
        
        with self.simulation.state_lock:
            active = self.simulation.active_robot
        sync = self.simulation.sync_enabled
        
        self._draw_robots_state(active)
            
        btn_color = COLOR_BUTTON_ON if sync else COLOR_BUTTON_OFF
        pygame.draw.rect(self.screen, (0, 0, 0), self.btn_sync.move(0, 5), border_radius=12)
        pygame.draw.rect(self.screen, btn_color, self.btn_sync, border_radius=12)
        pygame.draw.rect(self.screen, (255, 255, 255), self.btn_sync, width=2, border_radius=12)
        
        btn_text = "SYNC LIGADO (Música)" if sync else "SYNC DESLIGADO (Caos)"
        btn_surf = self.font.render(btn_text, True, (255, 255, 255) if sync else (255, 200, 200))
        self.screen.blit(btn_surf, (self.btn_sync.centerx - btn_surf.get_width()//2, self.btn_sync.centery - btn_surf.get_height()//2))

    def _draw_melody(self):
        self._draw_bg_grid()
        
        pygame.draw.rect(self.screen, (60, 60, 70), self.btn_back, border_radius=5)
        back_surf = self.small_font.render("< VOLTAR", True, (255, 255, 255))
        self.screen.blit(back_surf, (self.btn_back.centerx - back_surf.get_width()//2, self.btn_back.centery - back_surf.get_height()//2))
        
        title_text = "MODO CONCERTO (Orquestra OS)"
        title_shadow = self.title_font.render(title_text, True, (0, 0, 0))
        title_surf = self.title_font.render(title_text, True, (255, 200, 100))
        self.screen.blit(title_shadow, (self.width//2 - title_shadow.get_width()//2 + 3, 43))
        self.screen.blit(title_surf, (self.width//2 - title_surf.get_width()//2, 40))
        
        subtitle = self.small_font.render("Padrão Concorrência: Master/Worker (Maestro e Músicos)", True, (150, 150, 160))
        self.screen.blit(subtitle, (self.width//2 - subtitle.get_width()//2, 85))

        with self.concert_sim.state_lock:
            active = self.concert_sim.active_robot
            
        self._draw_robots_state(active)

    def draw(self):
        if self.state == STATE_MENU:
            self._draw_menu()
        elif self.state == STATE_EXPERIMENT:
            self._draw_experiment()
        elif self.state == STATE_MELODY:
            self._draw_melody()
            
        pygame.display.flip()

    def run(self):
        running = True
        while running:
            for event in pygame.event.get():
                if event.type == pygame.QUIT:
                    running = False
                elif event.type == pygame.MOUSEBUTTONDOWN and event.button == 1:
                    
                    if self.state == STATE_MENU:
                        if self.btn_menu_exp.collidepoint(event.pos):
                            self.state = STATE_EXPERIMENT
                            self.simulation.start()
                        elif self.btn_menu_mel.collidepoint(event.pos):
                            self.state = STATE_MELODY
                            self.concert_sim.start()
                            
                    elif self.state == STATE_EXPERIMENT:
                        if self.btn_back.collidepoint(event.pos):
                            self.simulation.stop()
                            self.state = STATE_MENU
                        elif self.btn_sync.collidepoint(event.pos):
                            self.simulation.toggle_sync()
                            
                    elif self.state == STATE_MELODY:
                        if self.btn_back.collidepoint(event.pos):
                            self.concert_sim.stop()
                            self.state = STATE_MENU
                            
            self.draw()
            self.clock.tick(60)
            
        self.simulation.stop()
        self.concert_sim.stop()
        pygame.quit()
        sys.exit()

if __name__ == "__main__":
    os.chdir(os.path.dirname(os.path.abspath(__file__)))
    app = VisualEngine()
    app.run()
