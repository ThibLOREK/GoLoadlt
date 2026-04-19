package sources

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

func init() {
	blocks.Register("source.csv", func() contracts.Block { return &CSVSource{} })
}

// CSVSource lit un fichier CSV et envoie chaque ligne vers le(s) port(s) de sortie.
type CSVSource struct{}

func (b *CSVSource) Type() string { return "source.csv" }

func (b *CSVSource) Run(bctx *contracts.BlockContext) error {
	path := strings.TrimSpace(bctx.Params["path"])
	if path == "" {
		return fmt.Errorf("source.csv: paramètre 'path' manquant")
	}

	if len(bctx.Outputs) == 0 {
		return fmt.Errorf("source.csv: aucun port de sortie connecté")
	}

	reader, closer, err := openCSVReader(path, bctx.Params)
	if err != nil {
		return err
	}
	defer closer()

	headers, err := readHeaders(reader, bctx.Params)
	if err != nil {
		return err
	}

	for {
		select {
		case <-bctx.Ctx.Done():
			for _, out := range bctx.Outputs {
				close(out.Ch)
			}
			return bctx.Ctx.Err()
		default:
		}

		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			for _, out := range bctx.Outputs {
				close(out.Ch)
			}
			return fmt.Errorf("source.csv: lecture ligne: %w", err)
		}

		if isEffectivelyEmpty(record) && strings.EqualFold(strings.TrimSpace(bctx.Params["skip_empty_lines"]), "true") {
			continue
		}

		row := make(contracts.DataRow, len(headers))
		for i, h := range headers {
			if i < len(record) {
				row[h] = record[i]
			} else {
				row[h] = ""
			}
		}

		for _, out := range bctx.Outputs {
			select {
			case out.Ch <- row:
			case <-bctx.Ctx.Done():
				for _, o := range bctx.Outputs {
					close(o.Ch)
				}
				return bctx.Ctx.Err()
			}
		}
	}

	for _, out := range bctx.Outputs {
		close(out.Ch)
	}
	return nil
}

func openCSVReader(path string, params map[string]string) (*csv.Reader, func(), error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("source.csv: ouverture fichier '%s': %w", path, err)
	}

	var r io.Reader = f
	encodingName := normalizeEncoding(params["encoding"])
	if decoder := decoderForEncoding(encodingName); decoder != nil {
		r = transform.NewReader(r, decoder)
	}

	br := bufio.NewReader(r)

	newline := strings.ToLower(strings.TrimSpace(params["newline"]))
	if newline == "cr" {
		r = newCRToLFReader(br)
		br = bufio.NewReader(r)
	}

	reader := csv.NewReader(br)
	reader.Comma = parseDelimiter(params["delimiter"])
	reader.LazyQuotes = parseBoolDefault(params["lazy_quotes"], true)
	reader.TrimLeadingSpace = parseBoolDefault(params["trim_leading_space"], true)
	reader.FieldsPerRecord = parseIntDefault(params["fields_per_record"], -1)
	reader.ReuseRecord = false

	return reader, func() { _ = f.Close() }, nil
}

func readHeaders(reader *csv.Reader, params map[string]string) ([]string, error) {
	hasHeader := parseBoolDefault(params["has_header"], true)
	customHeaders := parseHeaderList(params["headers"])

	if hasHeader {
		headers, err := reader.Read()
		if err != nil {
			return nil, fmt.Errorf("source.csv: lecture en-tête: %w", err)
		}
		for i := range headers {
			headers[i] = strings.TrimSpace(headers[i])
			if headers[i] == "" {
				headers[i] = fmt.Sprintf("column_%d", i+1)
			}
		}
		return headers, nil
	}

	if len(customHeaders) > 0 {
		return customHeaders, nil
	}

	return nil, fmt.Errorf("source.csv: sans en-tête, le paramètre 'headers' est requis (ex: id,name,amount)")
}

func parseDelimiter(v string) rune {
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

func parseBoolDefault(v string, def bool) bool {
	v = strings.TrimSpace(strings.ToLower(v))
	if v == "" {
		return def
	}
	return v == "true" || v == "1" || v == "yes" || v == "oui"
}

func parseIntDefault(v string, def int) int {
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

func parseHeaderList(v string) []string {
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

func normalizeEncoding(v string) string {
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

func decoderForEncoding(enc string) transform.Transformer {
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

func isEffectivelyEmpty(record []string) bool {
	for _, v := range record {
		if strings.TrimSpace(v) != "" {
			return false
		}
	}
	return true
}

type crToLFReader struct {
	r *bufio.Reader
}

func newCRToLFReader(r *bufio.Reader) io.Reader {
	return &crToLFReader{r: r}
}

func (r *crToLFReader) Read(p []byte) (int, error) {
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
