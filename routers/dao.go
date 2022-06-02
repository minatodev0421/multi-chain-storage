package routers

import (
	"fmt"
	"multi-chain-storage/common"
	"multi-chain-storage/common/errorinfo"
	"multi-chain-storage/service"
	"net/http"
	"strings"

	"github.com/filswan/go-swan-lib/logs"
	"github.com/gin-gonic/gin"
)

func Dao(router *gin.RouterGroup) {
	router.GET("/signature/deals_to_sign/:signer_wallet_address", GetDeals2Sign)
	router.POST("/signature", WriteDaoSignature)
	router.POST("/register", RegisterDao)
}

func GetDeals2Sign(c *gin.Context) {
	logs.GetLogger().Info("ip:", c.ClientIP(), ",port:", c.Request.URL.Port())
	signerWalletAddress := strings.Trim(c.Params.ByName("signer_wallet_address"), " ")
	if signerWalletAddress == "" || !strings.HasPrefix(signerWalletAddress, "0x") {
		errMsg := "signer_wallet_address is required and should be valid address"
		logs.GetLogger().Error(errMsg)
		c.JSON(http.StatusBadRequest, common.CreateErrorResponse(errorinfo.ERROR_PARAM_INVALID_VALUE, errMsg))
		return
	}

	dealList, err := service.GetDeals2Sign(signerWalletAddress)
	if err != nil {
		logs.GetLogger().Error(err)
		c.JSON(http.StatusInternalServerError, common.CreateErrorResponse(errorinfo.ERROR_INTERNAL, err.Error()))
		return
	}

	c.JSON(http.StatusOK, common.CreateSuccessResponse(dealList))
}

type DaoSignature struct {
	DealId                 int64  `json:"deal_id"`
	RecipientWalletAddress string `json:"recipient_wallet_address"`
	TxHash                 string `json:"tx_hash"`
}

func WriteDaoSignature(c *gin.Context) {
	var daoSignature DaoSignature
	err := c.BindJSON(&daoSignature)
	if err != nil {
		logs.GetLogger().Error(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, common.CreateErrorResponse(errorinfo.ERROR_PARAM_PARSE_TO_STRUCT))
		return
	}

	if daoSignature.DealId <= 0 {
		err := fmt.Errorf("deal_id should be greater than 0")
		logs.GetLogger().Error(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, common.CreateErrorResponse(errorinfo.ERROR_PARAM_INVALID_VALUE))
		return
	}

	if daoSignature.TxHash == "" || !strings.HasPrefix(daoSignature.TxHash, "0x") {
		err := fmt.Errorf("tx_hash is invalid")
		logs.GetLogger().Error(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, common.CreateErrorResponse(errorinfo.ERROR_PARAM_INVALID_VALUE))
		return
	}

	if daoSignature.RecipientWalletAddress == "" {
		err := fmt.Errorf("recipient_wallet_address is required")
		logs.GetLogger().Error(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, common.CreateErrorResponse(errorinfo.ERROR_PARAM_NULL))
		return
	}

	err = service.WriteDaoSignature(daoSignature.TxHash, daoSignature.RecipientWalletAddress, daoSignature.DealId)
	if err != nil {
		logs.GetLogger().Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, common.CreateErrorResponse(errorinfo.ERROR_INTERNAL, err.Error()))
		return
	}

	c.JSON(http.StatusOK, common.CreateSuccessResponse(nil))
}

type DaoInfo struct {
	WalletAddress string `json:"wallet_address"`
}

func RegisterDao(c *gin.Context) {
	var daoInfo DaoInfo
	err := c.BindJSON(&daoInfo)
	if err != nil {
		logs.GetLogger().Error(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, common.CreateErrorResponse(errorinfo.ERROR_PARAM_PARSE_TO_STRUCT))
		return
	}

	if daoInfo.WalletAddress == "" || !strings.HasPrefix(daoInfo.WalletAddress, "0x") {
		err := fmt.Errorf("wallet_address is invalid")
		logs.GetLogger().Error(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, common.CreateErrorResponse(errorinfo.ERROR_PARAM_INVALID_VALUE))
		return
	}

	err = service.RegisterDao(daoInfo.WalletAddress)
	if err != nil {
		logs.GetLogger().Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, common.CreateErrorResponse(errorinfo.ERROR_INTERNAL, err.Error()))
		return
	}

	c.JSON(http.StatusOK, common.CreateSuccessResponse(nil))
}
