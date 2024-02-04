package postgres

import (
	"context"
	"time"

	"github.com/celenium-io/celestia-indexer/internal/storage"
	sdk "github.com/dipdup-net/indexer-sdk/pkg/storage"
)

func (s *StorageTestSuite) TestRollupLeaderboard() {
	for _, column := range []string{
		sizeColumn, blobsCountColumn, timeColumn, "",
	} {
		ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer ctxCancel()

		rollups, err := s.storage.Rollup.Leaderboard(ctx, column, sdk.SortOrderDesc, 10, 0)
		s.Require().NoError(err, column)
		s.Require().Len(rollups, 3, column)

		rollup := rollups[0]
		s.Require().EqualValues("Rollup 3", rollup.Name, column)
		s.Require().EqualValues("The third", rollup.Description, column)
		s.Require().EqualValues(34, rollup.Size, column)
		s.Require().EqualValues(3, rollup.BlobsCount, column)
	}
}

func (s *StorageTestSuite) TestRollupStats() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	rollup, err := s.storage.Rollup.Stats(ctx, 1)
	s.Require().NoError(err)

	s.Require().EqualValues(30, rollup.Size)
	s.Require().EqualValues(2, rollup.BlobsCount)
}

func (s *StorageTestSuite) TestRollupNamespaces() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	nsIds, err := s.storage.Rollup.Namespaces(ctx, 1, 10, 0)
	s.Require().NoError(err)
	s.Require().Len(nsIds, 2)
}

func (s *StorageTestSuite) TestRollupProviders() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	providers, err := s.storage.Rollup.Providers(ctx, 1)
	s.Require().NoError(err)
	s.Require().Len(providers, 2)
}

func (s *StorageTestSuite) TestRollupSeries() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	for _, tf := range []string{
		"day", "hour", "month",
	} {
		for _, column := range []string{
			"size", "blobs_count",
		} {
			series, err := s.storage.Rollup.Series(ctx, 1, tf, column, storage.SeriesRequest{
				From: time.Date(2023, 1, 1, 1, 1, 1, 1, time.UTC),
			})
			s.Require().NoError(err)
			s.Require().Len(series, 2)

		}
	}
}