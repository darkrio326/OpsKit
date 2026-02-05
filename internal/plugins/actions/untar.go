package actions

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	coreerr "opskit/internal/core/errors"
	"opskit/internal/schema"
)

type untarAction struct{}

func (a *untarAction) Kind() string { return "untar" }

func (a *untarAction) Run(_ context.Context, req Request) (Result, error) {
	src := toString(req.Params["src"], "")
	dest := toString(req.Params["dest"], "")
	if src == "" || dest == "" {
		return Result{}, fmt.Errorf("untar requires params.src and params.dest")
	}

	f, err := os.Open(src)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Result{}, fmt.Errorf("%w: package file not found: %s", coreerr.ErrPreconditionFailed, src)
		}
		issue := &schema.Issue{ID: req.ID, Severity: schema.SeverityFail, Message: "open tar failed: " + err.Error()}
		return Result{ActionID: req.ID, Status: schema.StatusFailed, Severity: schema.SeverityFail, Message: issue.Message, Issue: issue}, nil
	}
	defer f.Close()

	var tr *tar.Reader
	if strings.HasSuffix(src, ".gz") || strings.HasSuffix(src, ".tgz") {
		gz, gzErr := gzip.NewReader(f)
		if gzErr != nil {
			return Result{}, gzErr
		}
		defer gz.Close()
		tr = tar.NewReader(gz)
	} else {
		tr = tar.NewReader(f)
	}

	if err := os.MkdirAll(dest, 0o755); err != nil {
		return Result{}, err
	}

	count := 0
	for {
		hdr, rErr := tr.Next()
		if rErr == io.EOF {
			break
		}
		if rErr != nil {
			return Result{}, rErr
		}

		target := filepath.Join(dest, hdr.Name)
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(dest)+string(os.PathSeparator)) && filepath.Clean(target) != filepath.Clean(dest) {
			return Result{}, fmt.Errorf("unsafe tar path: %s", hdr.Name)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(hdr.Mode)); err != nil {
				return Result{}, err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return Result{}, err
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				return Result{}, err
			}
			if _, err := io.Copy(out, tr); err != nil {
				_ = out.Close()
				return Result{}, err
			}
			_ = out.Close()
			count++
		}
	}

	return Result{ActionID: req.ID, Status: schema.StatusPassed, Severity: schema.SeverityInfo, Message: fmt.Sprintf("untar completed: %d files", count), Metrics: []schema.Metric{{Label: "untar_files", Value: fmt.Sprintf("%d", count)}}}, nil
}
