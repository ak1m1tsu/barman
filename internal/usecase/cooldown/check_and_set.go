package cooldown

import (
	"context"
	"time"

	cooldowndomain "github.com/ak1m1tsu/barman/internal/domain/cooldown"
)

const CooldownDuration = time.Hour

// CheckAndSetUseCase checks whether a user is on cooldown and, if not, records the current time.
type CheckAndSetUseCase struct {
	repo cooldowndomain.Repository
}

func NewCheckAndSet(repo cooldowndomain.Repository) *CheckAndSetUseCase {
	return &CheckAndSetUseCase{repo: repo}
}

// Execute returns true if the user is allowed (not on cooldown) and records the timestamp.
// Returns false (with nil error) if the user is still on cooldown.
func (uc *CheckAndSetUseCase) Execute(ctx context.Context, userID string) (bool, error) {
	usedAt, err := uc.repo.GetUsedAt(ctx, userID)
	if err != nil {
		return false, err
	}

	if !usedAt.IsZero() && time.Since(usedAt) < CooldownDuration {
		return false, nil
	}

	if err := uc.repo.SetUsedAt(ctx, userID, time.Now()); err != nil {
		return false, err
	}

	return true, nil
}
