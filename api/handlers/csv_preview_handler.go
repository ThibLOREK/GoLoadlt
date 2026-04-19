package handlers

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

type csvPreviewResponse struct {
	Success bool                `json:"success"`
	Columns []string            `json:"columns"`
	Rows    []map[string]string `json:"rows"`
	Meta    map[string]any      `json:"meta"`
	Error   string              `json:"error,omitempty"`
}

func (h *ProjectHandler) CSVPreview(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "projectID")

	path := strings.TrimSpace(r.URL.Query().Get("path"))
	if path == "" {
		writeJSON(w, http.StatusBadRequest, csvPreviewResponse{Success: false, Error: "paramètre 'path' manquant"})
		return
	}

	encodingName := normalizeEncodingPreview(r.URL.Query().Get("encoding"))
	delimiter := parseDelimiterPreview(r.URL.Query().Get("delimiter"))
	hasHeader := parseBoolDefaultPreview(r.URL.Query().Get("has_header"), true)
	headers := parseHeaderListPreview(r.URL.Query().Get("headers"))
	newline := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("newline")))
	previewRows := parseIntDefaultPreview(r.URL.Query().Get("limit"), 20)
	if previewRows <= 0 || previewRows > 100 {
		previewRows = 20
	}

	f, err := os.Open(path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, csvPreviewResponse{Success: false, Error: fmt.Sprintf("ouverture fichier: %v", err)})
		return
	}
	defer f.Close()

	var rr io.Reader = f
	if decoder := decoderForEncodingPreview(encodingName); decoder != nil {
		rr = transform.NewReader(rr, decoder)
	}
	br := bufio.NewReader(rr)
	if newline == "cr" {
		br = bufio.NewReader(newCRToLFReaderPreview(br))
	}

	reader := csv.NewReader(br)
	reader.Comma = delimiter
	reader.LazyQuotes = parseBoolDefaultPreview(r.URL.Query().Get("lazy_quotes"), true)
	reader.TrimLeadingSpace = parseBoolDefaultPreview(r.URL.Query().Get("trim_leading_space"), true)
	reader.FieldsPerRecord = parseIntDefaultPreview(r.URL.Query().Get("fields_per_record"), -1)

	columns := headers
	if hasHeader {
		columns, err = reader.Read()
		if err != nil {
			writeJSON(w, http.StatusBadRequest, csvPreviewResponse{Success: false, Error: fmt.Sprintf("lecture en-tête: %v", err)})
			return
		}
		for i := range columns {
			columns[i] = strings.TrimSpace(columns[i])
			if columns[i] == "" {
				columns[i] = fmt.Sprintf("column_%d", i+1)
			}
		}
	}
	if len(columns) == 0 {
		writeJSON(w, http.StatusBadRequest, csvPreviewResponse{Success: false, Error: "sans en-tête, le paramètre 'headers' est requis"})
		return
	}

	rows := make([]map[string]string, 0, previewRows)
	for len(rows) < previewRows {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			writeJSON(w, http.StatusBadRequest, csvPreviewResponse{Success: false, Error: fmt.Sprintf("lecture ligne: %v", err)})
			return
		}
		row := make(map[string]string, len(columns))
		for i, col := range columns {
			if i < len(rec) {
				row[col] = rec[i]
			} else {
				row[col] = ""
			}
		}
		rows = append(rows, row)
	}

	writeJSON(w, http.StatusOK, csvPreviewResponse{
		Success: true,
		Columns: columns,
		Rows:    rows,
		Meta: map[string]any{
			"encoding":      encodingName,
			"delimiter":     string(delimiter),
			"hasHeader":     hasHeader,
			"newline":       newline,
			"previewCount":  len(rows),
			"requestedLimit": previewRows,
		},
	})
}

func parseDelimiterPreview(v string) rune {
	v = strings.TrimSpace(v)
	if v == "" {
		return ','
	}
	switch strings.ToLower(v) {
	case `\\t`, `tab`, `\t`:
		return '\t'
	case `;`:
		return ';'
	case `|`:
		return '|'
	default:
		runes := []rune(v)
		if len(runes) == 1 {
			return runes[0]
		}
		return ','
	}
}

func parseBoolDefaultPreview(v string, def bool) bool {
	v = strings.TrimSpace(strings.ToLower(v))
	if v == "" {
		return def
	}
	return v == "true" || v == "1" || v == "yes" || v == "oui"
}

func parseIntDefaultPreview(v string, def int) int {
	v = strings.TrimSpace(v)
	if v == "" {
		return def
	}
	var n int
	_, err := fmt.Sscanf(v, "%d", &n)
	if err != nil {
		return def
	}
	return n
}

func parseHeaderListPreview(v string) []string {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for i, p := range parts {
		name := strings.TrimSpace(p)
		if name == "" {
			name = fmt.Sprintf("column_%d", i+1)
		}
		out = append(out, name)
	}
	return out
}

func normalizeEncodingPreview(v string) string {
	v = strings.TrimSpace(strings.ToLower(v))
	if v == "" {
		return "utf-8"
	}
	switch v {
	case "utf8":
		return "utf-8"
	case "latin1":
		return "iso-8859-1"
	default:
		return v
	}
}

func decoderForEncodingPreview(enc string) transform.Transformer {
	switch enc {
	case "utf-8", "utf8":
		return nil
	case "utf-16le":
		return unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder()
	case "utf-16be":
		return unicode.UTF16(unicode.BigEndian, unicode.UseBOM).NewDecoder()
	case "windows-1252", "cp1252":
		return charmap.Windows1252.NewDecoder()
	case "iso-8859-1", "latin1":
		return charmap.ISO8859_1.NewDecoder()
	default:
		return nil
	}
}

type crToLFReaderPreview struct {
	r *bufio.Reader
}

func newCRToLFReaderPreview(r *bufio.Reader) io.Reader {
	return &crToLFReaderPreview{r: r}
}

func (r *crToLFReaderPreview) Read(p []byte) (int, error) {
	n := 0
	for n < len(p) {
		b, err := r.r.ReadByte()
		if err != nil {
			if err == io.EOF && n > 0 {
				return n, nil
			}
			return n, err
		}
		if b == '\r' {
			p[n] = '\n'
			n++
			continue
		}
		p[n] = b
		n++
	}
	return n, nil
}
