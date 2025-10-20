package pos

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/go-github/v75/github"
	"github.com/k1LoW/go-github-client/v75/factory"
)

const maxCopySize = 1073741824 // 1GB

var ErrArtifactNotFound = errors.New("artifact not found")

func fetchLatestArtifact(ctx context.Context, owner, repo, name, fp string) ([]byte, error) {
	client, err := factory.NewGithubClient(factory.Timeout(10 * time.Second))
	if err != nil {
		return nil, err
	}
	const maxRedirect = 5
	page := 1
	for {
		l, res, err := client.Actions.ListArtifacts(ctx, owner, repo, &github.ListArtifactsOptions{
			Name: &name,
			ListOptions: github.ListOptions{
				Page:    page,
				PerPage: 100,
			},
		})
		if err != nil {
			return nil, err
		}
		page += 1
		for _, a := range l.Artifacts {
			u, _, err := client.Actions.DownloadArtifact(ctx, owner, repo, a.GetID(), maxRedirect)
			if err != nil {
				return nil, err
			}
			resp, err := http.Get(u.String())
			if err != nil {
				return nil, err
			}
			buf := new(bytes.Buffer)
			size, err := io.CopyN(buf, resp.Body, maxCopySize)
			if !errors.Is(err, io.EOF) {
				return nil, err
			}
			if size >= maxCopySize {
				return nil, fmt.Errorf("too large file size to copy: %d >= %d", size, maxCopySize)
			}
			reader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
			if err != nil {
				return nil, err
			}
			for _, file := range reader.File {
				if file.Name != fp {
					continue
				}
				in, err := file.Open()
				if err != nil {
					return nil, err
				}
				out := new(bytes.Buffer)
				size, err := io.CopyN(out, in, maxCopySize)
				if !errors.Is(err, io.EOF) {
					_ = in.Close() //nostyle:handlerrors
					return nil, err
				}
				if size >= maxCopySize {
					_ = in.Close() //nostyle:handlerrors
					return nil, fmt.Errorf("too large file size to copy: %d >= %d", size, maxCopySize)
				}
				if err := in.Close(); err != nil {
					return nil, err
				}
				return out.Bytes(), nil
			}
		}
		if res.NextPage == 0 {
			break
		}
	}
	return nil, ErrArtifactNotFound
}
