package handlers

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

type csvScanResponse struct {
	Success bool        `json:"success"`
	Scan    csvScanMeta `json:"scan"`
	Error   string      `json:"error,omitempty"`
}

type csvScanMeta struct {
	Path              string     `json:"path"`
	Encoding          string     `json:"encoding"`
	Delimiter         string     `json:"delimiter"`
	Newline           string     `json:"newline"`
	HasHeader         bool       `json:"hasHeader"`
	Headers           []string   `json:"headers"`
	DetectedColumns   int        `json:"detectedColumns"`
	SampleLines       []string   `json:"sampleLines"`
	Confidence        string     `json:"confidence"`
	SuggestedParams   scanParams `json:"suggestedParams"`
	Warnings          []string   `json:"warnings"`
}

type scanParams struct {
	Encoding         string `json:"encoding"`
	Delimiter        string `json:"delimiter"`
	Newline          string `json:"newline"`
	HasHeader        string `json:"has_header"`
	Headers          string `json:"headers"`
	LazyQuotes       string `json:"lazy_quotes"`
	TrimLeadingSpace string `json:"trim_leading_space"`
	SkipEmptyLines   string `json:"skip_empty_lines"`
	FieldsPerRecord  string `json:"fields_per_record"`
}

func (h *ProjectHandler) CSVScan(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "projectID")

	path := strings.TrimSpace(r.URL.Query().Get("path"))
	if path == "" {
		writeJSON(w, http.StatusBadRequest, csvScanResponse{Success: false, Error: "parametre 'path' manquant"})
		return
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, csvScanResponse{Success: false, Error: fmt.Sprintf("lecture fichier: %v", err)})
		return
	}
	if len(raw) == 0 {
		writeJSON(w, http.StatusBadRequest, csvScanResponse{Success: false, Error: "fichier vide"})
		return
	}

	enc := detectEncoding(raw)
	newline := detectNewline(raw)
	decoded, err := decodeBytes(raw, enc)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, csvScanResponse{Success: false, Error: fmt.Sprintf("decodage fichier: %v", err)})
		return
	}

	sampleLines := firstNonEmptyLines(decoded, 5)
	delimiter, confidence := detectDelimiter(sampleLines)
	parsedRows, err := parseSampleRows(decoded, delimiter, newline, 5)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, csvScanResponse{Success: false, Error: fmt.Sprintf("analyse csv: %v", err)})
		return
	}
	if len(parsedRows) == 0 {
		writeJSON(w, http.StatusBadRequest, csvScanResponse{Success: false, Error: "aucune ligne exploitable detectee"})
		return
	}

	detectedColumns := maxColumns(parsedRows)
	hasHeader := detectHasHeader(parsedRows)
	headers := buildHeaders(parsedRows, hasHeader, detectedColumns)

	warnings := make([]string, 0)
	if confidence != "high" {
		warnings = append(warnings, "Detection du separateur avec confiance moyenne, verifier le delimiteur si le rendu semble incorrect.")
	}
	if !hasHeader {
		warnings = append(warnings, "Aucune en-tete claire detectee, noms de colonnes generes automatiquement.")
	}

	writeJSON(w, http.StatusOK, csvScanResponse{
		Success: true,
		Scan: csvScanMeta{
			Path:            path,
			Encoding:        enc,
			Delimiter:       string(delimiter),
			Newline:         newline,
			HasHeader:       hasHeader,
			Headers:         headers,
			DetectedColumns: detectedColumns,
			SampleLines:     sampleLines,
			Confidence:      confidence,
			SuggestedParams: scanParams{
				Encoding:         enc,
				Delimiter:        string(delimiter),
				Newline:          newline,
				HasHeader:        boolString(hasHeader),
				Headers:          strings.Join(headers, ","),
				LazyQuotes:       "true",
				TrimLeadingSpace: "true",
				SkipEmptyLines:   "true",
				FieldsPerRecord:  "-1",
			},
			Warnings: warnings,
		},
	})
}

func detectEncoding(raw []byte) string {
	if len(raw) >= 2 {
		if raw[0] == 0xFF && raw[1] == 0xFE {
			return "utf-16le"
		}
		if raw[0] == 0xFE && raw[1] == 0xFF {
			return "utf-16be"
		}
	}
	if len(raw) >= 3 && raw[0] == 0xEF && raw[1] == 0xBB && raw[2] == 0xBF {
		return "utf-8"
	}
	if utf8.Valid(raw) {
		return "utf-8"
	}
	return "windows-1252"
}

func detectNewline(raw []byte) string {
	if bytes.Contains(raw, []byte("\r\n")) || bytes.Contains(raw, []byte("\n")) {
		return "auto"
	}
	if bytes.Contains(raw, []byte("\r")) {
		return "cr"
	}
	return "auto"
}

func decodeBytes(raw []byte, enc string) (string, error) {
	var reader io.Reader = bytes.NewReader(raw)
	if decoder := decoderForScanEncoding(enc); decoder != nil {
		reader = transform.NewReader(reader, decoder)
	}
	buf, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func decoderForScanEncoding(enc string) transform.Transformer {
	switch strings.ToLower(strings.TrimSpace(enc)) {
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

func firstNonEmptyLines(content string, limit int) []string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	parts := strings.Split(content, "\n")
	out := make([]string, 0, limit)
	for _, part := range parts {
		line := strings.TrimSpace(part)
		if line == "" {
			continue
		}
		out = append(out, line)
		if len(out) >= limit {
			break
		}
	}
	return out
}

func detectDelimiter(lines []string) (rune, string) {
	candidates := []rune{',', ';', '|', '\t'}
	best := ','
	bestScore := -1
	bestConsistency := -1
	for _, cand := range candidates {
		counts := make([]int, 0, len(lines))
		for _, line := range lines {
			counts = append(counts, countFieldsForDelimiter(line, cand))
		}
		if len(counts) == 0 {
			continue
		}
		minC, maxC := counts[0], counts[0]
		for _, c := range counts[1:] {
			if c < minC {
				minC = c
			}
			if c > maxC {
				maxC = c
			}
		}
		score := 0
		for _, c := range counts {
			score += c
		}
		consistency := maxC - minC
		if score > bestScore || (score == bestScore && consistency < bestConsistency) {
			best = cand
			bestScore = score
			bestConsistency = consistency
		}
	}
	confidence := "medium"
	if bestScore >= len(lines)*3 && bestConsistency <= 1 {
		confidence = "high"
		} else if bestScore <= len(lines) {
		confidence = "low"
	}
	return best, confidence
}

func countFieldsForDelimiter(line string, delimiter rune) int {
	r := csv.NewReader(strings.NewReader(line))
	r.Comma = delimiter
	r.LazyQuotes = true
	rec, err := r.Read()
	if err != nil {
		return 0
	}
	return len(rec)
}

func parseSampleRows(content string, delimiter rune, newline string, limit int) ([][]string, error) {
	reader := bufio.NewReader(strings.NewReader(content))
	var rr io.Reader = reader
	if newline == "cr" {
		rr = newCRToLFReaderPreview(reader)
	}
	csvReader := csv.NewReader(rr)
	csvReader.Comma = delimiter
	csvReader.LazyQuotes = true
	csvReader.TrimLeadingSpace = true
	rows := make([][]string, 0, limit)
	for len(rows) < limit {
		rec, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if isEmptyScanRecord(rec) {
			continue
		}
		rows = append(rows, rec)
	}
	return rows, nil
}

func isEmptyScanRecord(rec []string) bool {
	for _, v := range rec {
		if strings.TrimSpace(v) != "" {
			return false
		}
	}
	return true
}

func maxColumns(rows [][]string) int {
	max := 0
	for _, row := range rows {
		if len(row) > max {
			max = len(row)
		}
	}
	return max
}

func detectHasHeader(rows [][]string) bool {
	if len(rows) < 2 {
		return true
	}
	first := rows[0]
	second := rows[1]
	if len(first) == 0 {
		return false
	}
	headerish := 0
	for i := 0; i < len(first); i++ {
		v1 := strings.TrimSpace(first[i])
		v2 := ""
		if i < len(second) {
			v2 = strings.TrimSpace(second[i])
		}
		if looksLikeHeaderValue(v1) && !looksLikeHeaderValue(v2) {
			headerish++
		}
	}
	return headerish >= max(1, len(first)/2)
}

func looksLikeHeaderValue(v string) bool {
	if v == "" {
		return false
	}
	lower := strings.ToLower(v)
	if strings.Contains(lower, " ") || strings.Contains(lower, "_") {
		return true
	}
	for _, r := range lower {
		if (r >= 'a' && r <= 'z') || r == '_' {
			return true
		}
	}
	return false
}

func buildHeaders(rows [][]string, hasHeader bool, detectedColumns int) []string {
	if detectedColumns <= 0 {
		return nil
	}
	if hasHeader && len(rows) > 0 {
		headers := make([]string, detectedColumns)
		for i := 0; i < detectedColumns; i++ {
			if i < len(rows[0]) {
				h := strings.TrimSpace(rows[0][i])
				if h != "" {
					headers[i] = h
					continue
				}
			}
			headers[i] = fmt.Sprintf("column_%d", i+1)
		}
		return headers
	}
	headers := make([]string, detectedColumns)
	for i := 0; i < detectedColumns; i++ {
		headers[i] = fmt.Sprintf("column_%d", i+1)
	}
	return headers
}

func boolString(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
