package services

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ethereum/fat-tx/config"
	"github.com/ethereum/fat-tx/types"
	"github.com/ethereum/fat-tx/utils"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	tgbot "github.com/suiguo/hwlib/telegram_bot"
)

type CheckReceiptService struct {
	db     types.IDB
	config *config.Config
}

func NewCheckReceiptService(db types.IDB, c *config.Config) *CheckReceiptService {
	return &CheckReceiptService{
		db:     db,
		config: c,
	}
}

func (c *CheckReceiptService) CheckReceipt(task *types.TransactionTask) (finished bool, err error) {
	receipt, err := c.handleCheckReceipt(task)
	if err != nil {
		return false, err
	}
	b, err := json.Marshal(receipt)
	if err != nil {
		return false, err
	}
	task.Receipt = string(b)
	task.State = int(types.TxSuccessState)
	err = utils.CommitWithSession(c.db, func(s *xorm.Session) error {
		if err := c.db.UpdateTransactionTask(s, task); err != nil {
			logrus.Errorf("update transaction task error:%v tasks:[%v]", err, task)
			return err
		}
		return nil
	})
	if err != nil {
		return false, fmt.Errorf(" CommitWithSession in CheckReceipt err:%v", err)
	}
	return true, nil
}

func (c *CheckReceiptService) handleCheckReceipt(task *types.TransactionTask) (*ethtypes.Receipt, error) {
	rawTxBytes, err := hex.DecodeString(task.SignData)
	if err != nil {
		return nil, err
	}
	tx := new(ethtypes.Transaction)
	rlp.DecodeBytes(rawTxBytes, &tx)

	client, err := ethclient.Dial("http://43.198.66.226:8545")
	if err != nil {
		return nil, err
	}

	receipt, err := client.TransactionReceipt(context.Background(), common.HexToHash(task.Hash))
	if err != nil {
		return nil, err
	}

	return receipt, nil
}

func (c *CheckReceiptService) tgAlert(task *types.TransactionTask) {
	var (
		msg string
		err error
	)
	msg, err = createCheckMsg(task)
	if err != nil {
		logrus.Errorf("create assembly msg err:%v,state:%d,tid:%d", err, task.State, task.ID)
	}

	bot, err := tgbot.NewBot("5985674693:AAF94x_xI2RI69UTP-wt_QThldq-XEKGY8g")
	if err != nil {
		logrus.Fatal(err)
	}
	err = bot.SendMsg(1762573172, msg)
	if err != nil {
		logrus.Fatal(err)
	}
}
func createCheckMsg(task *types.TransactionTask) (string, error) {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("交易状态变化->交易获取收据完成\n\n"))
	buffer.WriteString(fmt.Sprintf("UserID: %v\n\n", task.UserID))
	buffer.WriteString(fmt.Sprintf("From: %v\n\n", task.From))
	buffer.WriteString(fmt.Sprintf("To: %v\n\n", task.To))
	buffer.WriteString(fmt.Sprintf("Data: %v\n\n", task.InputData))
	buffer.WriteString(fmt.Sprintf("Nonce: %v\n\n", task.Nonce))
	buffer.WriteString(fmt.Sprintf("GasPrice: %v\n\n", task.GasPrice))
	buffer.WriteString(fmt.Sprintf("Hash: %v\n\n", task.Hash))
	buffer.WriteString(fmt.Sprintf("Receipt: %v\n\n", task.Receipt))
	buffer.WriteString(fmt.Sprintf("State: %v\n\n", task.State))

	return buffer.String(), nil
}

func (c *CheckReceiptService) Run() error {
	tasks, err := c.db.GetOpenedCheckReceiptTasks()
	if err != nil {
		return fmt.Errorf("get tasks for check receipt err:%v", err)
	}

	if len(tasks) == 0 {
		return nil
	}

	for _, task := range tasks {
		_, err := c.CheckReceipt(task)
		if err == nil {
			c.tgAlert(task)
		}
	}
	return nil
}

func (c CheckReceiptService) Name() string {
	return "CheckReceipt"
}