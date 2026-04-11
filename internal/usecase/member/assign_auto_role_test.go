package member_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/ak1m1tsu/barman/internal/domain/guild"
	memberuc "github.com/ak1m1tsu/barman/internal/usecase/member"
	mockguild "github.com/ak1m1tsu/barman/mocks/guild"
	mockmember "github.com/ak1m1tsu/barman/mocks/member"
)

func TestAssignAutoRole_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("assigns role when auto-role is configured", func(t *testing.T) {
		repo := mockguild.NewMockRepository(t)
		assigner := mockmember.NewMockRoleAssigner(t)

		repo.EXPECT().FindByID(ctx, "guild1").
			Return(&domain.Guild{ID: "guild1", AutoRoleID: "role1"}, nil)
		assigner.EXPECT().AssignRole(ctx, "guild1", "user1", "role1").Return(nil)

		uc := memberuc.NewAssignAutoRole(repo, assigner)
		require.NoError(t, uc.Execute(ctx, "guild1", "user1"))
	})

	t.Run("does nothing when guild has no config", func(t *testing.T) {
		repo := mockguild.NewMockRepository(t)
		assigner := mockmember.NewMockRoleAssigner(t)

		repo.EXPECT().FindByID(ctx, "guild1").Return(nil, nil)

		uc := memberuc.NewAssignAutoRole(repo, assigner)
		require.NoError(t, uc.Execute(ctx, "guild1", "user1"))
	})

	t.Run("does nothing when auto-role is empty", func(t *testing.T) {
		repo := mockguild.NewMockRepository(t)
		assigner := mockmember.NewMockRoleAssigner(t)

		repo.EXPECT().FindByID(ctx, "guild1").
			Return(&domain.Guild{ID: "guild1", AutoRoleID: ""}, nil)

		uc := memberuc.NewAssignAutoRole(repo, assigner)
		require.NoError(t, uc.Execute(ctx, "guild1", "user1"))
	})

	t.Run("propagates repository error", func(t *testing.T) {
		repo := mockguild.NewMockRepository(t)
		assigner := mockmember.NewMockRoleAssigner(t)

		repo.EXPECT().FindByID(ctx, "guild1").Return(nil, errors.New("db error"))

		uc := memberuc.NewAssignAutoRole(repo, assigner)
		assert.Error(t, uc.Execute(ctx, "guild1", "user1"))
	})

	t.Run("propagates assigner error", func(t *testing.T) {
		repo := mockguild.NewMockRepository(t)
		assigner := mockmember.NewMockRoleAssigner(t)

		repo.EXPECT().FindByID(ctx, "guild1").
			Return(&domain.Guild{ID: "guild1", AutoRoleID: "role1"}, nil)
		assigner.EXPECT().AssignRole(ctx, "guild1", "user1", "role1").
			Return(errors.New("discord error"))

		uc := memberuc.NewAssignAutoRole(repo, assigner)
		assert.Error(t, uc.Execute(ctx, "guild1", "user1"))
	})
}
