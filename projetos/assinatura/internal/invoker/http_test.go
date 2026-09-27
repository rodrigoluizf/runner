package invoker

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSignHTTP_Sucesso(t *testing.T) {
	srv := fakeServer(t, "/sign", Response{
		Signature: "MOCKED_SIGNATURE_BASE64_==",
		Valid:     true,
		Message:   "Assinatura simulada gerada com sucesso.",
	})
	defer srv.Close()

	resp, err := SignHTTP(srv.URL, "documento", "")
	if err != nil {
		t.Fatalf("SignHTTP() error = %v", err)
	}
	if !resp.Valid {
		t.Errorf("SignHTTP() valid = false")
	}
	if resp.Signature == "" {
		t.Error("SignHTTP() assinatura vazia")
	}
}

func TestSignHTTP_ServidorIndisponivel(t *testing.T) {
	_, err := SignHTTP("http://localhost:19998", "doc", "")
	if err == nil {
		t.Error("SignHTTP() deveria retornar erro para servidor indisponível")
	}
}

func TestValidateHTTP_AssinaturaValida(t *testing.T) {
	srv := fakeServer(t, "/validate", Response{
		Valid:   true,
		Message: "Assinatura válida.",
	})
	defer srv.Close()

	resp, err := ValidateHTTP(srv.URL, "documento", "MOCKED_SIGNATURE_BASE64_==")
	if err != nil {
		t.Fatalf("ValidateHTTP() error = %v", err)
	}
	if !resp.Valid {
		t.Errorf("ValidateHTTP() valid = false para assinatura correta")
	}
}

func TestValidateHTTP_AssinaturaInvalida(t *testing.T) {
	srv := fakeServer(t, "/validate", Response{
		Valid:   false,
		Message: "Assinatura inválida.",
	})
	defer srv.Close()

	resp, err := ValidateHTTP(srv.URL, "documento", "errada")
	if err != nil {
		t.Fatalf("ValidateHTTP() error = %v", err)
	}
	if resp.Valid {
		t.Error("ValidateHTTP() valid = true para assinatura inválida")
	}
}

func TestPostJSON_StatusNaoOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, err := postJSON(srv.URL+"/sign", struct{}{})
	if err == nil {
		t.Error("postJSON() deveria retornar erro para status 500")
	}
}

func fakeServer(t *testing.T, path string, resp Response) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != path {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
}
