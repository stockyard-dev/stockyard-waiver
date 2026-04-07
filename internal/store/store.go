package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"
	_ "modernc.org/sqlite"
)

type DB struct { db *sql.DB }

type Templates struct {
	ID string `json:"id"`
	Title string `json:"title"`
	Body string `json:"body"`
	RequiresSignature bool `json:"requires_signature"`
	Active bool `json:"active"`
	CreatedAt string `json:"created_at"`
}

type Signatures struct {
	ID string `json:"id"`
	SignerName string `json:"signer_name"`
	SignerEmail string `json:"signer_email"`
	TemplateId string `json:"template_id"`
	SignatureData string `json:"signature_data"`
	IpAddress string `json:"ip_address"`
	SignedAt string `json:"signed_at"`
	Status string `json:"status"`
	CreatedAt string `json:"created_at"`
}

func Open(d string) (*DB, error) {
	if err := os.MkdirAll(d, 0755); err != nil { return nil, err }
	db, err := sql.Open("sqlite", filepath.Join(d, "waiver.db")+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil { return nil, err }
	db.SetMaxOpenConns(1)
	db.Exec(`CREATE TABLE IF NOT EXISTS templates(id TEXT PRIMARY KEY, title TEXT NOT NULL, body TEXT NOT NULL, requires_signature INTEGER DEFAULT 0, active INTEGER DEFAULT 0, created_at TEXT DEFAULT(datetime('now')))`)
	db.Exec(`CREATE TABLE IF NOT EXISTS signatures(id TEXT PRIMARY KEY, signer_name TEXT NOT NULL, signer_email TEXT DEFAULT '', template_id TEXT DEFAULT '', signature_data TEXT DEFAULT '', ip_address TEXT DEFAULT '', signed_at TEXT DEFAULT '', status TEXT DEFAULT '', created_at TEXT DEFAULT(datetime('now')))`)
	db.Exec(`CREATE TABLE IF NOT EXISTS extras(resource TEXT NOT NULL, record_id TEXT NOT NULL, data TEXT NOT NULL DEFAULT '{}', PRIMARY KEY(resource, record_id))`)
	return &DB{db: db}, nil
}

func (d *DB) Close() error { return d.db.Close() }
func genID() string { return fmt.Sprintf("%d", time.Now().UnixNano()) }
func now() string { return time.Now().UTC().Format(time.RFC3339) }

func (d *DB) CreateTemplates(e *Templates) error {
	e.ID = genID(); e.CreatedAt = now()
	_, err := d.db.Exec(`INSERT INTO templates(id, title, body, requires_signature, active, created_at) VALUES(?, ?, ?, ?, ?, ?)`, e.ID, e.Title, e.Body, e.RequiresSignature, e.Active, e.CreatedAt)
	return err
}

func (d *DB) GetTemplates(id string) *Templates {
	var e Templates
	if d.db.QueryRow(`SELECT id, title, body, requires_signature, active, created_at FROM templates WHERE id=?`, id).Scan(&e.ID, &e.Title, &e.Body, &e.RequiresSignature, &e.Active, &e.CreatedAt) != nil { return nil }
	return &e
}

func (d *DB) ListTemplates() []Templates {
	rows, _ := d.db.Query(`SELECT id, title, body, requires_signature, active, created_at FROM templates ORDER BY created_at DESC`)
	if rows == nil { return nil }; defer rows.Close()
	var o []Templates
	for rows.Next() { var e Templates; rows.Scan(&e.ID, &e.Title, &e.Body, &e.RequiresSignature, &e.Active, &e.CreatedAt); o = append(o, e) }
	return o
}

func (d *DB) UpdateTemplates(e *Templates) error {
	_, err := d.db.Exec(`UPDATE templates SET title=?, body=?, requires_signature=?, active=? WHERE id=?`, e.Title, e.Body, e.RequiresSignature, e.Active, e.ID)
	return err
}

func (d *DB) DeleteTemplates(id string) error {
	_, err := d.db.Exec(`DELETE FROM templates WHERE id=?`, id)
	return err
}

func (d *DB) CountTemplates() int {
	var n int; d.db.QueryRow(`SELECT COUNT(*) FROM templates`).Scan(&n); return n
}

func (d *DB) SearchTemplates(q string, filters map[string]string) []Templates {
	where := "1=1"
	args := []any{}
	if q != "" {
		where += " AND (title LIKE ? OR body LIKE ?)"
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
	}
	rows, _ := d.db.Query(`SELECT id, title, body, requires_signature, active, created_at FROM templates WHERE `+where+` ORDER BY created_at DESC`, args...)
	if rows == nil { return nil }; defer rows.Close()
	var o []Templates
	for rows.Next() { var e Templates; rows.Scan(&e.ID, &e.Title, &e.Body, &e.RequiresSignature, &e.Active, &e.CreatedAt); o = append(o, e) }
	return o
}

func (d *DB) CreateSignatures(e *Signatures) error {
	e.ID = genID(); e.CreatedAt = now()
	_, err := d.db.Exec(`INSERT INTO signatures(id, signer_name, signer_email, template_id, signature_data, ip_address, signed_at, status, created_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)`, e.ID, e.SignerName, e.SignerEmail, e.TemplateId, e.SignatureData, e.IpAddress, e.SignedAt, e.Status, e.CreatedAt)
	return err
}

func (d *DB) GetSignatures(id string) *Signatures {
	var e Signatures
	if d.db.QueryRow(`SELECT id, signer_name, signer_email, template_id, signature_data, ip_address, signed_at, status, created_at FROM signatures WHERE id=?`, id).Scan(&e.ID, &e.SignerName, &e.SignerEmail, &e.TemplateId, &e.SignatureData, &e.IpAddress, &e.SignedAt, &e.Status, &e.CreatedAt) != nil { return nil }
	return &e
}

func (d *DB) ListSignatures() []Signatures {
	rows, _ := d.db.Query(`SELECT id, signer_name, signer_email, template_id, signature_data, ip_address, signed_at, status, created_at FROM signatures ORDER BY created_at DESC`)
	if rows == nil { return nil }; defer rows.Close()
	var o []Signatures
	for rows.Next() { var e Signatures; rows.Scan(&e.ID, &e.SignerName, &e.SignerEmail, &e.TemplateId, &e.SignatureData, &e.IpAddress, &e.SignedAt, &e.Status, &e.CreatedAt); o = append(o, e) }
	return o
}

func (d *DB) UpdateSignatures(e *Signatures) error {
	_, err := d.db.Exec(`UPDATE signatures SET signer_name=?, signer_email=?, template_id=?, signature_data=?, ip_address=?, signed_at=?, status=? WHERE id=?`, e.SignerName, e.SignerEmail, e.TemplateId, e.SignatureData, e.IpAddress, e.SignedAt, e.Status, e.ID)
	return err
}

func (d *DB) DeleteSignatures(id string) error {
	_, err := d.db.Exec(`DELETE FROM signatures WHERE id=?`, id)
	return err
}

func (d *DB) CountSignatures() int {
	var n int; d.db.QueryRow(`SELECT COUNT(*) FROM signatures`).Scan(&n); return n
}

func (d *DB) SearchSignatures(q string, filters map[string]string) []Signatures {
	where := "1=1"
	args := []any{}
	if q != "" {
		where += " AND (signer_name LIKE ? OR signer_email LIKE ? OR template_id LIKE ? OR signature_data LIKE ? OR ip_address LIKE ?)"
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
	}
	if v, ok := filters["status"]; ok && v != "" { where += " AND status=?"; args = append(args, v) }
	rows, _ := d.db.Query(`SELECT id, signer_name, signer_email, template_id, signature_data, ip_address, signed_at, status, created_at FROM signatures WHERE `+where+` ORDER BY created_at DESC`, args...)
	if rows == nil { return nil }; defer rows.Close()
	var o []Signatures
	for rows.Next() { var e Signatures; rows.Scan(&e.ID, &e.SignerName, &e.SignerEmail, &e.TemplateId, &e.SignatureData, &e.IpAddress, &e.SignedAt, &e.Status, &e.CreatedAt); o = append(o, e) }
	return o
}

// GetExtras returns the JSON extras blob for a record. Returns "{}" if none.
func (d *DB) GetExtras(resource, recordID string) string {
	var data string
	err := d.db.QueryRow(`SELECT data FROM extras WHERE resource=? AND record_id=?`, resource, recordID).Scan(&data)
	if err != nil || data == "" {
		return "{}"
	}
	return data
}

// SetExtras stores the JSON extras blob for a record.
func (d *DB) SetExtras(resource, recordID, data string) error {
	if data == "" {
		data = "{}"
	}
	_, err := d.db.Exec(`INSERT INTO extras(resource, record_id, data) VALUES(?, ?, ?) ON CONFLICT(resource, record_id) DO UPDATE SET data=excluded.data`, resource, recordID, data)
	return err
}

// DeleteExtras removes extras when a record is deleted.
func (d *DB) DeleteExtras(resource, recordID string) error {
	_, err := d.db.Exec(`DELETE FROM extras WHERE resource=? AND record_id=?`, resource, recordID)
	return err
}

// AllExtras returns all extras for a resource type as a map of record_id → JSON string.
func (d *DB) AllExtras(resource string) map[string]string {
	out := make(map[string]string)
	rows, _ := d.db.Query(`SELECT record_id, data FROM extras WHERE resource=?`, resource)
	if rows == nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var id, data string
		rows.Scan(&id, &data)
		out[id] = data
	}
	return out
}
