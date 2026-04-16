package reaction

import (
	"context"

	"github.com/sirupsen/logrus"

	reactiondomain "github.com/ak1m1tsu/barman/internal/domain/reaction"
)

// IncrementStatUseCase records one use of a reaction type.
type IncrementStatUseCase struct {
	repo reactiondomain.StatsRepository
}

func NewIncrementStat(repo reactiondomain.StatsRepository) *IncrementStatUseCase {
	return &IncrementStatUseCase{repo: repo}
}

func (uc *IncrementStatUseCase) Execute(ctx context.Context, reactionType string) error {
	log := logrus.WithFields(logrus.Fields{
		"op":       "IncrementStat",
		"reaction": reactionType,
	})
	log.Info("incrementing reaction stat")

	if err := uc.repo.Increment(ctx, reactionType); err != nil {
		log.WithError(err).Error("failed to increment reaction stat")
		return err
	}

	log.Info("reaction stat incremented")
	return nil
}
