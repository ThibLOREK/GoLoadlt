package xmlstore

import (
	"context"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Store définit le contrat de persistance des pipelines XML.
type Store interface {
	Save(ctx context.Context, p *XMLPipeline) error
	Load(ctx context.Context, id string) (*XMLPipeline, error)
	List(ctx context.Context) ([]XMLPipeline, error)
	Delete(ctx context.Context, id string) error
}

// FileStore est une implémentation de Store basée sur le système de fichiers.
type FileStore struct {
	baseDir string
}

// NewFileStore crée un FileStore qui persiste les pipelines dans baseDir.
// Le répertoire est créé s'il n'existe pas.
func NewFileStore(baseDir string) (*FileStore, error) {
	if err := os.MkdirAll(baseDir, 0o750); err != nil {
		return nil, fmt.Errorf("xmlstore.NewFileStore: %w", err)
	}
	return &FileStore{baseDir: baseDir}, nil
}

// Save sérialise p en XML et l'écrit de manière atomique (tmp + rename).
func (fs *FileStore) Save(_ context.Context, p *XMLPipeline) error {
	data, err := xml.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("xmlstore.Save %s: marshal: %w", p.ID, err)
	}

	// Ajout de la déclaration XML standard.
	content := []byte(xml.Header)
	content = append(content, data...)
	content = append(content, '\n')

	// Écriture atomique : fichier temporaire dans le même répertoire,
	// puis os.Rename garantit l'atomicité sur les systèmes POSIX.
	tmpFile, err := os.CreateTemp(fs.baseDir, ".tmp-"+p.ID+"-")
	if err != nil {
		return fmt.Errorf("xmlstore.Save %s: create tmp: %w", p.ID, err)
	}
	tmpPath := tmpFile.Name()

	if _, err = tmpFile.Write(content); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
		return fmt.Errorf("xmlstore.Save %s: write tmp: %w", p.ID, err)
	}
	if err = tmpFile.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("xmlstore.Save %s: close tmp: %w", p.ID, err)
	}

	dest := filepath.Join(fs.baseDir, p.ID+".xml")
	if err = os.Rename(tmpPath, dest); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("xmlstore.Save %s: rename: %w", p.ID, err)
	}
	return nil
}

// Load lit et désérialise le pipeline identifié par id.
func (fs *FileStore) Load(_ context.Context, id string) (*XMLPipeline, error) {
	path := filepath.Join(fs.baseDir, id+".xml")
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("xmlstore.Load %s: %w", id, err)
	}
	defer f.Close()

	var p XMLPipeline
	if err = xml.NewDecoder(f).Decode(&p); err != nil {
		return nil, fmt.Errorf("xmlstore.Load %s: decode: %w", id, err)
	}
	return &p, nil
}

// List retourne tous les pipelines présents dans baseDir.
func (fs *FileStore) List(_ context.Context) ([]XMLPipeline, error) {
	entries, err := os.ReadDir(fs.baseDir)
	if err != nil {
		return nil, fmt.Errorf("xmlstore.List: %w", err)
	}

	var pipelines []XMLPipeline
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".xml") {
			continue
		}
		path := filepath.Join(fs.baseDir, e.Name())
		f, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("xmlstore.List: open %s: %w", e.Name(), err)
		}
		var p XMLPipeline
		decErr := xml.NewDecoder(f).Decode(&p)
		_ = f.Close()
		if decErr != nil {
			return nil, fmt.Errorf("xmlstore.List: decode %s: %w", e.Name(), decErr)
		}
		pipelines = append(pipelines, p)
	}
	if pipelines == nil {
		pipelines = []XMLPipeline{}
	}
	return pipelines, nil
}

// Delete supprime le fichier XML associé à id.
func (fs *FileStore) Delete(_ context.Context, id string) error {
	path := filepath.Join(fs.baseDir, id+".xml")
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("xmlstore.Delete %s: %w", id, err)
	}
	return nil
}
