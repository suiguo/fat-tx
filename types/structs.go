package types

import "time"

type Base struct {
	ID        uint      `json:"id" gorm:"primary_key"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BaseTask struct {
	State   int
	Message string
	params  string // used for create sub tasks
}

type PartReBalanceState = int

const (
	PartReBalanceInit PartReBalanceState = iota
	PartReBalanceCross
	PartReBalanceTransferIn
	PartReBalanceFarm
	PartReBalanceSuccess
	PartReBalanceFailed
)

type PartReBalanceTask struct {
	*Base
	*BaseTask
}

type AssetTransferTask struct {
	*Base
	*BaseTask
	RebalanceId  uint64 `xorm:"rebalance_id"`
	TransferType uint8  `xorm:"transfer_type"`
	Progress     string `xorm:"progress"`
}

type TransactionTask struct {
	*Base
	*BaseTask
	RebalanceId     uint64 `xorm:"rebalance_id"`
	TransferId      uint64 `xorm:"transfer_id"`
	Nonce           int    `xorm:"nonce"`
	ChainId         int    `xorm:"chain_id"`
	From            string `xorm:"from"`
	To              string `xorm:"to"`
	ContractAddress string `xorm:"contract_address"`
	Value           int    `xorm:"value"`
	UnSignData      string `xorm:"unsigned_data"`
	SignData        string `xorm:"signed_data"`
}

type InvestTask struct {
	*Base
	*BaseTask
	RebalanceId uint64 `xorm:"rebalance_id"`
}
