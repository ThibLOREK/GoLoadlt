package sources

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rinjold/go-etl-studio/internal/etl/blocks"
	"github.com/rinjold/go-etl-studio/internal/etl/contracts"
)

func init() {
	blocks.Register("source.directory", func() contracts.Block { return &Directory{} })
}

// Directory liste les fichiers d'un répertoire et émet une ligne par fichier.
// Colonnes émises : name, path, size, extension, modifiedAt
type Directory struct{}

func (b *Directory) Type() string { return "source.directory" }

func (b *Directory) Run(bctx *contracts.BlockContext) error {
	dir := bctx.Params["path"]
	if dir == "" {
		return fmt.Errorf("source.directory: paramètre 'path' manquant")
	}
	pattern := bctx.Params["pattern"] // ex: "*.csv" (optionnel)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("source.directory: lecture répertoire '%s': %w", dir, err)
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if pattern != "" {
			matched, _ := filepath.Match(pattern, e.Name())
			if !matched {
				continue
			}
		}
		info, _ := e.Info()
		fullPath := filepath.Join(dir, e.Name())
		row := contracts.DataRow{
			"name":       e.Name(),
			"path":       fullPath,
			"extension":  filepath.Ext(e.Name()),
			"size":       info.Size(),
			"modifiedAt": info.ModTime().Format("2006-01-02T15:04:05"),
		}
		for _, out := range bctx.Outputs {
			select {
			case out.Ch <- row:
			case <-bctx.Ctx.Done():
				for _, o := range bctx.Outputs { close(o.Ch) }
				return bctx.Ctx.Err()
			}
		}
	}
	for _, out := range bctx.Outputs { close(out.Ch) }
	return nil
}