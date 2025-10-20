package tailor

import (
	"context"
	"log/slog"
	"time"

	tailorv1 "buf.build/gen/go/tailor-inc/tailor/protocolbuffers/go/tailor/v1"
	"connectrpc.com/connect"
	"github.com/k1LoW/tailor-log/item"
	"github.com/k1LoW/tailor-log/pos"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	SourceFuncitonUnspecified = "tailor_platform.function.unspecified"
	SourceFunctionStandard    = "tailor_platform.function.standard"
	SourceFunctionJob         = "tailor_platform.function.job"

	functionPosKey = "function"
	maxPageSize    = 1000
	sortBy         = "finished_at"
)

func (c *Client) FetchFunctionLogs(ctx context.Context, pos *pos.Pos, out chan<- *item.Item) error {
	oldest := pos.Load(functionPosKey)
	latest := oldest
	defer func() {
		slog.Info("Fetched function logs", "oldest", oldest, "latest", latest)
		pos.Store(functionPosKey, latest)
	}()
	token := ""
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		executions, err := c.client.ListFunctionExecutions(ctx, connect.NewRequest(&tailorv1.ListFunctionExecutionsRequest{
			WorkspaceId:   c.cfg.WorkspaceID,
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
		slog.Info("Fetched function executions", "count", len(executions.Msg.GetExecutions()))
		for _, exec := range executions.Msg.GetExecutions() {
			if latest.Before(exec.FinishedAt.AsTime()) {
				latest = exec.FinishedAt.AsTime()
			}
			source := SourceFuncitonUnspecified
			switch exec.Type {
			case tailorv1.FunctionExecution_TYPE_STANDARD:
				source = SourceFunctionStandard
			case tailorv1.FunctionExecution_TYPE_JOB:
				source = SourceFunctionJob
			}
			level := item.LevelInfo
			switch exec.Status {
			case tailorv1.FunctionExecution_STATUS_RUNNING:
				continue
			case tailorv1.FunctionExecution_STATUS_SUCCESS:
				level = item.LevelInfo
			case tailorv1.FunctionExecution_STATUS_FAILED:
				level = item.LevelError
			}
			attrs := map[string]any{
				"tailor_platform.workspaceId":          exec.WorkspaceId,
				"tailor_platform.function.executionId": exec.Id,
				"tailor_platform.function.scriptName":  exec.ScriptName,
				"tailor_platform.function.type":        exec.Type.String(),
				"tailor_platform.function.status":      exec.Status.String(),
				"tailor_platform.function.logs":        exec.Logs,
				"tailor_platform.function.startedAt":   exec.StartedAt.AsTime().Format(time.RFC3339Nano),
				"tailor_platform.function.finishedAt":  exec.FinishedAt.AsTime().Format(time.RFC3339Nano),
			}
			out <- &item.Item{
				Source:  source,
				Time:    exec.FinishedAt.AsTime(),
				Level:   level,
				Message: exec.Result,
				Attrs:   attrs,
			}
		}
		nextToken := executions.Msg.GetNextPageToken()
		if nextToken == "" {
			break
		}
		token = nextToken
	}
	return nil
}
