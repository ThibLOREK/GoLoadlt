package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServeSpec_ContentType(t *testing.T) {
	h := NewOpenAPIHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/openapi.yaml", nil)
	w := httptest.NewRecorder()

	h.ServeSpec(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ServeSpec: status = %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/yaml" {
		t.Errorf("ServeSpec: Content-Type = %q, want \"application/yaml\"", ct)
	}
}

func TestServeSpec_BodyNotEmpty(t *testing.T) {
	h := NewOpenAPIHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/openapi.yaml", nil)
	w := httptest.NewRecorder()

	h.ServeSpec(w, req)

	body := w.Body.String()
	if len(body) == 0 {
		t.Fatal("ServeSpec: body vide, le fichier openapi.yaml n'est pas embedé")
	}
	if !strings.Contains(body, "openapi:") {
		t.Errorf("ServeSpec: le body ne contient pas \"openapi:\" — YAML invalide ou mauvais embed")
	}
}

func TestSwaggerUI_ContentType(t *testing.T) {
	h := NewOpenAPIHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/docs", nil)
	w := httptest.NewRecorder()

	h.SwaggerUI(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("SwaggerUI: status = %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("SwaggerUI: Content-Type = %q, want \"text/html\"", ct)
	}
}

func TestSwaggerUI_SpecURL(t *testing.T) {
	h := NewOpenAPIHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/docs", nil)
	w := httptest.NewRecorder()

	h.SwaggerUI(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "/api/v1/openapi.yaml") {
		t.Errorf("SwaggerUI: le body ne contient pas l'URL de spec /api/v1/openapi.yaml")
	}
	if !strings.Contains(body, "swagger-ui-bundle.js") {
		t.Errorf("SwaggerUI: le body ne contient pas swagger-ui-bundle.js (CDN unpkg)")
	}
}
