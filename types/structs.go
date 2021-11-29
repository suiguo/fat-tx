package types

import (
	"encoding/json"
	"time"

	"github.com/sirupsen/logrus"
)

type Base struct {
	ID        uint64    `xorm:"f_id" gorm:"primary_key"`
	CreatedAt time.Time `xorm:"created f_created_at"`
	UpdatedAt time.Time `xorm:"updated f_updated_at"`
}

type BaseTask struct {
	State   int    `xorm:"f_state"`
	Message string `xorm:"f_message"`
}

type PartReBalanceState = int

type CrossState = int
type CrossSubState int

const (
	PartReBalanceInit PartReBalanceState = iota
	PartReBalanceCross
	PartReBalanceTransferIn
	PartReBalanceInvest
	PartReBalanceSuccess
	PartReBalanceFailed
)
const (
	ToCreateSubTask CrossState = iota
	SubTaskCreated
	TaskSuc               //all sub task suc
	ToCross CrossSubState = iota
	Crossing
	Crossed
)

type TransactionType int

const (
	ReceiveFromBridge TransactionType = iota
	Invest
	Approve
)

type TaskState int

const (
	StateSuccess TaskState = iota
	StateOngoing
	StateFailed
)

type TransactionState int

const (
	TxUnInitState TransactionState = iota
	TxAuditState
	TxValidatorState
	TxSignedState
	TxCheckReceiptState
	TxSuccessState
	TxFailedState
)

type PartReBalanceTask struct {
	*Base     `xorm:"extends"`
	*BaseTask `xorm:"extends"`
	Params    string `xorm:"f_params"`
}

func (p *PartReBalanceTask) TableName() string {
	return "t_part_rebalance_task"
}

func (p *PartReBalanceTask) ReadParams() (params *Params, err error) {
	params = &Params{}
	if err = json.Unmarshal([]byte(p.Params), params); err != nil {
		logrus.Errorf("Unmarshal PartReBalanceTask params error:%v task:[%v]", err, p)
		return
	}

	return
}

type TransactionTask struct {
	*Base           `xorm:"extends"`
	*BaseTask       `xorm:"extends"`
	RebalanceId     uint64 `xorm:"f_rebalance_id"`
	TransactionType int    `xorm:"f_type"`
	//Nonce           int    `xorm:"f_nonce"`
	ChainId   int    `xorm:"f_chain_id"`
	ChainName string `xorm:"f_chain_name"`
	Params    string `xorm:"f_params"`
	//Decimal         int    `xorm:"f_decimal"`
	From            string `xorm:"f_from"`
	To              string `xorm:"f_to"`
	//Value           string `xorm:"f_value"`
	InputData   string `xorm:"f_input_data"`
	Cipher      string `xorm:"f_cipher"`
	EncryptData string `xorm:"f_encrypt_data"`
	SignData    string `xorm:"f_signed_data"`
	OrderId     int    `xorm:"f_order_id"`
	Hash        string `xorm:"f_hash"`
}

func (t *TransactionTask) TableName() string {
	return "t_transaction_task"
}

type CrossTask struct {
	*Base         `xorm:"extends"`
	RebalanceId   uint64 `xorm:"rebalance_id"`
	ChainFrom     string `xorm:"chain_from"`
	ChainFromAddr string `xorm:"chain_from_addr"`
	ChainTo       string `xorm:"chain_to"`
	ChainToAddr   string `xorm:"chain_to_addr"`
	CurrencyFrom  string `xorm:"currency_from"`
	CurrencyTo    string `xorm:"currency_to"`
	Amount        string `xorm:"amount"`
	State         int    `xorm:"state"`
}

type CrossSubTask struct {
	*Base        `xorm:"extends"`
	TaskNo       uint64 `xorm:"task_no"`
	BridgeTaskId uint64 `xorm:"bridge_task_id"` //跨链桥task_id
	ParentTaskId uint64 `xorm:"parent_id"`      //父任务id
	// ChainFrom    string
	// ChainTo      string
	// CurrencyFrom string
	// CurrencyTo   string
	Amount string `xorm:"amount"`
	State  int    `xorm:"state"`
}

type ApproveRecord struct {
	*Base   `xorm:"extends"`
	From    string `xorm:"f_from"`
	Token   string `xorm:"f_token"`
	Spender string `xorm:"f_spender"`
}

func (t *ApproveRecord) TableName() string {
	return "t_approve"
}