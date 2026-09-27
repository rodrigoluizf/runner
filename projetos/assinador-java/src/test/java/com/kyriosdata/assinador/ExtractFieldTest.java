package com.kyriosdata.assinador;

import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.*;

class ExtractFieldTest {

    @Test
    void extraiCampoSimples() {
        assertEquals("hello", AssinadorServer.extractField("{\"content\":\"hello\"}", "content"));
    }

    @Test
    void extraiCampoComEspaco() {
        assertEquals("hello", AssinadorServer.extractField("{\"content\": \"hello\"}", "content"));
    }

    @Test
    void retornaNullParaCampoNulo() {
        assertNull(AssinadorServer.extractField("{\"content\":null}", "content"));
    }

    @Test
    void retornaNullParaCampoAusente() {
        assertNull(AssinadorServer.extractField("{\"token\":\"x\"}", "content"));
    }

    @Test
    void extraiStringComCaracteresEscapados() {
        assertEquals("a\"b", AssinadorServer.extractField("{\"content\":\"a\\\"b\"}", "content"));
    }

    @Test
    void extraiSegundoCampo() {
        String json = "{\"content\":\"doc\",\"signature\":\"SIG==\"}";
        assertEquals("SIG==", AssinadorServer.extractField(json, "signature"));
        assertEquals("doc", AssinadorServer.extractField(json, "content"));
    }
}
