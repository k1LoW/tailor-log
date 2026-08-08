package tailor

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	tailorv1 "buf.build/gen/go/tailor-inc/tailor/protocolbuffers/go/tailor/v1"
	"connectrpc.com/connect"
	"github.com/IGLOU-EU/go-wildcard/v2"
	"github.com/k1LoW/tailor-log/item"
	"github.com/k1LoW/tailor-log/pos"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	SourceFunctionUnspecified = "tailor_platform.function.unspecified"
	SourceFunctionStandard    = "tailor_platform.function.standard"
	SourceFunctionJob         = "tailor_platform.function.job"

	functionPosKey         = "function"
	functionInputKeyPrefix = "function"
	maxPageSize            = 100
	sortBy                 = "finished_at"
)

func (c *Client) FetchFunctionLogs(ctx context.Context, pos *pos.Pos, out chan<- *item.Item) error {
	oldest := pos.Load(functionPosKey)
	slog.Info("Fetching function logs", "oldest", oldest)
	latest := oldest
	defer func() {
		slog.Info("Fetched function logs", "oldest", oldest, "latest", latest)
		pos.Store(functionPosKey, latest)
	}()
	token := ""
	for {
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
			inputKey := fmt.Sprintf("%s:%s", functionInputKeyPrefix, exec.ScriptName)
			matched := false
			for _, pattern := range c.cfg.Inputs {
				if wildcard.Match(pattern, inputKey) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
			source := SourceFunctionUnspecified
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

			// Fetch individual execution to get logs (List response does not include logs)
			detail, err := c.client.GetFunctionExecution(ctx, connect.NewRequest(&tailorv1.GetFunctionExecutionRequest{
				WorkspaceId: c.cfg.WorkspaceID,
				ExecutionId: exec.Id,
			}))
			if err != nil {
				return err
			}
			detailExec := detail.Msg.GetExecution()

			var result any = detailExec.Result
			var resultJSON map[string]any
			// construct result as json if string is valid json
			if len(detailExec.Result) > 0 && (detailExec.Result[0] == '{' || detailExec.Result[0] == '[') {
				if err := json.Unmarshal([]byte(detailExec.Result), &resultJSON); err == nil {
					result = resultJSON
				}
			}

			attrs := map[string]any{
				"tailor_platform.workspaceId":          detailExec.WorkspaceId,
				"tailor_platform.function.executionId": detailExec.Id,
				"tailor_platform.function.scriptName":  detailExec.ScriptName,
				"tailor_platform.function.type":        detailExec.Type.String(),
				"tailor_platform.function.status":      detailExec.Status.String(),
				"tailor_platform.function.logs":        detailExec.Logs,
				"tailor_platform.function.startedAt":   detailExec.StartedAt.AsTime().Format(time.RFC3339Nano),
				"tailor_platform.function.finishedAt":  detailExec.FinishedAt.AsTime().Format(time.RFC3339Nano),
				"tailor_platform.function.result":      result,
			}
			out <- &item.Item{
				Source:  source,
				Time:    detailExec.FinishedAt.AsTime(),
				Level:   level,
				Message: detailExec.Result,
				Attrs:   attrs,
			}
			if latest.Before(detailExec.FinishedAt.AsTime()) {
				latest = detailExec.FinishedAt.AsTime()
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
