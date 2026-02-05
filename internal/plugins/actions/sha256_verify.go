package actions

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	coreerr "opskit/internal/core/errors"
	"opskit/internal/schema"
)

type sha256VerifyAction struct{}

func (a *sha256VerifyAction) Kind() string { return "sha256_verify" }

func (a *sha256VerifyAction) Run(_ context.Context, req Request) (Result, error) {
	filePath := toString(req.Params["file"], "")
	expected := strings.ToLower(toString(req.Params["expected_sha256"], ""))
	if filePath == "" || expected == "" {
		return Result{}, fmt.Errorf("sha256_verify requires params.file and params.expected_sha256")
	}

	f, err := os.Open(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Result{}, fmt.Errorf("%w: package file not found: %s", coreerr.ErrPreconditionFailed, filePath)
		}
		return Result{}, err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return Result{}, err
	}
	actual := fmt.Sprintf("%x", h.Sum(nil))
	if actual != expected {
		issue := &schema.Issue{ID: req.ID, Severity: schema.SeverityFail, Message: "sha256 mismatch", Advice: "actual=" + actual}
		return Result{ActionID: req.ID, Status: schema.StatusFailed, Severity: schema.SeverityFail, Message: issue.Message, Issue: issue}, nil
	}

	return Result{ActionID: req.ID, Status: schema.StatusPassed, Severity: schema.SeverityInfo, Message: "sha256 verified"}, nil
}
