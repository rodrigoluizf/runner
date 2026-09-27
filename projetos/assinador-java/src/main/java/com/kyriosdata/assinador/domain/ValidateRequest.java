package com.kyriosdata.assinador.domain;

public class ValidateRequest {
    private String content;
    private String signature;

    public ValidateRequest(String content, String signature) {
        this.content = content;
        this.signature = signature;
    }

    public String getContent() {
        return content;
    }

    public String getSignature() {
        return signature;
    }
}
