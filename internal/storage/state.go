package storage

import (
	"context"
	"time"

	"github.com/dipdup-net/indexer-sdk/pkg/storage"
	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

//go:generate mockgen -source=$GOFILE -destination=mock/$GOFILE -package=mock -typed
type IState interface {
	storage.Table[*State]

	ByName(ctx context.Context, name string) (State, error)
}

// State -
type State struct {
	bun.BaseModel `bun:"state" comment:"Current indexer state"`

	ID                 uint64          `bun:",pk,autoincrement"         comment:"Unique internal identity"`
	Name               string          `bun:",unique:state_name"        comment:"Indexer name"`
	LastHeight         uint64          `bun:"last_height"               comment:"Last block height"`
	LastTime           time.Time       `bun:"last_time"                 comment:"Time of last block"`
	ChainId            string          `bun:"chain_id"                  comment:"Celestia chain id"`
	TotalTx            uint64          `bun:"total_tx"                  comment:"Transactions count in celestia"`
	TotalAccounts      uint64          `bun:"total_accounts"            comment:"Accounts count in celestia"`
	TotalNamespaces    uint64          `bun:"total_namespaces"          comment:"Namespaces count in celestia"`
	TotalNamespaceSize uint64          `bun:"total_namspaces_size"      comment:"Total namespace size"`
	TotalSupply        decimal.Decimal `bun:"total_supply,type:numeric" comment:"Total supply in celestia"`
}

// TableName -
func (State) TableName() string {
	return "state"
}
