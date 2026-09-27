package com.kyriosdata.assinador;

import org.junit.jupiter.api.*;

import java.io.File;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.*;
import java.security.*;
import java.security.spec.RSAKeyGenParameterSpec;
import java.util.Comparator;

import static org.junit.jupiter.api.Assertions.*;

/**
 * Testes de integração PKCS#11 usando SoftHSM2 como simulador de token criptográfico.
 *
 * Requer softhsm2 instalado e a variável SOFTHSM2_CONF definida apontando para
 * o arquivo de configuração que será criado durante o setup (configurado via
 * Maven surefire no pom.xml). Se SoftHSM2 não estiver disponível, todos os
 * testes são pulados via Assumptions.
 */
class Pkcs11IntegrationTest {

    private static final String SOFTHSM_LIB = "/usr/lib/softhsm/libsofthsm2.so";
    private static final String PIN = "1234";

    private static Path tokenDir;
    private static Provider pkcs11Provider;

    @BeforeAll
    static void setUpToken() throws Exception {
        Assumptions.assumeTrue(new File(SOFTHSM_LIB).exists(),
            "libsofthsm2.so não encontrada — pulando testes de integração PKCS11");

        String shsmConf = System.getenv("SOFTHSM2_CONF");
        Assumptions.assumeTrue(shsmConf != null && !shsmConf.isBlank(),
            "SOFTHSM2_CONF não definido — pulando testes de integração PKCS11");

        // Cria diretório de tokens isolado ao lado do conf
        tokenDir = Paths.get(shsmConf).getParent().resolve("softhsm2-tokens");
        Files.createDirectories(tokenDir);

        // Escreve softhsm2.conf apontando para o diretório de tokens
        Files.writeString(Paths.get(shsmConf),
            "directories.tokendir = " + tokenDir + "/\n" +
            "objectstore.backend = file\n" +
            "log.level = ERROR\n");

        // Inicializa o token e obtém o slot atribuído
        long slot = initToken(shsmConf);

        // Configura o provider SunPKCS11 com o slot do token criado
        String inlineCfg = "--name=SoftHSM2IT\n" +
            "library=" + SOFTHSM_LIB + "\n" +
            "slot=" + slot + "\n";

        pkcs11Provider = Security.getProvider("SunPKCS11").configure(inlineCfg);
        Security.addProvider(pkcs11Provider);

        // Faz login no token para habilitar operações com chave privada
        KeyStore ks = KeyStore.getInstance("PKCS11", pkcs11Provider);
        ks.load(null, PIN.toCharArray());
    }

    @AfterAll
    static void tearDown() throws Exception {
        if (pkcs11Provider != null) {
            Security.removeProvider(pkcs11Provider.getName());
        }
        if (tokenDir != null && Files.exists(tokenDir)) {
            Files.walk(tokenDir)
                .sorted(Comparator.reverseOrder())
                .forEach(p -> p.toFile().delete());
        }
    }

    @Test
    void pkcs11_provedorCarregadoComSucesso() {
        assertNotNull(pkcs11Provider, "Provider PKCS11 deve estar carregado");
        assertTrue(pkcs11Provider.getName().startsWith("SunPKCS11"),
            "Provider deve ser SunPKCS11, obtido: " + pkcs11Provider.getName());
    }

    @Test
    void pkcs11_geraParDeChavesRSANoToken() throws Exception {
        KeyPairGenerator kpg = KeyPairGenerator.getInstance("RSA", pkcs11Provider);
        kpg.initialize(new RSAKeyGenParameterSpec(2048, RSAKeyGenParameterSpec.F4));
        KeyPair kp = kpg.generateKeyPair();

        assertNotNull(kp.getPrivate(), "Chave privada deve ser gerada no token");
        assertNotNull(kp.getPublic(), "Chave pública deve ser gerada no token");
    }

    @Test
    void pkcs11_assinaEVerificaComChaveNoToken() throws Exception {
        KeyPairGenerator kpg = KeyPairGenerator.getInstance("RSA", pkcs11Provider);
        kpg.initialize(new RSAKeyGenParameterSpec(2048, RSAKeyGenParameterSpec.F4));
        KeyPair kp = kpg.generateKeyPair();

        byte[] data = "dados para assinar via PKCS11 real".getBytes(StandardCharsets.UTF_8);

        Signature signer = Signature.getInstance("SHA256withRSA", pkcs11Provider);
        signer.initSign(kp.getPrivate());
        signer.update(data);
        byte[] signature = signer.sign();

        assertNotNull(signature);
        assertTrue(signature.length > 0, "Assinatura não deve ser vazia");

        Signature verifier = Signature.getInstance("SHA256withRSA", pkcs11Provider);
        verifier.initVerify(kp.getPublic());
        verifier.update(data);
        assertTrue(verifier.verify(signature),
            "Assinatura produzida via PKCS11 deve ser verificável com a chave pública do token");
    }

    @Test
    void pkcs11_documentoAdulteradoDeveSerRejeitado() throws Exception {
        KeyPairGenerator kpg = KeyPairGenerator.getInstance("RSA", pkcs11Provider);
        kpg.initialize(new RSAKeyGenParameterSpec(2048, RSAKeyGenParameterSpec.F4));
        KeyPair kp = kpg.generateKeyPair();

        byte[] original = "documento original".getBytes(StandardCharsets.UTF_8);
        byte[] adulterado = "documento adulterado".getBytes(StandardCharsets.UTF_8);

        Signature signer = Signature.getInstance("SHA256withRSA", pkcs11Provider);
        signer.initSign(kp.getPrivate());
        signer.update(original);
        byte[] signature = signer.sign();

        Signature verifier = Signature.getInstance("SHA256withRSA", pkcs11Provider);
        verifier.initVerify(kp.getPublic());
        verifier.update(adulterado);
        assertFalse(verifier.verify(signature),
            "Assinatura não deve ser válida para documento adulterado");
    }

    // --- helper ---

    private static long initToken(String shsmConf) throws IOException, InterruptedException {
        ProcessBuilder pb = new ProcessBuilder(
            "softhsm2-util", "--init-token", "--free",
            "--label", "test-token",
            "--pin", PIN, "--so-pin", PIN);
        pb.environment().put("SOFTHSM2_CONF", shsmConf);
        pb.redirectErrorStream(true);
        Process p = pb.start();
        String out = new String(p.getInputStream().readAllBytes(), StandardCharsets.UTF_8);
        if (p.waitFor() != 0) {
            throw new IllegalStateException("softhsm2-util --init-token falhou: " + out);
        }
        // Saída esperada: "The token has been initialized and is reassigned to slot NNNN"
        for (String line : out.split("\\R")) {
            if (line.contains("reassigned to slot")) {
                String[] parts = line.trim().split("\\s+");
                return Long.parseLong(parts[parts.length - 1]);
            }
        }
        throw new IllegalStateException(
            "Não foi possível determinar o slot do token na saída: " + out);
    }
}
