package ProdutorConsumidor;

public class Consumidor extends Thread {
    private Buffer buf;
    private int numero;

    public Consumidor(Buffer buf, int numero) {
        this.buf = buf;
        this.numero = numero;
    }

    public void run() { // retira 10 valores do buffer
        int x;
        for (int i = 0; i < 5; i++) {
            x = buf.get();
            System.out.println("Consumidor #" + numero + " get: " + x);
        }
    }
}
