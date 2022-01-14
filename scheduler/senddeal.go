package scheduler

import (
	"github.com/filswan/go-swan-client/command"

	"path/filepath"
	"payment-bridge/common/constants"
	"payment-bridge/config"
	"payment-bridge/database"
	"payment-bridge/models"
	"time"

	"github.com/filswan/go-swan-lib/logs"

	libconstants "github.com/filswan/go-swan-lib/constants"
	"github.com/robfig/cron"
)

func SendDealScheduler() {
	c := cron.New()
	err := c.AddFunc(config.GetConfig().ScheduleRule.SendDealRule, func() {
		logs.GetLogger().Println("send deal scheduler is running at " + time.Now().Format("2006-01-02 15:04:05"))
		err := sendDeals()
		if err != nil {
			logs.GetLogger().Error(err)
			return
		}
	})
	if err != nil {
		logs.GetLogger().Fatal(err)
		return
	}
	c.Start()
}

func sendDeals() error {
	dealList, err := GetTaskListShouldBeSendDealFromLocal()
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	cmdAutoBidDeal := &command.CmdAutoBidDeal{
		SwanApiUrl:             config.GetConfig().SwanApi.ApiUrl,
		SwanApiKey:             config.GetConfig().SwanApi.ApiKey,
		SwanAccessToken:        config.GetConfig().SwanApi.AccessToken,
		LotusClientApiUrl:      config.GetConfig().Lotus.ClientApiUrl,
		LotusClientAccessToken: config.GetConfig().Lotus.ClientAccessToken,
		SenderWallet:           config.GetConfig().FileCoinWallet,
	}
	cmdAutoBidDeal.DealSourceIds = append(cmdAutoBidDeal.DealSourceIds, libconstants.TASK_SOURCE_ID_SWAN_PAYMENT)

	for _, deal := range dealList {
		logs.GetLogger().Info("start to send deal for task:", deal.TaskUuid)
		cmdAutoBidDeal.OutputDir = filepath.Dir(deal.CarFilePath)

		_, fileDescs, err := cmdAutoBidDeal.SendAutoBidDealsByTaskUuid(deal.TaskUuid)
		if err != nil {
			logs.GetLogger().Error(err)
			continue
		}

		if len(fileDescs) == 0 {
			logs.GetLogger().Info("no deals sent")
			continue
		}

		deal.SendDealStatus = constants.SEND_DEAL_STATUS_SUCCESS
		deal.ClientWalletAddress = cmdAutoBidDeal.SenderWallet
		deal.DealCid = fileDescs[0].Deals[0].DealCid
		deal.MinerFid = fileDescs[0].Deals[0].MinerFid

		err = database.SaveOne(deal)
		if err != nil {
			logs.GetLogger().Error(err)
			continue
		}
	}

	return nil
}

func GetTaskListShouldBeSendDealFromLocal() ([]*models.DealFile, error) {
	whereCondition := "send_deal_status ='' and lower(lock_payment_status)=lower('" + constants.LOCK_PAYMENT_STATUS_PROCESSING + "') and task_uuid != '' "
	dealList, err := models.FindDealFileList(whereCondition, "create_at desc", "50", "0")
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	return dealList, nil
}

type TaskDetailResult struct {
	Data struct {
		AverageBid       string        `json:"average_bid"`
		Bid              []interface{} `json:"bid"`
		BidCount         int           `json:"bid_count"`
		Deal             []Deal        `json:"deal"`
		DealCompleteRate string        `json:"deal_complete_rate"`
		Miner            Miner         `json:"miner"`
		Poster           Poster        `json:"poster"`
		Task             Task          `json:"task"`
		TotalDealCount   int           `json:"total_deal_count"`
		TotalItems       int           `json:"total_items"`
	} `json:"data"`
	Status string `json:"status"`
}

type Miner struct {
	AddressBalance       string      `json:"address_balance"`
	AdjustedPower        string      `json:"adjusted_power"`
	AutoBidTaskCnt       int         `json:"auto_bid_task_cnt"`
	AutoBidTaskPerDay    int         `json:"auto_bid_task_per_day"`
	BidMode              int         `json:"bid_mode"`
	LastAutoBidAt        int64       `json:"last_auto_bid_at"`
	Location             string      `json:"location"`
	MaxPieceSize         string      `json:"max_piece_size"`
	MinPieceSize         string      `json:"min_piece_size"`
	MinerID              string      `json:"miner_id"`
	OfflineDealAvailable bool        `json:"offline_deal_available"`
	Price                interface{} `json:"price"`
	Score                int         `json:"score"`
	StartEpoch           int         `json:"start_epoch"`
	Status               string      `json:"status"`
	UpdateTimeStr        string      `json:"update_time_str"`
	VerifiedPrice        interface{} `json:"verified_price"`
	YearlyPrice          interface{} `json:"yearly_price"`
	YearlyVerifiedPrice  interface{} `json:"yearly_verified_price"`
}

type Task struct {
	BidMode        int         `json:"bid_mode"`
	CreatedOn      string      `json:"created_on"`
	CuratedDataset interface{} `json:"curated_dataset"`
	Description    interface{} `json:"description"`
	Duration       int         `json:"duration"`
	ExpireDays     int         `json:"expire_days"`
	FastRetrieval  int         `json:"fast_retrieval"`
	IsPublic       int         `json:"is_public"`
	MaxPrice       string      `json:"max_price"`
	MinPrice       interface{} `json:"min_price"`
	MinerID        interface{} `json:"miner_id"`
	SourceID       int         `json:"source_id"`
	Status         string      `json:"status"`
	Tags           interface{} `json:"tags"`
	TaskFileName   string      `json:"task_file_name"`
	TaskID         int         `json:"task_id"`
	TaskName       string      `json:"task_name"`
	Type           string      `json:"type"`
	UpdatedOn      string      `json:"updated_on"`
	UUID           string      `json:"uuid"`
}

type Deal struct {
	ContractID    string      `json:"contract_id"`
	Cost          interface{} `json:"cost"`
	CreatedAt     string      `json:"created_at"`
	DealCid       interface{} `json:"deal_cid"`
	FileName      string      `json:"file_name"`
	FilePath      interface{} `json:"file_path"`
	FileSize      string      `json:"file_size"`
	FileSourceURL string      `json:"file_source_url"`
	ID            int         `json:"id"`
	Md5Origin     string      `json:"md5_origin"`
	MinerID       interface{} `json:"miner_id"`
	Note          interface{} `json:"note"`
	PayloadCid    string      `json:"payload_cid"`
	PieceCid      string      `json:"piece_cid"`
	PinStatus     string      `json:"pin_status"`
	StartEpoch    int         `json:"start_epoch"`
	Status        string      `json:"status"`
	TaskID        int         `json:"task_id"`
	UpdatedAt     string      `json:"updated_at"`
	UserID        int         `json:"user_id"`
}

type Poster struct {
	AvatarURL         string      `json:"avatar_url"`
	CompleteTaskCount int         `json:"complete_task_count"`
	ContactInfo       interface{} `json:"contact_info"`
	MemberSince       string      `json:"member_since"`
}
