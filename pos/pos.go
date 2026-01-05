package pos

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/k1LoW/go-github-actions/artifact"
)

var minTimeOffset = -18 * time.Hour

const (
	posTypeFile      = "file"
	posTypeArtificat = "artifact"

	posFilePrefix        = "tailor-log.pos"
	posArtifactKeyPrefix = "tailor-log-pos"
)

type Pos struct {
	m           sync.Map
	workspaceID string
	minTime     time.Time
}

func New(workspaceID string) *Pos {
	return &Pos{
		workspaceID: workspaceID,
		minTime:     time.Now().Add(minTimeOffset),
	}
}

func At(workspaceID string, t time.Time) *Pos {
	return &Pos{
		workspaceID: workspaceID,
		minTime:     t,
	}
}

func (p *Pos) Store(key string, value time.Time) {
	p.m.Store(key, value)
}

func (p *Pos) Load(key string) time.Time {
	if v, ok := p.m.Load(key); ok {
		t, ok := v.(time.Time)
		if !ok {
			return p.minTime
		}
		// If the stored time is older than minTime, use minTime instead
		if t.Before(p.minTime) {
			return p.minTime
		}
		return t
	}
	return p.minTime
}

func RestoreFrom(ctx context.Context, posType, workspaceID string) (*Pos, error) {
	posFileName := fmt.Sprintf("%s.%s.json", posFilePrefix, workspaceID)
	switch posType {
	case posTypeFile:
		b, err := os.ReadFile(posFileName)
		if err != nil {
			if os.IsNotExist(err) {
				slog.Info("Position file does not exist, starting from default position", "file", posFileName)
				return New(workspaceID), nil
			}
			return nil, err
		}
		slog.Info("Restored position from file", "file", posFileName)
		return Restore(workspaceID, b)
	case posTypeArtificat:
		posArtifactKey := fmt.Sprintf("%s-%s", posArtifactKeyPrefix, workspaceID)
		ownerrepo := strings.Split(os.Getenv("GITHUB_REPOSITORY"), "/")
		if len(ownerrepo) != 2 {
			return nil, fmt.Errorf("invalid GITHUB_REPOSITORY: %s", os.Getenv("GITHUB_REPOSITORY"))
		}
		owner := ownerrepo[0]
		repo := ownerrepo[1]
		b, err := fetchLatestArtifact(ctx, owner, repo, posArtifactKey, posFileName)
		if err != nil {
			if errors.Is(err, ErrArtifactNotFound) {
				slog.Info("Position artifact does not exist, starting from default position", "key", posArtifactKey)
				return New(workspaceID), nil
			}
			return nil, err
		}
		pos, err := Restore(workspaceID, b)
		if err != nil {
			return nil, err
		}
		count := 0
		pos.m.Range(func(key, value any) bool {
			count++
			return true
		})
		slog.Info("Restored position from artifact", "key", posArtifactKey, "pos_count", count)
		return pos, nil
	default:
		return nil, errors.New("unknown pos type: " + posType)
	}
}

func Restore(workspaceID string, b []byte) (*Pos, error) {
	return RestoreAt(workspaceID, b, time.Now())
}

func RestoreAt(workspaceID string, b []byte, now time.Time) (*Pos, error) {
	m := map[string]time.Time{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	p := At(workspaceID, now.Add(minTimeOffset))
	for k, v := range m {
		p.Store(k, v)
	}
	return p, nil
}

func (p *Pos) DumpTo(ctx context.Context, posType string) error {
	posFileName := fmt.Sprintf("%s.%s.json", posFilePrefix, p.workspaceID)
	switch posType {
	case posTypeFile:
		b, err := p.Dump()
		if err != nil {
			return err
		}
		slog.Info("Dumped position to file", "file", posFileName)
		return os.WriteFile(posFileName, b, 0600)
	case posTypeArtificat:
		posArtifactKey := fmt.Sprintf("%s-%s", posArtifactKeyPrefix, p.workspaceID)
		b, err := p.Dump()
		if err != nil {
			return err
		}
		slog.Info("Dumped position to file", "file", posFileName)
		if err := artifact.Upload(ctx, posArtifactKey, posFileName, bytes.NewReader(b)); err != nil {
			return err
		}
		slog.Info("Uploaded position to artifact", "key", posArtifactKey)
		return nil
	default:
		return errors.New("unknown pos type: " + posType)
	}
}

func (p *Pos) Dump() ([]byte, error) {
	m := map[string]time.Time{}
	p.m.Range(func(key, value any) bool {
		t, ok := value.(time.Time)
		if !ok {
			return false
		}
		k, ok := key.(string)
		if !ok {
			return false
		}
		m[k] = t
		return true
	})
	return json.Marshal(m)
}
