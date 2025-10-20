package tailor

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	tailorv1 "buf.build/gen/go/tailor-inc/tailor/protocolbuffers/go/tailor/v1"
	"connectrpc.com/connect"
	"github.com/k1LoW/tailor-log/item"
	"github.com/k1LoW/tailor-log/pos"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	SourcePipelineResolver      = "tailor_platform.pipeline_resolver"
	concurrencyPipelineResolver = 10
)

func (c *Client) FetchPipelineResolverLogs(ctx context.Context, pos *pos.Pos, out chan<- *item.Item) error {
	eg, _ := errgroup.WithContext(ctx)
	eg.SetLimit(concurrencyPipelineResolver)
	token := ""
	for {
		pipelines, err := c.client.ListPipelineServices(ctx, connect.NewRequest(&tailorv1.ListPipelineServicesRequest{
			WorkspaceId: c.cfg.WorkspaceID,
			PageSize:    maxPageSize,
			PageToken:   token,
		}))
		if err != nil {
			return err
		}
		for _, pipeline := range pipelines.Msg.GetPipelineServices() {
			token := ""
			namespaceName := pipeline.GetNamespace().GetName()
			for {
				resolvers, err := c.client.ListPipelineResolvers(ctx, connect.NewRequest(&tailorv1.ListPipelineResolversRequest{
					WorkspaceId:   c.cfg.WorkspaceID,
					NamespaceName: namespaceName,
					PageSize:      maxPageSize,
					PageToken:     token,
				}))
				if err != nil {
					return err
				}
				for _, resolver := range resolvers.Msg.GetPipelineResolvers() {
					name := resolver.GetName()
					eg.Go(func() error {
						return c.fetchPipelineResolverLogs(ctx, namespaceName, name, pos, out)
					})
				}
				nextToken := resolvers.Msg.GetNextPageToken()
				if nextToken == "" {
					break
				}
				token = nextToken
			}
		}
		nextToken := pipelines.Msg.GetNextPageToken()
		if nextToken == "" {
			break
		}
		token = nextToken
	}
	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}

func (c *Client) fetchPipelineResolverLogs(ctx context.Context, namespaceName, name string, pos *pos.Pos, out chan<- *item.Item) error {
	posKey := fmt.Sprintf("pipeline:%s:resolver:%s", namespaceName, name)
	oldest := pos.Load(posKey)
	latest := oldest
	defer func() {
		slog.Info("Fetched pipeline resolver logs", "namespace", namespaceName, "name", name, "oldest", oldest, "latest", latest)
		pos.Store(posKey, latest)
	}()
	token := ""
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		results, err := c.client.ListPipelineResolverExecutionResults(ctx, connect.NewRequest(&tailorv1.ListPipelineResolverExecutionResultsRequest{
			WorkspaceId:   c.cfg.WorkspaceID,
			NamespaceName: namespaceName,
			ResolverName:  name,
			View:          tailorv1.PipelineResolverExecutionResultView_PIPELINE_RESOLVER_EXECUTION_RESULT_VIEW_FULL,
			PageSize:      maxPageSize,
			PageToken:     token,
			PageDirection: tailorv1.PageDirection_PAGE_DIRECTION_ASC,
			Filter: &tailorv1.Filter{
				Condition: &tailorv1.Condition{
					Field:    sortBy,
					Operator: tailorv1.Condition_OPERATOR_GT,
					Value: &structpb.Value{
						Kind: &structpb.Value_StringValue{StringValue: oldest.Format(time.RFC3339Nano)},
					},
				},
			},
			SortBy: sortBy,
		}))
		if err != nil {
			return err
		}
		slog.Info("Fetched pipeline resolver execution results", "namespace", namespaceName, "name", name, "count", len(results.Msg.GetResults()))
		for _, result := range results.Msg.GetResults() {
			if latest.Before(result.FinishedAt.AsTime()) {
				latest = result.FinishedAt.AsTime()
			}
			source := SourcePipelineResolver
			var level item.Level
			switch result.Status {
			case tailorv1.PipelineResolverExecutionResult_RESULT_STATUS_DONE, tailorv1.PipelineResolverExecutionResult_RESULT_STATUS_RETRIED:
				level = item.LevelInfo
			case tailorv1.PipelineResolverExecutionResult_RESULT_STATUS_FAILED, tailorv1.PipelineResolverExecutionResult_RESULT_STATUS_ABORTED:
				level = item.LevelError
			default:
				continue
			}

			attrs := map[string]any{
				"tailor_platform.workspaceId":                      c.cfg.WorkspaceID,
				"tailor_platform.pipeline_resolver.namespace":      namespaceName,
				"tailor_platform.pipeline_resolver.name":           name,
				"tailor_platform.pipeline_resolver.executionId":    result.SourceExecutionId,
				"tailor_platform.pipeline_resolver.status":         result.Status.String(),
				"tailor_platform.pipeline_resolver.initialContext": result.InitialContext.AsMap(),
				"tailor_platform.pipeline_resolver.context":        result.Context.AsMap(),
				"tailor_platform.pipeline_resolver.startedAt":      result.StartedAt.AsTime().Format(time.RFC3339Nano),
				"tailor_platform.pipeline_resolver.finishedAt":     result.FinishedAt.AsTime().Format(time.RFC3339Nano),
			}

			var message string
			if result.Error != "" {
				message = result.Error
			} else {
				message = fmt.Sprintf("Pipeline resolver '%s' executed successfully", name)
			}

			out <- &item.Item{
				Source:  source,
				Time:    result.FinishedAt.AsTime(),
				Level:   level,
				Message: message,
				Attrs:   attrs,
			}
		}
		nextToken := results.Msg.GetNextPageToken()
		if nextToken == "" {
			break
		}
		token = nextToken
	}
	return nil
}
