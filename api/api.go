package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/ethereum/Hui-TxState/config"
	"github.com/ethereum/Hui-TxState/types"
	"github.com/ethereum/Hui-TxState/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-xorm/xorm"
	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	tgbot "github.com/suiguo/hwlib/telegram_bot"
)

const ADDRLEN = 42

type apiData struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Data      string `json:"data"`
	Uid       string `json:"uid"`
	RequestId string `json:"requestId"`
	ChainId   string `json:"chainId"`
	Value     string `json:"value"`
	// From string `json:"from"`

}

// from := gjson.Get(data, "from")
// to := gjson.Get(data, "to")
// inputData := gjson.Get(data, "data")
// userID := gjson.Get(data, "uid")
// requestID := gjson.Get(data, "requestId")
// chainId := gjson.Get(data, "chainId")
// value := gjson.Get(data, "value")
type ApiService struct {
	db     types.IDB
	config *config.Config
}

func NewApiService(db types.IDB, c *config.Config) *ApiService {
	return &ApiService{
		db:     db,
		config: c,
	}
}

func (s *ApiService) Run(conf config.ServerConf) {
	r := gin.Default()

	r.POST("/tx/create", s.AddTask)

	err := r.Run(fmt.Sprintf(":%d", conf.Port))
	if err != nil {
		logrus.Fatalf("start http server err:%v", err)
	}
}

func checkAddr(addr string) error {
	if addr[:2] != "0x" {
		return errors.New("addr must start with 0x")
	}
	if len(addr) != ADDRLEN {
		return errors.New("addr len wrong ,must 40")
	}
	return nil
}

func checkInput(addr string) error {
	if addr[:2] != "0x" {
		return errors.New("addr must start with 0x")
	}
	return nil
}

// 接收注册过来的消息，存入db作为tx初始状态
func (s *ApiService) AddTask(c *gin.Context) {
	res := types.HttpRes{}
	res.Code = -99 //默认错误
	res.Message = "未知错误"
	defer func() {
		c.SecureJSON(200, res)
	}()
	out, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		res.Code = -1
		res.Message = err.Error()
		return
	}
	data := &apiData{}
	err = json.Unmarshal(out, data)
	if err != nil {
		res.Code = -2
		res.Message = err.Error()
		return
	}
	from := data.From
	to := data.To
	inputData := data.Data
	userID := data.Uid
	requestID := data.RequestId
	chainId := data.ChainId
	value := data.Value

	fmt.Println(chainId)
	//check params
	err = checkAddr(from)
	if err != nil {
		res.Code = http.StatusBadRequest
		res.Message = err.Error()
		return
		// c.SecureJSON(http.StatusBadRequest, res)
	}
	err = checkAddr(to)
	if err != nil {
		res.Code = http.StatusBadRequest
		res.Message = err.Error()
		return
		// c.SecureJSON(http.StatusBadRequest, res)
	}
	err = checkInput(inputData)
	if err != nil {
		res.Code = http.StatusBadRequest
		res.Message = err.Error()
		return
		// c.SecureJSON(http.StatusBadRequest, res)
	}
	//插入task
	task := types.TransactionTask{
		UUID:      time.Now().Unix(),
		UserID:    userID,
		From:      from,
		To:        to,
		Value:     value,
		InputData: inputData,
		ChainId:   8888,
		RequestId: requestID,
	}
	task.State = int(types.TxInitState)

	err = utils.CommitWithSession(s.db, func(session *xorm.Session) (execErr error) {
		if err := s.db.SaveTxTask(session, &task); err != nil {
			logrus.Errorf("save transaction task error:%v tasks:[%v]", err, task)
			return err
		}
		s.tgAlert(&task)
		return nil
	})
	if err != nil {
		res.Code = http.StatusInternalServerError
		res.Message = err.Error()
		return
		// c.SecureJSON(http.StatusInternalServerError, res)
	}

	res.Code = 0
	res.Message = "success"
	// c.SecureJSON(http.StatusOK, res)
}

func (c *ApiService) tgAlert(task *types.TransactionTask) {
	var (
		msg string
		err error
	)
	msg, err = createInitMsg(task)
	if err != nil {
		logrus.Errorf("create init msg err:%v,state:%d,tid:%d", err, task.State, task.ID)
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

func createInitMsg(task *types.TransactionTask) (string, error) {
	//告警消息
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("交易状态变化:->交易初始\n\n"))
	buffer.WriteString(fmt.Sprintf("UserID: %v\n\n", task.UserID))
	buffer.WriteString(fmt.Sprintf("From: %v\n\n", task.From))
	buffer.WriteString(fmt.Sprintf("To: %v\n\n", task.To))
	buffer.WriteString(fmt.Sprintf("Data: %v\n\n", task.InputData))
	buffer.WriteString(fmt.Sprintf("State: %v\n\n", task.State))

	return buffer.String(), nil
}
