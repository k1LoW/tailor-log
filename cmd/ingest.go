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
	"os/signal"
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
	datagogTags    []string
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
		defer func() {
			cancel()
			if errr := donegroup.WaitWithTimeout(ctx, waitTimeout); errr != nil {
				err = errors.Join(err, errr)
			}
			if errr := p.DumpTo(cmd.Context(), posType); errr != nil {
				err = errors.Join(err, errr)
			}
		}()
		cfg := &config.Config{}
		cfg.WorkspaceID = workspaceID
		cfg.Outputs.Datadog.Service = datadogService
		cfg.Outputs.Datadog.Tags = datagogTags
		c, err := tailor.New(cfg)
		if err != nil {
			return err
		}
		out := make(chan *item.Item)
		defer close(out)
		dd, err := datadog.New(cfg)
		if err != nil {
			return err
		}
		donegroup.Go(ctx, func() error {
			return dd.SendLogs(ctx, out)
		})
		eg, ctx := errgroup.WithContext(ctx)
		eg.Go(func() error {
			return c.FetchFunctionLogs(ctx, p, out)
		})
		eg.Go(func() error {
			return c.FetchPipelineResolverLogs(ctx, p, out)
		})
		if err := eg.Wait(); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(ingestCmd)
	ingestCmd.Flags().StringVarP(&posType, "pos", "", "file", "position type (file|artifact)")
	ingestCmd.Flags().StringVarP(&datadogService, "datadog-service", "", "", "Datadog service name")
	ingestCmd.Flags().StringSliceVarP(&datagogTags, "datadog-tag", "", []string{}, "Datadog tag")
}
