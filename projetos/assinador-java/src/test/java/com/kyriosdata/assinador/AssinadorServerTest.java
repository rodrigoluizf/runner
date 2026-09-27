package com.kyriosdata.assinador;

import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import java.io.IOException;
import java.io.OutputStream;
import java.net.HttpURLConnection;
import java.net.ServerSocket;
import java.net.URI;
import java.nio.charset.StandardCharsets;

import static org.junit.jupiter.api.Assertions.*;

class AssinadorServerTest {

    private AssinadorServer server;
    private int port;

    @BeforeEach
    void setUp() throws Exception {
        port = freePort();
        server = new AssinadorServer(port, 0);
        server.start();
        waitUntilReady(port);
    }

    @AfterEach
    void tearDown() {
        if (server != null) server.stop();
    }

    @Test
    void healthDeveRetornarOk() throws Exception {
        HttpURLConnection conn = get("/health");
        assertEquals(200, conn.getResponseCode());
        String body = new String(conn.getInputStream().readAllBytes(), StandardCharsets.UTF_8);
        assertTrue(body.contains("\"status\":\"ok\""));
        conn.disconnect();
    }

    @Test
    void signComConteudoValidoDeveRetornarAssinatura() throws Exception {
        HttpURLConnection conn = post("/sign", "{\"content\":\"documento\",\"token\":null}");
        assertEquals(200, conn.getResponseCode());
        String body = new String(conn.getInputStream().readAllBytes(), StandardCharsets.UTF_8);
        assertTrue(body.contains("\"valid\":true"));
        assertTrue(body.contains(FakeSignatureService.FAKE_SIGNATURE));
        conn.disconnect();
    }

    @Test
    void signComConteudoVazioDeveRetornarErro() throws Exception {
        HttpURLConnection conn = post("/sign", "{\"content\":\"\",\"token\":null}");
        assertEquals(200, conn.getResponseCode());
        String body = new String(conn.getInputStream().readAllBytes(), StandardCharsets.UTF_8);
        assertTrue(body.contains("\"valid\":false"));
        conn.disconnect();
    }

    @Test
    void signComConteudoNuloDeveRetornarErro() throws Exception {
        HttpURLConnection conn = post("/sign", "{\"content\":null}");
        assertEquals(200, conn.getResponseCode());
        String body = new String(conn.getInputStream().readAllBytes(), StandardCharsets.UTF_8);
        assertTrue(body.contains("\"valid\":false"));
        conn.disconnect();
    }

    @Test
    void validateComAssinaturaCorretaDeveRetornarValido() throws Exception {
        String sig = FakeSignatureService.FAKE_SIGNATURE;
        HttpURLConnection conn = post("/validate",
            "{\"content\":\"doc\",\"signature\":\"" + sig + "\"}");
        assertEquals(200, conn.getResponseCode());
        String body = new String(conn.getInputStream().readAllBytes(), StandardCharsets.UTF_8);
        assertTrue(body.contains("\"valid\":true"));
        conn.disconnect();
    }

    @Test
    void validateComAssinaturaErradaDeveRetornarInvalido() throws Exception {
        HttpURLConnection conn = post("/validate",
            "{\"content\":\"doc\",\"signature\":\"assinatura-errada\"}");
        assertEquals(200, conn.getResponseCode());
        String body = new String(conn.getInputStream().readAllBytes(), StandardCharsets.UTF_8);
        assertTrue(body.contains("\"valid\":false"));
        conn.disconnect();
    }

    @Test
    void validateComAssinaturaNulaDeveRetornarErro() throws Exception {
        HttpURLConnection conn = post("/validate",
            "{\"content\":\"doc\",\"signature\":null}");
        assertEquals(200, conn.getResponseCode());
        String body = new String(conn.getInputStream().readAllBytes(), StandardCharsets.UTF_8);
        assertTrue(body.contains("\"valid\":false"));
        conn.disconnect();
    }

    @Test
    void metodosInvalidosDevemRetornar405() throws Exception {
        HttpURLConnection conn = get("/sign");
        assertEquals(405, conn.getResponseCode());
        conn.disconnect();
    }

    @Test
    void shutdownDeveEncerrarServidor() throws Exception {
        HttpURLConnection conn = post("/shutdown", "{}");
        assertEquals(200, conn.getResponseCode());
        conn.disconnect();
        Thread.sleep(300);
        assertFalse(isAlive(port), "Servidor deveria ter encerrado após /shutdown");
    }

    // --- helpers ---

    private static int freePort() throws IOException {
        try (ServerSocket s = new ServerSocket(0)) {
            return s.getLocalPort();
        }
    }

    private static void waitUntilReady(int port) throws Exception {
        long deadline = System.currentTimeMillis() + 5_000;
        while (System.currentTimeMillis() < deadline) {
            if (isAlive(port)) return;
            Thread.sleep(50);
        }
        throw new IllegalStateException("Servidor não ficou pronto na porta " + port);
    }

    private static boolean isAlive(int port) {
        try {
            HttpURLConnection conn = (HttpURLConnection) URI.create("http://localhost:" + port + "/health").toURL().openConnection();
            conn.setConnectTimeout(500);
            conn.setReadTimeout(500);
            int code = conn.getResponseCode();
            conn.disconnect();
            return code == 200;
        } catch (Exception e) {
            return false;
        }
    }

    private HttpURLConnection get(String path) throws Exception {
        HttpURLConnection conn = (HttpURLConnection) URI.create("http://localhost:" + port + path).toURL().openConnection();
        conn.setRequestMethod("GET");
        conn.setConnectTimeout(2000);
        conn.setReadTimeout(2000);
        return conn;
    }

    private HttpURLConnection post(String path, String body) throws Exception {
        HttpURLConnection conn = (HttpURLConnection) URI.create("http://localhost:" + port + path).toURL().openConnection();
        conn.setRequestMethod("POST");
        conn.setRequestProperty("Content-Type", "application/json; charset=utf-8");
        conn.setDoOutput(true);
        conn.setConnectTimeout(2000);
        conn.setReadTimeout(2000);
        try (OutputStream os = conn.getOutputStream()) {
            os.write(body.getBytes(StandardCharsets.UTF_8));
        }
        return conn;
    }
}
