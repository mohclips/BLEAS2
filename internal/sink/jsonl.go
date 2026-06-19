// Package sink contains writers that persist captured advertisements.
package sink

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Rotation strategies for the JSONL sink.
const (
	RotateDaily = "daily"
	RotateSize  = "size"
	RotateNone  = "none"
)

// JSONLConfig configures the JSONL writer.
type JSONLConfig struct {
	Dir           string // directory to write files into; created if missing
	Rotation      string // RotateDaily | RotateSize | RotateNone
	MaxSize       int64  // bytes; only used when Rotation == RotateSize
	CompressAfter int    // gzip files older than N days at rotation time; 0 = off
}

// JSONL is a goroutine-safe writer that appends one JSON record per line.
// Files are named `bleas-YYYY-MM-DD.jsonl` (date rotation) or
// `bleas-YYYY-MM-DDTHH-MM-SS.jsonl` (size rotation).
type JSONL struct {
	cfg JSONLConfig

	mu       sync.Mutex
	cur      *os.File
	curName  string
	curDate  string
	curSize  int64
}

// NewJSONL prepares the output directory and returns a writer.
func NewJSONL(cfg JSONLConfig) (*JSONL, error) {
	if cfg.Dir == "" {
		return nil, fmt.Errorf("jsonl: dir is required")
	}
	if cfg.Rotation == "" {
		cfg.Rotation = RotateDaily
	}
	if err := os.MkdirAll(cfg.Dir, 0o755); err != nil {
		return nil, fmt.Errorf("jsonl: mkdir %s: %w", cfg.Dir, err)
	}
	return &JSONL{cfg: cfg}, nil
}

// Write appends record as a JSON line. record must be valid JSON without a
// trailing newline; one is added.
func (j *JSONL) Write(record []byte) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	if err := j.rotateIfNeeded(); err != nil {
		return err
	}

	n, err := j.cur.Write(record)
	if err == nil && (len(record) == 0 || record[len(record)-1] != '\n') {
		var m int
		m, err = j.cur.Write([]byte{'\n'})
		n += m
	}
	j.curSize += int64(n)
	return err
}

// Close flushes and closes the current file.
func (j *JSONL) Close() error {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.cur == nil {
		return nil
	}
	err := j.cur.Close()
	j.cur = nil
	return err
}

// rotateIfNeeded opens a new file when the date or size threshold rolls over.
// Caller must hold j.mu.
func (j *JSONL) rotateIfNeeded() error {
	today := time.Now().Format("2006-01-02")
	if j.cur != nil {
		switch j.cfg.Rotation {
		case RotateNone:
			return nil
		case RotateDaily:
			if today == j.curDate {
				return nil
			}
		case RotateSize:
			if j.cfg.MaxSize <= 0 || j.curSize < j.cfg.MaxSize {
				return nil
			}
		}
	}

	if j.cur != nil {
		_ = j.cur.Close()
		j.cur = nil
	}

	// Always include the open-time in the filename so distinct runs (and
	// post-rotation files) get unique names — never overwrites or
	// silently appends to a file from a previous run.
	name := fmt.Sprintf("bleas-%s.jsonl", time.Now().UTC().Format("2006-01-02T15-04-05Z"))
	path := filepath.Join(j.cfg.Dir, name)

	// O_EXCL guarantees we never reopen an existing file; if two openings
	// land in the same second (rare), bump to the next available name.
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o644)
	for n := 1; os.IsExist(err) && n < 100; n++ {
		path = filepath.Join(j.cfg.Dir, fmt.Sprintf("bleas-%s.%d.jsonl",
			time.Now().UTC().Format("2006-01-02T15-04-05Z"), n))
		f, err = os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o644)
	}
	if err != nil {
		return fmt.Errorf("jsonl: open %s: %w", path, err)
	}
	j.cur = f
	j.curName = path
	j.curDate = today
	j.curSize = 0 // fresh file (guaranteed by O_EXCL)

	// Best-effort gzip of old files; never fatal.
	if j.cfg.CompressAfter > 0 {
		go j.compressOld(j.cfg.CompressAfter)
	}
	return nil
}

// compressOld gzips any *.jsonl files older than olderDays. Skips the currently
// open file. Best-effort: logs nothing, errors swallowed (sink mustn't crash
// the scanner).
func (j *JSONL) compressOld(olderDays int) {
	cutoff := time.Now().Add(-time.Duration(olderDays) * 24 * time.Hour)
	entries, err := os.ReadDir(j.cfg.Dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if filepath.Ext(name) != ".jsonl" {
			continue
		}
		full := filepath.Join(j.cfg.Dir, name)
		if full == j.curName {
			continue
		}
		info, err := e.Info()
		if err != nil || info.ModTime().After(cutoff) {
			continue
		}
		_ = gzipFile(full)
	}
}

func gzipFile(path string) error {
	in, err := os.Open(path)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(path+".gz", os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}

	gz := gzip.NewWriter(out)
	if _, err := io.Copy(gz, in); err != nil {
		_ = gz.Close()
		_ = out.Close()
		_ = os.Remove(path + ".gz")
		return err
	}
	if err := gz.Close(); err != nil {
		_ = out.Close()
		_ = os.Remove(path + ".gz")
		return err
	}
	if err := out.Close(); err != nil {
		_ = os.Remove(path + ".gz")
		return err
	}
	return os.Remove(path)
}
