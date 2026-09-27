package com.kyriosdata.assinador;

import com.kyriosdata.assinador.domain.SignRequest;
import com.kyriosdata.assinador.domain.SignatureResponse;
import com.kyriosdata.assinador.domain.ValidateRequest;

public class FakeSignatureService implements SignatureService {

    static final String FAKE_SIGNATURE = "MOCKED_SIGNATURE_BASE64_==";

    /** Comprimento mínimo do conteúdo (em caracteres, após trim). */
    static final int MIN_CONTENT_LENGTH = 1;

    /** Comprimento máximo do conteúdo (em caracteres). */
    static final int MAX_CONTENT_LENGTH = 65536;

    @Override
    public SignatureResponse sign(SignRequest request) {
        String contentError = validateContent(request == null ? null : request.getContent());
        if (contentError != null) {
            return new SignatureResponse(null, false, contentError);
        }
        return new SignatureResponse(FAKE_SIGNATURE, true, "Assinatura simulada gerada com sucesso.");
    }

    @Override
    public SignatureResponse validate(ValidateRequest request) {
        if (request == null) {
            return new SignatureResponse(null, false, "Requisição de validação não pode ser nula.");
        }
        String contentError = validateContent(request.getContent());
        if (contentError != null) {
            return new SignatureResponse(null, false, contentError);
        }
        String signatureError = validateSignature(request.getSignature());
        if (signatureError != null) {
            return new SignatureResponse(null, false, signatureError);
        }
        boolean valid = FAKE_SIGNATURE.equals(request.getSignature());
        String message = valid ? "Assinatura válida." : "Assinatura inválida.";
        return new SignatureResponse(null, valid, message);
    }

    private static String validateContent(String content) {
        if (content == null) {
            return "Parâmetro 'content' é obrigatório e não pode ser nulo.";
        }
        if (content.isBlank()) {
            return "Parâmetro 'content' é obrigatório e não pode ser vazio.";
        }
        if (content.length() > MAX_CONTENT_LENGTH) {
            return "Parâmetro 'content' excede o tamanho máximo de " + MAX_CONTENT_LENGTH + " caracteres.";
        }
        return null;
    }

    private static String validateSignature(String signature) {
        if (signature == null) {
            return "Parâmetro 'signature' é obrigatório e não pode ser nulo.";
        }
        if (signature.isBlank()) {
            return "Parâmetro 'signature' é obrigatório e não pode ser vazio.";
        }
        return null;
    }
}
