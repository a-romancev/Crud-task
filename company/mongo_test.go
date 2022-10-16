package company

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMongo(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
		t.Run("Happy path", func(t *testing.T) {
			t.Parallel()

			companies := NewMongo(dockerMongo(t))

			created, err := companies.Create(ctx, Company{
				ID:   uuid.New(),
				Name: uuid.NewString(),
			})
			require.NoError(t, err)

			fetched, err := companies.FetchOne(ctx, Lookup{
				ID: created.ID,
			})
			require.NoError(t, err)
			assert.Equal(t, created, fetched)
		})

		t.Run("Duplicated name returns error", func(t *testing.T) {
			t.Parallel()

			companies := NewMongo(dockerMongo(t))

			_, err := companies.Create(ctx, Company{
				ID:   uuid.New(),
				Name: "test",
			})
			require.NoError(t, err)
			_, err = companies.Create(ctx, Company{
				ID:   uuid.New(),
				Name: "test",
			})
			require.ErrorIs(t, err, ErrDuplicatedEntry)
		})
	})

	t.Run("Fetch", func(t *testing.T) {
		t.Parallel()

		companies := NewMongo(dockerMongo(t))

		created := []Company{
			{
				ID:   uuid.New(),
				Name: uuid.NewString(),
			},
			{
				ID:   uuid.New(),
				Name: uuid.NewString(),
			},
		}
		for _, c := range created {
			_, err := companies.Create(ctx, c)
			require.NoError(t, err)
		}

		fetched, err := companies.Fetch(ctx, Lookup{})
		require.NoError(t, err)
		assert.ElementsMatch(t, created, fetched)
	})
}
