package invoker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindJar_EnvVar(t *testing.T) {
	tmp := t.TempDir()
	jar := filepath.Join(tmp, "assinador.jar")
	if err := os.WriteFile(jar, []byte("PK"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ASSINADOR_JAR", jar)

	got, err := findJar()
	if err != nil {
		t.Fatalf("findJar() error = %v", err)
	}
	if got != jar {
		t.Errorf("findJar() = %q, want %q", got, jar)
	}
}

func TestFindJar_EnvVarInexistente(t *testing.T) {
	t.Setenv("ASSINADOR_JAR", "/nao/existe/assinador.jar")
	_, err := findJar()
	if err == nil {
		t.Fatal("esperava erro para ASSINADOR_JAR apontando para arquivo inexistente")
	}
}

func TestResponseFormat_Sucesso(t *testing.T) {
	r := &Response{Valid: true, Signature: "SIG123", Message: "Assinatura gerada"}
	got := r.Format()
	want := "[OK] Assinatura gerada\nAssinatura: SIG123"
	if got != want {
		t.Errorf("Format() = %q, want %q", got, want)
	}
}

func TestResponseFormat_SucessoSemAssinatura(t *testing.T) {
	r := &Response{Valid: true, Message: "Assinatura válida."}
	got := r.Format()
	want := "[OK] Assinatura válida."
	if got != want {
		t.Errorf("Format() = %q, want %q", got, want)
	}
}

func TestResponseFormat_Falha(t *testing.T) {
	r := &Response{Valid: false, Message: "Parâmetro ausente"}
	got := r.Format()
	want := "[FALHA] Parâmetro ausente"
	if got != want {
		t.Errorf("Format() = %q, want %q", got, want)
	}
}

// Testes de integração — requerem ASSINADOR_JAR configurado. Omitidos com -short.

func TestSign_Integration(t *testing.T) {
	skipIfNotIntegration(t)

	resp, err := Sign("documento de teste", "")
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}
	if !resp.Valid {
		t.Errorf("Sign() valid = false, message = %q", resp.Message)
	}
	if resp.Signature == "" {
		t.Error("Sign() retornou assinatura vazia")
	}
}

func TestSign_ComToken_Integration(t *testing.T) {
	skipIfNotIntegration(t)

	resp, err := Sign("documento de teste", "token123")
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}
	if !resp.Valid {
		t.Errorf("Sign() com token valid = false, message = %q", resp.Message)
	}
}

func TestValidate_AssinaturaCorreta_Integration(t *testing.T) {
	skipIfNotIntegration(t)

	resp, err := Validate("documento de teste", "MOCKED_SIGNATURE_BASE64_==")
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if !resp.Valid {
		t.Errorf("Validate() valid = false para assinatura correta, message = %q", resp.Message)
	}
}

func TestValidate_AssinaturaErrada_Integration(t *testing.T) {
	skipIfNotIntegration(t)

	resp, err := Validate("documento de teste", "assinatura-invalida")
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if resp.Valid {
		t.Error("Validate() valid = true para assinatura inválida, esperava false")
	}
}

func TestSign_ContentVazio_Integration(t *testing.T) {
	skipIfNotIntegration(t)

	resp, err := Sign("", "")
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}
	if resp.Valid {
		t.Error("Sign() com content vazio deve retornar valid = false")
	}
}

func skipIfNotIntegration(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("omitido no modo short")
	}
	if os.Getenv("ASSINADOR_JAR") == "" {
		t.Skip("ASSINADOR_JAR não definido; omitindo teste de integração")
	}
}
