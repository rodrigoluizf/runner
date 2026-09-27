package com.kyriosdata.assinador;

import com.kyriosdata.assinador.domain.SignRequest;
import com.kyriosdata.assinador.domain.SignatureResponse;
import com.kyriosdata.assinador.domain.ValidateRequest;

public class Main {

    public static void main(String[] args) {
        if (args.length < 1) {
            System.err.println("Uso: assinador <sign|validate|server> [opções]");
            System.err.println("  --content <valor>       Conteúdo a assinar ou validar");
            System.err.println("  --token <valor>         Token de autenticação (opcional)");
            System.err.println("  --signature <valor>     Assinatura a validar");
            System.err.println("  --pkcs11                Habilitar suporte a dispositivo criptográfico PKCS#11");
            System.err.println("  --pkcs11-config <path>  Caminho para o arquivo de configuração PKCS#11");
            System.err.println("  --port <número>         Porta do servidor HTTP (modo server, padrão: 8080)");
            System.err.println("  --timeout <minutos>     Encerrar servidor após N minutos de inatividade");
            System.exit(2);
        }

        String operation = args[0];

        if ("server".equals(operation)) {
            int port = 8080;
            int timeout = 0;
            boolean pkcs11 = false;
            String pkcs11Config = null;
            for (int i = 1; i < args.length; i++) {
                if ("--port".equals(args[i]) && i + 1 < args.length) port = Integer.parseInt(args[++i]);
                else if ("--timeout".equals(args[i]) && i + 1 < args.length) timeout = Integer.parseInt(args[++i]);
                else if ("--pkcs11".equals(args[i])) pkcs11 = true;
                else if ("--pkcs11-config".equals(args[i]) && i + 1 < args.length) pkcs11Config = args[++i];
            }
            if (pkcs11) {
                Pkcs11Helper.initialize(pkcs11Config);
            }
            try {
                new AssinadorServer(port, timeout).start();
                Thread.currentThread().join();
            } catch (Exception e) {
                System.err.println("Erro ao iniciar servidor: " + e.getMessage());
                System.exit(1);
            }
            return;
        }

        String content = null;
        String token = null;
        String signature = null;
        boolean pkcs11 = false;
        String pkcs11Config = null;

        for (int i = 1; i < args.length; i++) {
            switch (args[i]) {
                case "--content":
                    if (i + 1 < args.length) content = args[++i];
                    break;
                case "--token":
                    if (i + 1 < args.length) token = args[++i];
                    break;
                case "--signature":
                    if (i + 1 < args.length) signature = args[++i];
                    break;
                case "--pkcs11":
                    pkcs11 = true;
                    break;
                case "--pkcs11-config":
                    if (i + 1 < args.length) pkcs11Config = args[++i];
                    break;
                default:
                    System.err.println("Parâmetro desconhecido: " + args[i]);
                    System.exit(2);
            }
        }

        if (pkcs11) {
            boolean loaded = Pkcs11Helper.initialize(pkcs11Config);
            if (!loaded) {
                System.err.println("Aviso: dispositivo PKCS#11 não disponível. Operando no modo simulado.");
            }
        }

        FakeSignatureService service = new FakeSignatureService();
        SignatureResponse response;

        switch (operation) {
            case "sign":
                response = service.sign(new SignRequest(content, token));
                break;
            case "validate":
                response = service.validate(new ValidateRequest(content, signature));
                break;
            default:
                System.err.println("Operação desconhecida: " + operation + ". Use 'sign', 'validate' ou 'server'.");
                System.exit(2);
                return;
        }

        System.out.println(response.toJson());
    }
}
