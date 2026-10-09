import numpy as np
import wave
import struct
import os

class AudioSynthesizer:
    """
    Componente responsável pela síntese de áudio procedural (DSP).
    Gera as ondas sonoras matematicamente para simular sons de bateria,
    evitando a necessidade de arquivos externos (assets).
    """

    def __init__(self, output_dir: str = "assets", sample_rate: int = 44100):
        self.output_dir = output_dir
        self.sample_rate = sample_rate
        if not os.path.exists(self.output_dir):
            os.makedirs(self.output_dir)

    def _save_wav(self, filename: str, audio_data: np.ndarray) -> str:
        """Salva um array numpy de áudio em um arquivo .wav."""
        filepath = os.path.join(self.output_dir, filename)
        
        # Garante que os dados estão no formato correto de 16-bit PCM
        audio_data = np.int16(audio_data * 32767)
        
        with wave.open(filepath, 'w') as wav_file:
            wav_file.setnchannels(1)  # Mono
            wav_file.setsampwidth(2)  # 2 bytes = 16 bit
            wav_file.setframerate(self.sample_rate)
            for sample in audio_data:
                wav_file.writeframesraw(struct.pack('<h', sample))
                
        return filepath

    def generate_note(self, filename: str, frequency: float, duration: float = 0.3) -> str:
        """Gera uma nota musical estilo Sintetizador Retro com harmônicos."""
        t = np.linspace(0, duration, int(self.sample_rate * duration), False)
        
        # Onda senoidal base + harmônico para dar corpo ao som (estilo órgão/synth)
        wave_data = np.sin(2 * np.pi * frequency * t) + 0.3 * np.sin(2 * np.pi * (frequency * 2) * t)
        
        # Envelope ADSR simplificado (Attack rápido, Release suave)
        attack_len = int(len(t) * 0.05)
        release_len = int(len(t) * 0.4)
        envelope = np.ones_like(t)
        envelope[:attack_len] = np.linspace(0, 1, attack_len)
        envelope[-release_len:] = np.linspace(1, 0, release_len)
        
        audio = wave_data * envelope
        
        # Normalização de volume para não estourar o som
        audio = audio / np.max(np.abs(audio))
        # Reduz o volume geral em 30% para ficar agradável
        audio = audio * 0.7 
        
        return self._save_wav(filename, audio)

    def generate_all(self):
        """Gera o acorde de Dó Maior (C - E - G) para os três robôs."""
        print("[AudioSynthesizer] Sintetizando notas musicais...")
        
        # Frequências reais das notas musicais
        path_bart = self.generate_note("bart.wav", 261.63)   # Dó (C4)
        path_lisa = self.generate_note("lisa.wav", 329.63)   # Mi (E4)
        path_maggie = self.generate_note("maggie.wav", 392.00) # Sol (G4)
        
        print(f"[AudioSynthesizer] Arquivos gerados com sucesso na pasta '{self.output_dir}'.")
        return path_bart, path_lisa, path_maggie

if __name__ == "__main__":
    # Teste isolado do componente
    synth = AudioSynthesizer()
    synth.generate_all()
