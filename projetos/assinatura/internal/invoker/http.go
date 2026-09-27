package invoker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type signHTTPRequest struct {
	Content string  `json:"content"`
	Token   *string `json:"token"`
}

type validateHTTPRequest struct {
	Content   string `json:"content"`
	Signature string `json:"signature"`
}

// SignHTTP envia uma requisição de assinatura para o servidor HTTP.
func SignHTTP(baseURL, content, token string) (*Response, error) {
	req := signHTTPRequest{Content: content}
	if token != "" {
		req.Token = &token
	}
	return postJSON(baseURL+"/sign", req)
}

// ValidateHTTP envia uma requisição de validação para o servidor HTTP.
func ValidateHTTP(baseURL, content, signature string) (*Response, error) {
	req := validateHTTPRequest{Content: content, Signature: signature}
	return postJSON(baseURL+"/validate", req)
}

// FindJar retorna o caminho do assinador.jar.
func FindJar() (string, error) {
	return findJar()
}

func postJSON(url string, payload any) (*Response, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar requisição: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(body)) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao servidor: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("servidor retornou status %d", resp.StatusCode)
	}

	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("resposta inválida do servidor: %w", err)
	}
	return &result, nil
}
