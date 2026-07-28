package defi

import (
	"context"
	"errors"
	"fmt"

	"github.com/tn606024/defi-simplify/amount"
)

type checkpointBeforeStep struct {
	step        FlowStep
	checkpoints []amount.CheckpointRef
}

// CheckpointBefore declares token-balance checkpoints immediately before the
// wrapped step's single call. The references can later be consumed through
// amount.CheckpointDelta.
func CheckpointBefore(step FlowStep, checkpoints ...amount.CheckpointRef) FlowStep {
	return &checkpointBeforeStep{
		step:        step,
		checkpoints: append([]amount.CheckpointRef(nil), checkpoints...),
	}
}

func (s *checkpointBeforeStep) Build(ctx context.Context, env BuildEnv) (BuiltStep, error) {
	if s == nil || s.step == nil {
		return BuiltStep{}, errors.New("checkpoint step is nil")
	}
	built, err := s.step.Build(ctx, env)
	if err != nil {
		return built, err
	}
	if len(s.checkpoints) == 0 {
		return built, errors.New("at least one checkpoint is required")
	}
	if len(built.Calls) != 1 {
		return built, fmt.Errorf("checkpoint wrapper requires exactly one planned call, got %d", len(built.Calls))
	}
	for _, checkpoint := range s.checkpoints {
		if err := checkpoint.Validate(); err != nil {
			return built, err
		}
		built.Calls[0].CheckpointsBefore = append(
			built.Calls[0].CheckpointsBefore,
			CheckpointDeclaration{Ref: checkpoint},
		)
	}
	return built, nil
}
