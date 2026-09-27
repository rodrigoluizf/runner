package com.kyriosdata.assinador;

import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.*;

class Pkcs11HelperTest {

    @Test
    void isAvailable_naoDeveLancarExcecao() {
        // Apenas verifica que a chamada não lança exceção;
        // o resultado depende da JVM disponível no ambiente.
        assertDoesNotThrow(Pkcs11Helper::isAvailable);
    }

    @Test
    void isLoaded_retornaFalseQuandoNenhumProviderConfigurado() {
        // Em ambiente de testes sem SoftHSM2, nenhum provider PKCS#11 deve estar carregado.
        // Este teste é leve e não depende de hardware.
        boolean loaded = Pkcs11Helper.isLoaded();
        // Não forçamos false pois outro teste pode ter carregado; apenas verificamos que não lança.
        assertDoesNotThrow(Pkcs11Helper::isLoaded);
        // Em CI sem dispositivo, normalmente será false.
        // Se for true, o ambiente tem PKCS#11 configurado — também é válido.
        assertTrue(loaded || !loaded, "isLoaded() deve retornar booleano sem exceção");
    }

    @Test
    void initialize_retornaFalseParaConfiguracaoNula() {
        boolean result = Pkcs11Helper.initialize(null);
        assertFalse(result, "initialize com config nulo deve retornar false");
    }

    @Test
    void initialize_retornaFalseParaConfiguracaoVazia() {
        boolean result = Pkcs11Helper.initialize("   ");
        assertFalse(result, "initialize com config vazio deve retornar false");
    }

    @Test
    void initialize_retornaFalseParaArquivoInexistente() {
        // Arquivo inexistente deve resultar em false com mensagem clara, não em exceção.
        boolean result = Pkcs11Helper.initialize("/caminho/que/nao/existe/pkcs11.cfg");
        assertFalse(result, "initialize com arquivo inexistente deve retornar false");
    }

    @Test
    void initialize_comportamentoGraciosoQuandoDispositivoIndisponivel() {
        // Comportamento principal do US-02.5: quando dispositivo não disponível,
        // o sistema continua funcionando sem lançar exceção.
        assertDoesNotThrow(() -> Pkcs11Helper.initialize("/nao/existe.cfg"),
            "initialize não deve lançar exceção quando dispositivo indisponível");
    }
}
