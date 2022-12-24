package services

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/fat-tx/config"
	"github.com/ethereum/fat-tx/types"
	"github.com/ethereum/fat-tx/utils"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/go-xorm/xorm"
	"github.com/sirupsen/logrus"
	tgbot "github.com/suiguo/hwlib/telegram_bot"
	"log"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
)

type BroadcastService struct {
	db     types.IDB
	config *config.Config
}

func NewBoradcastService(db types.IDB, c *config.Config) *BroadcastService {
	return &BroadcastService{
		db:     db,
		config: c,
	}
}

func (c *BroadcastService) BroadcastTx(task *types.TransactionTask) (finished bool, err error) {
	hash, err := c.handleBroadcastTx(task)
	if err != nil {
		return false, err
	}
	task.Hash = hash
	task.State = int(types.TxBroadcastState)
	err = utils.CommitWithSession(c.db, func(s *xorm.Session) error {
		if err := c.db.UpdateTransactionTask(s, task); err != nil {
			logrus.Errorf("update transaction task error:%v tasks:[%v]", err, task)
			return err
		}
		return nil
	})
	if err != nil {
		return false, fmt.Errorf(" CommitWithSession in BroadcastTx err:%v", err)
	}
	return true, nil
}

func (c *BroadcastService) handleBroadcastTx(task *types.TransactionTask) (string, error) {
	rawTxBytes, err := hex.DecodeString(task.SignData)
	if err != nil {
		log.Fatal(err)
	}
	tx := new(ethtypes.Transaction)
	rlp.DecodeBytes(rawTxBytes, &tx)

	client, err := ethclient.Dial("http://43.198.66.226:8545")
	if err != nil {
		return "", err
	}

	err = client.SendTransaction(context.Background(), tx)
	if err != nil {
		return "", err
	}

	fmt.Printf("tx sent: %s", tx.Hash().Hex())

	//check hash ==?tx.hash
	if tx.Hash().Hex() != task.Hash {
		fmt.Printf("tx sent hash : %s is not equal to task.hash : %s", tx.Hash().Hex(), task.Hash)
	}

	return tx.Hash().Hex(), nil
}

func (c *BroadcastService) tgAlert(task *types.TransactionTask) {
	var (
		msg string
		err error
	)
	msg, err = createBoradcastMsg(task)
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
func createBoradcastMsg(task *types.TransactionTask) (string, error) {
	//告警消息
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("交易状态变化->交易广播完成\n\n"))
	buffer.WriteString(fmt.Sprintf("UserID: %v\n\n", task.UserID))
	buffer.WriteString(fmt.Sprintf("From: %v\n\n", task.From))
	buffer.WriteString(fmt.Sprintf("To: %v\n\n", task.To))
	buffer.WriteString(fmt.Sprintf("Data: %v\n\n", task.InputData))
	buffer.WriteString(fmt.Sprintf("Nonce: %v\n\n", task.Nonce))
	buffer.WriteString(fmt.Sprintf("GasPrice: %v\n\n", task.GasPrice))
	buffer.WriteString(fmt.Sprintf("Hash: %v\n\n", task.Hash))
	buffer.WriteString(fmt.Sprintf("State: %v\n\n", task.State))

	return buffer.String(), nil
}

func (c *BroadcastService) Run() error {
	tasks, err := c.db.GetOpenedBroadcastTasks()
	if err != nil {
		return fmt.Errorf("get tasks for broadcast err:%v", err)
	}

	if len(tasks) == 0 {
		return nil
	}

	for _, task := range tasks {
		_, err := c.BroadcastTx(task)
		if err == nil {
			c.tgAlert(task)
		}
	}
	return nil
}

func (c BroadcastService) Name() string {
	return "Broadcast"
}