package ProdutorConsumidor;

public class ProdCons {
    public static void main(String[] args) {
        Buffer buf = new Buffer();
        int i;
        for (i = 0; i < 1; i++) {
            new Consumidor(buf, i).start();
            new Produtor(buf, i).start();
        }
    }
}
