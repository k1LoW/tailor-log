package datadog

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"strings"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/k1LoW/donegroup"
	"github.com/k1LoW/tailor-log/config"
	"github.com/k1LoW/tailor-log/item"
)

const maxLogsPerRequest = 100

type Client struct {
	logApi *datadogV2.LogsApi
	cfg    *config.Config
}

func New(cfg *config.Config) (*Client, error) {
	if cfg == nil {
		return nil, errors.New("config is required")
	}
	if cfg.Outputs.Datadog.Service == "" {
		return nil, errors.New("datadog service is required")
	}
	if len(cfg.Outputs.Datadog.Tags) == 0 {
		return nil, errors.New("datadog tags are required")
	}
	configuration := datadog.NewConfiguration()
	apiClient := datadog.NewAPIClient(configuration)
	api := datadogV2.NewLogsApi(apiClient)
	return &Client{
		logApi: api,
		cfg:    cfg,
	}, nil
}

func (c *Client) SendLogs(ctx context.Context, in <-chan *item.Item) error {
	ctx = datadog.NewDefaultContext(ctx)
	ctx = donegroup.WithoutCancel(ctx)
	buf := make([]datadogV2.HTTPLogItem, 0, maxLogsPerRequest)
	for it := range in {
		properties := map[string]any{
			"@timestamp": it.Time.Format(time.RFC3339Nano),
			// NOTE: If the status field is included in the message, it will be overwritten
			"status": slog.Level(it.Level).String(),
		}
		maps.Copy(properties, it.Attrs)
		buf = append(buf, datadogV2.HTTPLogItem{
			Ddsource:             datadog.PtrString(it.Source),
			Ddtags:               datadog.PtrString(strings.Join(c.cfg.Outputs.Datadog.Tags, ",")),
			Message:              it.Message,
			Service:              datadog.PtrString(c.cfg.Outputs.Datadog.Service),
			AdditionalProperties: properties,
		})
		if len(buf) >= maxLogsPerRequest {
			slog.Info("Submitting logs to Datadog", "count", len(buf))
			_, _, err := c.logApi.SubmitLog(ctx, buf, *datadogV2.NewSubmitLogOptionalParameters().WithContentEncoding(datadogV2.CONTENTENCODING_GZIP))
			if err != nil {
				return fmt.Errorf("failed to submit logs to Datadog: %w", err)
			}
			buf = buf[:0] // reset
		}
	}
	if len(buf) > 0 {
		slog.Info("Submitting logs to Datadog", "count", len(buf))
		_, _, err := c.logApi.SubmitLog(ctx, buf, *datadogV2.NewSubmitLogOptionalParameters().WithContentEncoding(datadogV2.CONTENTENCODING_GZIP))
		if err != nil {
			return fmt.Errorf("failed to submit logs to Datadog: %w", err)
		}
	}
	return nil
}
