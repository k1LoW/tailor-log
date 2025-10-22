/*
Copyright © 2025 Ken'ichiro Oyama <k1lowxb@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"log/slog"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/k1LoW/donegroup"
	"github.com/k1LoW/errors"
	"github.com/k1LoW/tailor-log/config"
	"github.com/k1LoW/tailor-log/datadog"
	"github.com/k1LoW/tailor-log/item"
	"github.com/k1LoW/tailor-log/pos"
	"github.com/k1LoW/tailor-log/tailor"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

const waitTimeout = 60 * time.Second

var (
	posType        string
	datadogService string
	datadogTags    []string
)

var ingestCmd = &cobra.Command{
	Use:   "ingest",
	Short: "Ingest logs from Tailor Plaform to Datadog",
	Long:  `Ingest logs from Tailor Plaform to Datadog.`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		ctx, _ := donegroup.WithCancel(cmd.Context())
		ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
		p, err := pos.RestoreFrom(ctx, posType, workspaceID)
		if err != nil {
			return err
		}
		var sendErr error
		out := make(chan *item.Item)
		defer func() {
			cancel()
			close(out)
			if errr := donegroup.WaitWithTimeout(ctx, waitTimeout); errr != nil {
				err = errors.Join(err, errr)
			}
			if sendErr != nil {
				slog.Error("Skip saving position due to send error", "error", sendErr)
				return
			}
			if errr := p.DumpTo(cmd.Context(), posType); errr != nil {
				err = errors.Join(err, errr)
			}
		}()
		go func() {
			// Memory usage logging
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					var m runtime.MemStats
					runtime.ReadMemStats(&m)
					slog.Info("Memory usage", slog.Int64("alloc(MB)", int64(m.Alloc)/1024/1024), slog.Int64("total_alloc(MB)", int64(m.TotalAlloc)/1024/1024), slog.Int64("sys(MB)", int64(m.Sys)/1024/1024), slog.Int64("num_gc", int64(m.NumGC)))
				case <-ctx.Done():
					return
				}
			}
		}()

		cfg := &config.Config{}
		cfg.WorkspaceID = workspaceID
		var splitted []string
		for _, input := range inputs {
			splitted = append(splitted, strings.Split(input, ",")...)
		}
		cfg.Inputs = splitted
		cfg.Outputs.Datadog.Service = datadogService
		var tags []string
		for _, tag := range datadogTags {
			tags = append(tags, strings.Split(tag, ",")...)
		}
		cfg.Outputs.Datadog.Tags = tags
		c, err := tailor.New(cfg)
		if err != nil {
			return err
		}
		dd, err := datadog.New(cfg)
		if err != nil {
			return err
		}
		donegroup.Go(ctx, func() error {
			err := dd.SendLogs(ctx, out)
			if err != nil {
				cancel()
				sendErr = err
				for it := range out {
					slog.Error("Failed to send item to Datadog", "item", it)
				}
			}
			return err
		})
		eg, ctx := errgroup.WithContext(ctx)
		eg.Go(func() error {
			return c.FetchFunctionLogs(ctx, p, out)
		})
		eg.Go(func() error {
			return c.FetchPipelineResolverLogs(ctx, p, out)
		})
		if err := eg.Wait(); err != nil {
			slog.Error("Error in fetching logs", "error", err)
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(ingestCmd)
	ingestCmd.Flags().StringVarP(&posType, "pos", "", "file", "position type (file|artifact)")
	ingestCmd.Flags().StringVarP(&datadogService, "datadog-service", "", "", "Datadog service name")
	ingestCmd.Flags().StringSliceVarP(&datadogTags, "datadog-tag", "", []string{}, "Datadog tag")
}
