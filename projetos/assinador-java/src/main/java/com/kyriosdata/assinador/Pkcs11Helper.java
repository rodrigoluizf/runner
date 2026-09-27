package com.kyriosdata.assinador;

import java.security.Provider;
import java.security.Security;

/**
 * Auxiliar para inicializar o provider PKCS#11 via SunPKCS11.
 *
 * <p>Uso esperado: fornecer um arquivo de configuração PKCS#11 (ex.: softhsm2.cfg)
 * apontando para a biblioteca do dispositivo criptográfico. Quando o dispositivo
 * não está disponível, {@link #initialize(String)} retorna {@code false} e o
 * sistema continua operando no modo simulado.
 *
 * <p>Exemplo de arquivo de configuração (softhsm2.cfg):
 * <pre>
 * name = SoftHSM2
 * library = /usr/lib/softhsm/libsofthsm2.so
 * slot = 0
 * </pre>
 *
 * <p>Para verificar o setup:
 * <pre>
 *   softhsm2-util --init-token --slot 0 --label "test"
 *   assinador sign --content "doc" --pkcs11 --pkcs11-config softhsm2.cfg
 * </pre>
 */
public class Pkcs11Helper {

    private Pkcs11Helper() {}

    /**
     * Tenta inicializar o provider SunPKCS11 com o arquivo de configuração fornecido.
     *
     * @param configPath caminho para o arquivo de configuração PKCS#11
     * @return {@code true} se o provider foi carregado com sucesso; {@code false} caso contrário
     */
    public static boolean initialize(String configPath) {
        if (configPath == null || configPath.isBlank()) {
            System.err.println("PKCS#11: caminho de configuração não informado.");
            return false;
        }

        try {
            Provider sunPkcs11 = Security.getProvider("SunPKCS11");
            if (sunPkcs11 == null) {
                System.err.println("PKCS#11: provider SunPKCS11 não disponível nesta JVM.");
                return false;
            }

            Provider configured = sunPkcs11.configure(configPath);
            Security.addProvider(configured);
            System.err.println("PKCS#11: provider configurado com sucesso via " + configPath);
            return true;

        } catch (Exception e) {
            System.err.println("PKCS#11: falha ao inicializar (" + e.getMessage() + "). "
                + "Verifique se o dispositivo está conectado e o arquivo de configuração é válido.");
            return false;
        }
    }

    /**
     * Retorna {@code true} se o provider SunPKCS11 está disponível nesta JVM.
     */
    public static boolean isAvailable() {
        return Security.getProvider("SunPKCS11") != null;
    }

    /**
     * Retorna {@code true} se um provider PKCS#11 já foi adicionado à configuração de segurança.
     */
    public static boolean isLoaded() {
        for (Provider p : Security.getProviders()) {
            if (p.getName().startsWith("SunPKCS11-")) {
                return true;
            }
        }
        return false;
    }
}
