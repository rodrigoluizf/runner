package com.kyriosdata.assinador;

import com.kyriosdata.assinador.domain.SignRequest;
import com.kyriosdata.assinador.domain.SignatureResponse;
import com.kyriosdata.assinador.domain.ValidateRequest;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.*;

class FakeSignatureServiceTest {

    private FakeSignatureService service;

    @BeforeEach
    void setUp() {
        service = new FakeSignatureService();
    }

    // ── sign: cenários de sucesso ─────────────────────────────────────────────

    @Test
    void deveRetornarAssinaturaSimuladaParaRequisicaoValida() {
        SignRequest request = new SignRequest("conteudo do documento", null);
        SignatureResponse response = service.sign(request);

        assertTrue(response.isValid());
        assertEquals(FakeSignatureService.FAKE_SIGNATURE, response.getSignature());
        assertNotNull(response.getMessage());
    }

    @Test
    void deveRetornarAssinaturaSimuladaComToken() {
        SignRequest request = new SignRequest("documento", "meu-token");
        SignatureResponse response = service.sign(request);

        assertTrue(response.isValid());
        assertEquals(FakeSignatureService.FAKE_SIGNATURE, response.getSignature());
    }

    // ── sign: validação de parâmetros ─────────────────────────────────────────

    @Test
    void deveRejeitarRequisicaoDeAssinaturaComConteudoNulo() {
        SignRequest request = new SignRequest(null, null);
        SignatureResponse response = service.sign(request);

        assertFalse(response.isValid());
        assertNotNull(response.getMessage());
        assertTrue(response.getMessage().contains("'content'"),
            "Mensagem deveria identificar o parâmetro inválido: " + response.getMessage());
    }

    @Test
    void deveRejeitarRequisicaoDeAssinaturaComConteudoVazio() {
        SignRequest request = new SignRequest("  ", null);
        SignatureResponse response = service.sign(request);

        assertFalse(response.isValid());
        assertNotNull(response.getMessage());
    }

    @Test
    void deveRejeitarRequisicaoDeAssinaturaComConteudoString() {
        SignRequest request = new SignRequest("", null);
        SignatureResponse response = service.sign(request);

        assertFalse(response.isValid());
    }

    @Test
    void deveRejeitarConteudoExcessivamenteLongo() {
        String conteudoLongo = "a".repeat(FakeSignatureService.MAX_CONTENT_LENGTH + 1);
        SignRequest request = new SignRequest(conteudoLongo, null);
        SignatureResponse response = service.sign(request);

        assertFalse(response.isValid());
        assertNotNull(response.getMessage());
        assertTrue(response.getMessage().contains("tamanho máximo"),
            "Mensagem deveria mencionar tamanho máximo: " + response.getMessage());
    }

    @Test
    void deveAceitarConteudoComTamanhoMaximo() {
        String conteudoMaximo = "a".repeat(FakeSignatureService.MAX_CONTENT_LENGTH);
        SignRequest request = new SignRequest(conteudoMaximo, null);
        SignatureResponse response = service.sign(request);

        assertTrue(response.isValid());
    }

    // ── validate: cenários de sucesso ─────────────────────────────────────────

    @Test
    void deveValidarAssinaturaSimuladaCorretamente() {
        ValidateRequest request = new ValidateRequest("conteudo", FakeSignatureService.FAKE_SIGNATURE);
        SignatureResponse response = service.validate(request);

        assertTrue(response.isValid());
        assertEquals("Assinatura válida.", response.getMessage());
    }

    // ── validate: validação de parâmetros ─────────────────────────────────────

    @Test
    void deveRejeiarAssinaturaInvalida() {
        ValidateRequest request = new ValidateRequest("conteudo", "assinatura-errada");
        SignatureResponse response = service.validate(request);

        assertFalse(response.isValid());
        assertEquals("Assinatura inválida.", response.getMessage());
    }

    @Test
    void deveRejeitarValidacaoComConteudoNulo() {
        ValidateRequest request = new ValidateRequest(null, FakeSignatureService.FAKE_SIGNATURE);
        SignatureResponse response = service.validate(request);

        assertFalse(response.isValid());
        assertNotNull(response.getMessage());
        assertTrue(response.getMessage().contains("'content'"),
            "Mensagem deveria identificar o parâmetro inválido: " + response.getMessage());
    }

    @Test
    void deveRejeitarValidacaoComConteudoVazio() {
        ValidateRequest request = new ValidateRequest("", FakeSignatureService.FAKE_SIGNATURE);
        SignatureResponse response = service.validate(request);

        assertFalse(response.isValid());
    }

    @Test
    void deveRejeitarValidacaoComAssinaturaNula() {
        ValidateRequest request = new ValidateRequest("conteudo", null);
        SignatureResponse response = service.validate(request);

        assertFalse(response.isValid());
        assertNotNull(response.getMessage());
        assertTrue(response.getMessage().contains("'signature'"),
            "Mensagem deveria identificar o parâmetro inválido: " + response.getMessage());
    }

    @Test
    void deveRejeitarValidacaoComAssinaturaVazia() {
        ValidateRequest request = new ValidateRequest("conteudo", "   ");
        SignatureResponse response = service.validate(request);

        assertFalse(response.isValid());
        assertNotNull(response.getMessage());
    }

    @Test
    void deveRejeitarValidacaoComConteudoLongo() {
        String conteudoLongo = "x".repeat(FakeSignatureService.MAX_CONTENT_LENGTH + 1);
        ValidateRequest request = new ValidateRequest(conteudoLongo, FakeSignatureService.FAKE_SIGNATURE);
        SignatureResponse response = service.validate(request);

        assertFalse(response.isValid());
    }

    @Test
    void deveRejeitar_content_antes_de_verificar_signature() {
        // Garante que content é validado antes de signature (ordem correta)
        ValidateRequest request = new ValidateRequest(null, null);
        SignatureResponse response = service.validate(request);

        assertFalse(response.isValid());
        assertTrue(response.getMessage().contains("'content'"),
            "Com ambos nulos, 'content' deve ser validado primeiro: " + response.getMessage());
    }

    // ── mensagens de erro ─────────────────────────────────────────────────────

    @Test
    void mensagemDeErroDeveIndicarParametroInvalido() {
        SignRequest request = new SignRequest(null, null);
        SignatureResponse response = service.sign(request);

        assertFalse(response.isValid());
        assertNotNull(response.getMessage());
        assertFalse(response.getMessage().isBlank(), "Mensagem de erro não deve ser vazia");
    }

    @Test
    void mensagemDeSucessoDeveExistir() {
        SignRequest request = new SignRequest("documento", null);
        SignatureResponse response = service.sign(request);

        assertTrue(response.isValid());
        assertNotNull(response.getMessage());
        assertFalse(response.getMessage().isBlank(), "Mensagem de sucesso não deve ser vazia");
    }
}
