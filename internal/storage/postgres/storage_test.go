package postgres

import (
	"context"
	"database/sql"
	"encoding/hex"
	"testing"
	"time"

	"github.com/dipdup-io/celestia-indexer/internal/storage"
	"github.com/dipdup-net/go-lib/config"
	"github.com/dipdup-net/go-lib/database"
	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/suite"
)

const testIndexerName = "test_indexer"

// StorageTestSuite -
type StorageTestSuite struct {
	suite.Suite
	psqlContainer *database.PostgreSQLContainer
	storage       Storage
	pm            database.RangePartitionManager
}

// SetupSuite -
func (s *StorageTestSuite) SetupSuite() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer ctxCancel()

	psqlContainer, err := database.NewPostgreSQLContainer(ctx, database.PostgreSQLContainerConfig{
		User:     "user",
		Password: "password",
		Database: "db_test",
		Port:     5432,
		Image:    "postgres:15",
	})
	s.Require().NoError(err)
	s.psqlContainer = psqlContainer

	strg, err := Create(ctx, config.Database{
		Kind:     config.DBKindPostgres,
		User:     s.psqlContainer.Config.User,
		Database: s.psqlContainer.Config.Database,
		Password: s.psqlContainer.Config.Password,
		Host:     s.psqlContainer.Config.Host,
		Port:     s.psqlContainer.MappedPort().Int(),
	})
	s.Require().NoError(err)
	s.storage = strg

	s.pm = database.NewPartitionManager(s.storage.Connection(), database.PartitionByYear)
	currentTime, err := time.Parse(time.RFC3339, "2023-07-04T03:10:57+00:00")
	s.Require().NoError(err)
	err = s.pm.CreatePartitions(ctx, currentTime, storage.Tx{}.TableName(), storage.Event{}.TableName(), storage.Message{}.TableName())
	s.Require().NoError(err)

	db, err := sql.Open("postgres", s.psqlContainer.GetDSN())
	s.Require().NoError(err)

	fixtures, err := testfixtures.New(
		testfixtures.Database(db),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory("../../../test/data"),
	)
	s.Require().NoError(err)
	s.Require().NoError(fixtures.Load())
	s.Require().NoError(db.Close())
}

// TearDownSuite -
func (s *StorageTestSuite) TearDownSuite() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	s.Require().NoError(s.storage.Close())
	s.Require().NoError(s.psqlContainer.Terminate(ctx))
}

func (s *StorageTestSuite) TestStateGetByName() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	state, err := s.storage.State.ByName(ctx, testIndexerName)
	s.Require().NoError(err)
	s.Require().EqualValues(1, state.ID)
	s.Require().EqualValues(1000, state.LastHeight)
	s.Require().EqualValues(394067, state.TotalTx)
	s.Require().EqualValues(12512357, state.TotalAccounts)
	s.Require().EqualValues(1000, state.TotalNamespaces)
	s.Require().EqualValues(324234, state.TotalNamespaceSize)
	s.Require().Equal(testIndexerName, state.Name)
}

func (s *StorageTestSuite) TestStateGetByNameFailed() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	_, err := s.storage.State.ByName(ctx, "unknown")
	s.Require().Error(err)
}

func (s *StorageTestSuite) TestBlockLast() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	block, err := s.storage.Blocks.Last(ctx)
	s.Require().NoError(err)
	s.Require().EqualValues(1000, block.Height)
	s.Require().EqualValues("1", block.VersionApp)
	s.Require().EqualValues("11", block.VersionBlock)
	s.Require().EqualValues(0, block.TxCount)

	hash, err := hex.DecodeString("6A30C94091DA7C436D64E62111D6890D772E351823C41496B4E52F28F5B000BF")
	s.Require().NoError(err)
	s.Require().Equal(hash, block.Hash)
}

func (s *StorageTestSuite) TestBlockByHeight() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	block, err := s.storage.Blocks.ByHeight(ctx, 1000)
	s.Require().NoError(err)
	s.Require().EqualValues(1000, block.Height)
	s.Require().EqualValues("1", block.VersionApp)
	s.Require().EqualValues("11", block.VersionBlock)
	s.Require().EqualValues(0, block.TxCount)

	hash, err := hex.DecodeString("6A30C94091DA7C436D64E62111D6890D772E351823C41496B4E52F28F5B000BF")
	s.Require().NoError(err)
	s.Require().Equal(hash, block.Hash)
}

func (s *StorageTestSuite) TestAddressByHash() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	hash, err := hex.DecodeString("5F7A8DDFE6136FE76B65B9066D4F816D707F")
	s.Require().NoError(err)

	address, err := s.storage.Address.ByHash(ctx, hash)
	s.Require().NoError(err)
	s.Require().EqualValues(1, address.Id)
	s.Require().EqualValues(100, address.Height)
	s.Require().Equal("123", address.Balance.String())
	s.Require().Equal(hash, address.Hash)
}

func (s *StorageTestSuite) TestEventByTxId() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	events, err := s.storage.Event.ByTxId(ctx, 1)
	s.Require().NoError(err)
	s.Require().Len(events, 1)
	s.Require().EqualValues(2, events[0].Id)
	s.Require().EqualValues(1000, events[0].Height)
	s.Require().EqualValues(1, events[0].Position)
	s.Require().Equal(storage.EventTypeMint, events[0].Type)
}

func (s *StorageTestSuite) TestEventByBlock() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	events, err := s.storage.Event.ByBlock(ctx, 1000)
	s.Require().NoError(err)
	s.Require().Len(events, 1)
	s.Require().EqualValues(1, events[0].Id)
	s.Require().EqualValues(1000, events[0].Height)
	s.Require().EqualValues(0, events[0].Position)
	s.Require().Equal(storage.EventTypeBurn, events[0].Type)
}

func (s *StorageTestSuite) TestMessageByTxId() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	msgs, err := s.storage.Message.ByTxId(ctx, 1)
	s.Require().NoError(err)
	s.Require().Len(msgs, 2)
	s.Require().EqualValues(1, msgs[0].Id)
	s.Require().EqualValues(1000, msgs[0].Height)
	s.Require().EqualValues(0, msgs[0].Position)
	s.Require().Equal(storage.MsgTypeWithdrawDelegatorReward, msgs[0].Type)
}

func (s *StorageTestSuite) TestNamespaceId() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	namespaceId, err := hex.DecodeString("5F7A8DDFE6136FE76B65B9066D4F816D707F")
	s.Require().NoError(err)

	namespaces, err := s.storage.Namespace.ByNamespaceId(ctx, namespaceId)
	s.Require().NoError(err)
	s.Require().Len(namespaces, 2)

	s.Require().EqualValues(1, namespaces[0].ID)
	s.Require().EqualValues(0, namespaces[0].Version)
	s.Require().EqualValues(1234, namespaces[0].Size)
	s.Require().Equal(namespaceId, namespaces[0].NamespaceID)

	s.Require().EqualValues(2, namespaces[1].ID)
	s.Require().EqualValues(1, namespaces[1].Version)
	s.Require().EqualValues(1255, namespaces[1].Size)
	s.Require().Equal(namespaceId, namespaces[1].NamespaceID)
}

func (s *StorageTestSuite) TestNamespaceIdAndVersion() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	namespaceId, err := hex.DecodeString("5F7A8DDFE6136FE76B65B9066D4F816D707F")
	s.Require().NoError(err)

	namespace, err := s.storage.Namespace.ByNamespaceIdAndVersion(ctx, namespaceId, 1)
	s.Require().NoError(err)

	s.Require().EqualValues(2, namespace.ID)
	s.Require().EqualValues(1, namespace.Version)
	s.Require().EqualValues(1255, namespace.Size)
	s.Require().Equal(namespaceId, namespace.NamespaceID)
}

func (s *StorageTestSuite) TestTxByHash() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxCancel()

	txHash, err := hex.DecodeString("652452A670018D629CC116E510BA88C1CABE061336661B1F3D206D248BD558AF")
	s.Require().NoError(err)

	tx, err := s.storage.Tx.ByHash(ctx, txHash)
	s.Require().NoError(err)

	s.Require().EqualValues(1, tx.Id)
	s.Require().EqualValues(0, tx.Position)
	s.Require().EqualValues(1000, tx.Height)
	s.Require().EqualValues(0, tx.TimeoutHeight)
	s.Require().EqualValues(80410, tx.GasWanted)
	s.Require().EqualValues(77483, tx.GasUsed)
	s.Require().EqualValues(1, tx.EventsCount)
	s.Require().EqualValues(2, tx.MessagesCount)
	s.Require().Equal(txHash, tx.Hash)
	s.Require().Equal(storage.StatusSuccess, tx.Status)
	s.Require().Equal("memo", tx.Memo)
	s.Require().Equal("sdk", tx.Codespace)
	s.Require().Equal("80410", tx.Fee.String())
}

func (s *StorageTestSuite) TestNotify() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := s.storage.Notificator.Subscribe(ctx, "test")
	s.Require().NoError(err)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	var ticks int
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-s.storage.Notificator.Listen():
			log.Info().Str("msg", msg.Extra).Str("channel", msg.Channel).Msg("new message")
			s.Require().Equal("test", msg.Channel)
			s.Require().Equal("message", msg.Extra)
			if ticks == 2 {
				return
			}
		case <-ticker.C:
			ticks++
			err = s.storage.Notificator.Notify(ctx, "test", "message")
			s.Require().NoError(err)
		}
	}
}

func TestSuiteStorage_Run(t *testing.T) {
	suite.Run(t, new(StorageTestSuite))
}