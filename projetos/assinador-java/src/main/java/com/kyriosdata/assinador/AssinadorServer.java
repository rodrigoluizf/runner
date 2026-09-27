package com.kyriosdata.assinador;

import com.kyriosdata.assinador.domain.SignRequest;
import com.kyriosdata.assinador.domain.SignatureResponse;
import com.kyriosdata.assinador.domain.ValidateRequest;
import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpServer;

import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.net.InetSocketAddress;
import java.nio.charset.StandardCharsets;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicLong;

public class AssinadorServer {

    public static final int DEFAULT_PORT = 8080;

    private final int port;
    private final int timeoutMinutes;
    private final FakeSignatureService service = new FakeSignatureService();
    private HttpServer server;
    private ScheduledExecutorService scheduler;
    private final AtomicLong lastActivity = new AtomicLong(System.currentTimeMillis());

    public AssinadorServer(int port, int timeoutMinutes) {
        this.port = port;
        this.timeoutMinutes = timeoutMinutes;
    }

    public void start() throws IOException {
        server = HttpServer.create(new InetSocketAddress(port), 0);
        server.createContext("/sign", this::handleSign);
        server.createContext("/validate", this::handleValidate);
        server.createContext("/health", this::handleHealth);
        server.createContext("/shutdown", this::handleShutdown);
        server.setExecutor(Executors.newCachedThreadPool());
        server.start();

        long pid = ProcessHandle.current().pid();
        System.out.println("{\"status\":\"started\",\"port\":" + port + ",\"pid\":" + pid + "}");
        System.out.flush();

        if (timeoutMinutes > 0) {
            scheduler = Executors.newSingleThreadScheduledExecutor();
            scheduler.scheduleAtFixedRate(this::checkInactivity, 1, 1, TimeUnit.MINUTES);
        }
    }

    public void stop() {
        if (scheduler != null) scheduler.shutdownNow();
        if (server != null) server.stop(1);
    }

    int getPort() {
        return port;
    }

    private void checkInactivity() {
        long idleMs = System.currentTimeMillis() - lastActivity.get();
        if (idleMs >= (long) timeoutMinutes * 60_000L) {
            System.err.println("Encerrando por inatividade após " + timeoutMinutes + " minuto(s).");
            stop();
        }
    }

    private void handleSign(HttpExchange ex) throws IOException {
        if (!"POST".equalsIgnoreCase(ex.getRequestMethod())) {
            sendError(ex, 405, "Método não permitido");
            return;
        }
        lastActivity.set(System.currentTimeMillis());
        String body = readBody(ex);
        String content = extractField(body, "content");
        String token = extractField(body, "token");
        SignatureResponse resp = service.sign(new SignRequest(content, token));
        sendJson(ex, 200, resp.toJson());
    }

    private void handleValidate(HttpExchange ex) throws IOException {
        if (!"POST".equalsIgnoreCase(ex.getRequestMethod())) {
            sendError(ex, 405, "Método não permitido");
            return;
        }
        lastActivity.set(System.currentTimeMillis());
        String body = readBody(ex);
        String content = extractField(body, "content");
        String signature = extractField(body, "signature");
        SignatureResponse resp = service.validate(new ValidateRequest(content, signature));
        sendJson(ex, 200, resp.toJson());
    }

    private void handleHealth(HttpExchange ex) throws IOException {
        long pid = ProcessHandle.current().pid();
        sendJson(ex, 200, "{\"status\":\"ok\",\"pid\":" + pid + ",\"port\":" + port + "}");
    }

    private void handleShutdown(HttpExchange ex) throws IOException {
        if (!"POST".equalsIgnoreCase(ex.getRequestMethod())) {
            sendError(ex, 405, "Método não permitido");
            return;
        }
        sendJson(ex, 200, "{\"status\":\"shutting_down\"}");
        new Thread(() -> {
            try { Thread.sleep(100); } catch (InterruptedException ignored) { Thread.currentThread().interrupt(); }
            stop();
        }).start();
    }

    private static String readBody(HttpExchange ex) throws IOException {
        try (InputStream is = ex.getRequestBody()) {
            return new String(is.readAllBytes(), StandardCharsets.UTF_8);
        }
    }

    private static void sendJson(HttpExchange ex, int status, String json) throws IOException {
        byte[] bytes = json.getBytes(StandardCharsets.UTF_8);
        ex.getResponseHeaders().set("Content-Type", "application/json; charset=utf-8");
        ex.sendResponseHeaders(status, bytes.length);
        try (OutputStream os = ex.getResponseBody()) {
            os.write(bytes);
        }
    }

    private static void sendError(HttpExchange ex, int status, String message) throws IOException {
        sendJson(ex, status, "{\"error\":\"" + message + "\"}");
    }

    static String extractField(String json, String field) {
        String key = "\"" + field + "\"";
        int keyIdx = json.indexOf(key);
        if (keyIdx < 0) return null;

        int colonIdx = json.indexOf(':', keyIdx + key.length());
        if (colonIdx < 0) return null;

        int i = colonIdx + 1;
        while (i < json.length() && Character.isWhitespace(json.charAt(i))) i++;
        if (i >= json.length()) return null;

        char first = json.charAt(i);
        if (first == 'n') return null; // null

        if (first == '"') {
            StringBuilder sb = new StringBuilder();
            i++;
            while (i < json.length()) {
                char c = json.charAt(i);
                if (c == '\\' && i + 1 < json.length()) {
                    char esc = json.charAt(i + 1);
                    switch (esc) {
                        case '"':  sb.append('"');  break;
                        case '\\': sb.append('\\'); break;
                        case 'n':  sb.append('\n'); break;
                        case 'r':  sb.append('\r'); break;
                        case 't':  sb.append('\t'); break;
                        default:   sb.append(esc);  break;
                    }
                    i += 2;
                } else if (c == '"') {
                    break;
                } else {
                    sb.append(c);
                    i++;
                }
            }
            return sb.toString();
        }
        return null;
    }
}
