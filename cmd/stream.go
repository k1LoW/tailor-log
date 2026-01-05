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
	"io"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/k1LoW/donegroup"
	"github.com/k1LoW/duration"
	"github.com/k1LoW/errors"
	"github.com/k1LoW/tailor-log/config"
	"github.com/k1LoW/tailor-log/item"
	"github.com/k1LoW/tailor-log/pos"
	"github.com/k1LoW/tailor-log/tailor"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var fetchInterval string

var streamCmd = &cobra.Command{
	Use:   "stream",
	Short: "Stream logs from Tailor Platform",
	Long:  `Stream logs from Tailor Platform.`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		// Disable default slog output
		devnull := slog.New(slog.NewTextHandler(io.Discard, nil))
		slog.SetDefault(devnull)
		logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

		ctx, _ := donegroup.WithCancel(cmd.Context())
		ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
		p := pos.At(workspaceID, time.Now())
		out := make(chan *item.Item)
		defer func() {
			cancel()
			close(out)
			if errr := donegroup.WaitWithTimeout(ctx, waitTimeout); errr != nil {
				err = errors.Join(err, errr)
			}
		}()

		cfg := &config.Config{}
		cfg.WorkspaceID = workspaceID
		var splitted []string
		for _, input := range inputs {
			splitted = append(splitted, strings.Split(input, ",")...)
		}
		cfg.Inputs = splitted
		c, err := tailor.New(cfg)
		if err != nil {
			return err
		}

		donegroup.Go(ctx, func() error {
			for it := range out {
				attrs := []slog.Attr{
					{Key: slog.TimeKey, Value: slog.TimeValue(it.Time)},
					{Key: "source", Value: slog.StringValue(it.Source)},
				}
				for k, v := range it.Attrs {
					attrs = append(attrs, slog.Attr{Key: k, Value: slog.AnyValue(v)})
				}
				logger.LogAttrs(
					ctx,
					slog.Level(it.Level),
					it.Message,
					attrs...,
				)
			}
			return nil
		})
		dur, err := duration.Parse(fetchInterval)
		if err != nil {
			return err
		}
		ticker := time.NewTicker(dur)
		for {
			eg := new(errgroup.Group)
			eg.Go(func() error {
				return c.FetchFunctionLogs(ctx, p, out)
			})
			eg.Go(func() error {
				return c.FetchPipelineResolverLogs(ctx, p, out)
			})
			if err := eg.Wait(); err != nil {
				return err
			}
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(streamCmd)
	streamCmd.Flags().StringVarP(&fetchInterval, "fetch-interval", "", "5sec", "Fetch interval duration")
}
