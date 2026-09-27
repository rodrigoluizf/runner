package com.kyriosdata.assinador.domain;

public class SignRequest {
    private String content;
    private String token;

    public SignRequest(String content, String token) {
        this.content = content;
        this.token = token;
    }

    public String getContent() {
        return content;
    }

    public String getToken() {
        return token;
    }
}
