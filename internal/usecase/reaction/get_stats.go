package reaction

import (
	"context"

	"github.com/sirupsen/logrus"

	reactiondomain "github.com/ak1m1tsu/barman/internal/domain/reaction"
)

// GetStatsUseCase returns a map of reaction type → usage count.
type GetStatsUseCase struct {
	repo reactiondomain.StatsRepository
}

// NewGetStats returns a GetStatsUseCase backed by the given stats repository.
func NewGetStats(repo reactiondomain.StatsRepository) *GetStatsUseCase {
	return &GetStatsUseCase{repo: repo}
}

func (uc *GetStatsUseCase) Execute(ctx context.Context) (map[string]int64, error) {
	log := logrus.WithField("op", "GetStats")
	log.Info("fetching reaction stats")

	stats, err := uc.repo.GetAll(ctx)
	if err != nil {
		log.WithError(err).Error("failed to fetch reaction stats")
		return nil, err
	}

	log.WithField("entries", len(stats)).Info("reaction stats fetched")
	return stats, nil
}
