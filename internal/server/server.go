package server

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/stockyard-dev/stockyard-waiver/internal/store"
)

type Server struct {
	db     *store.DB
	mux    *http.ServeMux
	limits  Limits
	dataDir string
	pCfg    map[string]json.RawMessage
}

func New(db *store.DB, limits Limits, dataDir string) *Server {
	s := &Server{db: db, mux: http.NewServeMux(), limits: limits, dataDir: dataDir}
	s.loadPersonalConfig()
	s.mux.HandleFunc("GET /api/templates", s.listTemplates)
	s.mux.HandleFunc("POST /api/templates", s.createTemplates)
	s.mux.HandleFunc("GET /api/templates/export.csv", s.exportTemplates)
	s.mux.HandleFunc("GET /api/templates/{id}", s.getTemplates)
	s.mux.HandleFunc("PUT /api/templates/{id}", s.updateTemplates)
	s.mux.HandleFunc("DELETE /api/templates/{id}", s.delTemplates)
	s.mux.HandleFunc("GET /api/signatures", s.listSignatures)
	s.mux.HandleFunc("POST /api/signatures", s.createSignatures)
	s.mux.HandleFunc("GET /api/signatures/export.csv", s.exportSignatures)
	s.mux.HandleFunc("GET /api/signatures/{id}", s.getSignatures)
	s.mux.HandleFunc("PUT /api/signatures/{id}", s.updateSignatures)
	s.mux.HandleFunc("DELETE /api/signatures/{id}", s.delSignatures)
	s.mux.HandleFunc("GET /api/stats", s.stats)
	s.mux.HandleFunc("GET /api/health", s.health)
	s.mux.HandleFunc("GET /health", s.health)
	s.mux.HandleFunc("GET /ui", s.dashboard)
	s.mux.HandleFunc("GET /ui/", s.dashboard)
	s.mux.HandleFunc("GET /", s.root)
	s.mux.HandleFunc("GET /api/tier", s.tierHandler)
	s.mux.HandleFunc("GET /api/config", s.configHandler)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.mux.ServeHTTP(w, r) }
func wj(w http.ResponseWriter, c int, v any) { w.Header().Set("Content-Type", "application/json"); w.WriteHeader(c); json.NewEncoder(w).Encode(v) }
func we(w http.ResponseWriter, c int, m string) { wj(w, c, map[string]string{"error": m}) }
func (s *Server) root(w http.ResponseWriter, r *http.Request) { if r.URL.Path != "/" { http.NotFound(w, r); return }; http.Redirect(w, r, "/ui", 302) }
func oe[T any](s []T) []T { if s == nil { return []T{} }; return s }
func init() { log.SetFlags(log.LstdFlags | log.Lshortfile) }

func (s *Server) listTemplates(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	filters := map[string]string{}
	if q != "" || len(filters) > 0 { wj(w, 200, map[string]any{"templates": oe(s.db.SearchTemplates(q, filters))}); return }
	wj(w, 200, map[string]any{"templates": oe(s.db.ListTemplates())})
}

func (s *Server) createTemplates(w http.ResponseWriter, r *http.Request) {
	if s.limits.Tier == "none" { we(w, 402, "No license key. Start a 14-day trial at https://stockyard.dev/for/"); return }
	if s.limits.TrialExpired { we(w, 402, "Trial expired. Subscribe at https://stockyard.dev/pricing/"); return }
	var e store.Templates
	json.NewDecoder(r.Body).Decode(&e)
	if e.Title == "" { we(w, 400, "title required"); return }
	if e.Body == "" { we(w, 400, "body required"); return }
	s.db.CreateTemplates(&e)
	wj(w, 201, s.db.GetTemplates(e.ID))
}

func (s *Server) getTemplates(w http.ResponseWriter, r *http.Request) {
	e := s.db.GetTemplates(r.PathValue("id"))
	if e == nil { we(w, 404, "not found"); return }
	wj(w, 200, e)
}

func (s *Server) updateTemplates(w http.ResponseWriter, r *http.Request) {
	existing := s.db.GetTemplates(r.PathValue("id"))
	if existing == nil { we(w, 404, "not found"); return }
	var patch store.Templates
	json.NewDecoder(r.Body).Decode(&patch)
	patch.ID = existing.ID; patch.CreatedAt = existing.CreatedAt
	if patch.Title == "" { patch.Title = existing.Title }
	if patch.Body == "" { patch.Body = existing.Body }
	s.db.UpdateTemplates(&patch)
	wj(w, 200, s.db.GetTemplates(patch.ID))
}

func (s *Server) delTemplates(w http.ResponseWriter, r *http.Request) {
	s.db.DeleteTemplates(r.PathValue("id"))
	wj(w, 200, map[string]string{"deleted": "ok"})
}

func (s *Server) exportTemplates(w http.ResponseWriter, r *http.Request) {
	items := s.db.ListTemplates()
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=templates.csv")
	cw := csv.NewWriter(w)
	cw.Write([]string{"id", "title", "body", "requires_signature", "active", "created_at"})
	for _, e := range items {
		cw.Write([]string{e.ID, fmt.Sprintf("%v", e.Title), fmt.Sprintf("%v", e.Body), fmt.Sprintf("%v", e.RequiresSignature), fmt.Sprintf("%v", e.Active), e.CreatedAt})
	}
	cw.Flush()
}

func (s *Server) listSignatures(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	filters := map[string]string{}
	if v := r.URL.Query().Get("status"); v != "" { filters["status"] = v }
	if q != "" || len(filters) > 0 { wj(w, 200, map[string]any{"signatures": oe(s.db.SearchSignatures(q, filters))}); return }
	wj(w, 200, map[string]any{"signatures": oe(s.db.ListSignatures())})
}

func (s *Server) createSignatures(w http.ResponseWriter, r *http.Request) {
	var e store.Signatures
	json.NewDecoder(r.Body).Decode(&e)
	if e.SignerName == "" { we(w, 400, "signer_name required"); return }
	s.db.CreateSignatures(&e)
	wj(w, 201, s.db.GetSignatures(e.ID))
}

func (s *Server) getSignatures(w http.ResponseWriter, r *http.Request) {
	e := s.db.GetSignatures(r.PathValue("id"))
	if e == nil { we(w, 404, "not found"); return }
	wj(w, 200, e)
}

func (s *Server) updateSignatures(w http.ResponseWriter, r *http.Request) {
	existing := s.db.GetSignatures(r.PathValue("id"))
	if existing == nil { we(w, 404, "not found"); return }
	var patch store.Signatures
	json.NewDecoder(r.Body).Decode(&patch)
	patch.ID = existing.ID; patch.CreatedAt = existing.CreatedAt
	if patch.SignerName == "" { patch.SignerName = existing.SignerName }
	if patch.SignerEmail == "" { patch.SignerEmail = existing.SignerEmail }
	if patch.TemplateId == "" { patch.TemplateId = existing.TemplateId }
	if patch.SignatureData == "" { patch.SignatureData = existing.SignatureData }
	if patch.IpAddress == "" { patch.IpAddress = existing.IpAddress }
	if patch.SignedAt == "" { patch.SignedAt = existing.SignedAt }
	if patch.Status == "" { patch.Status = existing.Status }
	s.db.UpdateSignatures(&patch)
	wj(w, 200, s.db.GetSignatures(patch.ID))
}

func (s *Server) delSignatures(w http.ResponseWriter, r *http.Request) {
	s.db.DeleteSignatures(r.PathValue("id"))
	wj(w, 200, map[string]string{"deleted": "ok"})
}

func (s *Server) exportSignatures(w http.ResponseWriter, r *http.Request) {
	items := s.db.ListSignatures()
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=signatures.csv")
	cw := csv.NewWriter(w)
	cw.Write([]string{"id", "signer_name", "signer_email", "template_id", "signature_data", "ip_address", "signed_at", "status", "created_at"})
	for _, e := range items {
		cw.Write([]string{e.ID, fmt.Sprintf("%v", e.SignerName), fmt.Sprintf("%v", e.SignerEmail), fmt.Sprintf("%v", e.TemplateId), fmt.Sprintf("%v", e.SignatureData), fmt.Sprintf("%v", e.IpAddress), fmt.Sprintf("%v", e.SignedAt), fmt.Sprintf("%v", e.Status), e.CreatedAt})
	}
	cw.Flush()
}

func (s *Server) stats(w http.ResponseWriter, r *http.Request) {
	m := map[string]any{}
	m["templates_total"] = s.db.CountTemplates()
	m["signatures_total"] = s.db.CountSignatures()
	wj(w, 200, m)
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	m := map[string]any{"status": "ok", "service": "waiver"}
	m["templates"] = s.db.CountTemplates()
	m["signatures"] = s.db.CountSignatures()
	wj(w, 200, m)
}

// loadPersonalConfig reads config.json from the data directory.
func (s *Server) loadPersonalConfig() {
	path := filepath.Join(s.dataDir, "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var cfg map[string]json.RawMessage
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Printf("warning: could not parse config.json: %v", err)
		return
	}
	s.pCfg = cfg
	log.Printf("loaded personalization from %s", path)
}

func (s *Server) configHandler(w http.ResponseWriter, r *http.Request) {
	if s.pCfg == nil {
		wj(w, 200, map[string]any{})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.pCfg)
}
