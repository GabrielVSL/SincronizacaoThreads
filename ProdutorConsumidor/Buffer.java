package ProdutorConsumidor;

public class Buffer {
    private int valor;
    private boolean cheio = false;

    public synchronized int get() {
        while (!cheio) {
            try {
                wait();
            } catch (InterruptedException e) {
            }
        }
        cheio = false;
        notifyAll();
        System.out.println("get " + valor);
        return valor;
    }

    public synchronized void put(int valor) {
        while (cheio) {
            try {
                wait();
            } catch (InterruptedException e) {
            }
        }
        this.valor = valor;
        cheio = true;
        System.out.println("put " + valor);
        notifyAll();
    }
}
