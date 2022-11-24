package realmarket

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/KeithZHIJIAN/nce-realmarket/utils"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

const PLUpdateInterval = 5 * time.Second

func PLAgentStart() {
	time.Sleep(10 * time.Second)
	addPLTrendRecord(time.Now().Truncate(PLUpdateInterval))
	for {
		now := time.Now()
		tick := now.Truncate(PLUpdateInterval).Add(PLUpdateInterval)
		time.Sleep(tick.Sub(now))
		addPLTrendRecord(tick)
	}
}

func addPLTrendRecord(curr time.Time) {
	PLs := make(map[string]decimal.Decimal)
	rows := utils.ReadUsersBalance()
	for rows.Next() {
		var userID string
		var amount decimal.Decimal
		var locked decimal.Decimal
		//orderid, quantity, price, openquantity, fillcost, createdat, updatedat
		if err := rows.Scan(&userID, &amount, &locked); err != nil {
			panic(fmt.Errorf(err.Error()))
		}
		PLs[userID] = amount.Add(locked)
	}
	for user := range PLs {
		rows = utils.ReadUserAsset(user)
		for rows.Next() {
			var symbol string
			var amount decimal.Decimal
			var locked decimal.Decimal
			if err := rows.Scan(&symbol, &amount, &locked); err != nil {
				panic(fmt.Errorf(err.Error()))
			}
			value, ok := MarketPrices.Load(strings.ToUpper(symbol))
			if !ok {
				continue
			}
			PLs[user] = PLs[user].Add(value.(decimal.Decimal).Mul(amount.Add(locked)))
		}
	}
	for user, amount := range PLs {
		utils.CreatePL(user, amount, curr)
	}
}

type ProfitOrLoss struct {
	Time   time.Time       `json:"time"`
	Amount decimal.Decimal `json:"amount"`
}

func PLTrendHandler(c *gin.Context) {
	userID := c.Query("userid")
	rows := utils.ReadUserPL(userID)
	pls := make([]ProfitOrLoss, 0)
	for rows.Next() {
		pl := ProfitOrLoss{}
		if err := rows.Scan(&pl.Time, &pl.Amount); err != nil {
			panic(fmt.Errorf(err.Error()))
		}
		pls = append(pls, pl)
	}
	c.JSON(http.StatusOK, pls)
}

func topupHandler(c *gin.Context) {
	var requestBody struct {
		UserID string          `json:"userid"`
		Amount decimal.Decimal `json:"amount"`
	}
	if err := c.BindJSON(&requestBody); err != nil {
		log.Println("[websocket server]: Top up bind JSON failed, error: ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if requestBody.Amount.IsPositive() {
		if err := utils.CreateTopUps(requestBody.UserID, requestBody.Amount, time.Now()); err != nil {
			log.Println("[websocket server]: Create top up record failed, error: ", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	} else {
		log.Println("[websocket server]: Top up amount negative.")
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("Negative top up amount.")})
		return
	}
}

func cumulativePLHandler(c *gin.Context) {
	userID := c.Query("userid")
	rows := utils.ReadLastPL(userID)
	var lastPL decimal.Decimal
	for rows.Next() {
		if err := rows.Scan(&lastPL); err != nil {
			panic(fmt.Errorf(err.Error()))
		}
	}
	rows = utils.ReadFirstPL(userID)
	var firstPL decimal.Decimal
	for rows.Next() {
		if err := rows.Scan(&firstPL); err != nil {
			panic(fmt.Errorf(err.Error()))
		}
	}
	rows = utils.ReadCumulativeTopups(userID)
	var cumulativeTopups decimal.Decimal
	for rows.Next() {
		if err := rows.Scan(&cumulativeTopups); err != nil {
			log.Println("No topup record")
			cumulativeTopups = decimal.Zero
		}
	}
	c.JSON(http.StatusOK, gin.H{"cumu": lastPL.Sub(firstPL).Sub(cumulativeTopups), "ytm": lastPL.Sub(firstPL).Sub(cumulativeTopups).DivRound(lastPL.Sub(cumulativeTopups), 4).Mul(decimal.NewFromInt(100))})
}

func currentLearnStageHandler(c *gin.Context) {
	userID := c.Query("userid")
	rows := utils.ReadUserLearnStageByID(userID)
	learnstage := 0
	for rows.Next() {
		if err := rows.Scan(&learnstage); err != nil {
			panic(fmt.Errorf(err.Error()))
		}
	}
	c.JSON(http.StatusOK, gin.H{"learnstage": learnstage})
}

func balanceHandler(c *gin.Context) {
	userID := c.Query("userid")
	rows := utils.ReadUserBalanceByID(userID)
	var balance decimal.Decimal
	for rows.Next() {
		if err := rows.Scan(&balance); err != nil {
			panic(fmt.Errorf(err.Error()))
		}
	}
	c.JSON(http.StatusOK, gin.H{"balance": balance})
}

func assetsHandler(c *gin.Context) {
	log.Println("assetsHandler called")
	walletID := c.Query("walletid")
	symbol := strings.ToLower(c.Query("symbol"))
	rows := utils.ReadWalletAssetByIDAndSymbol(walletID, symbol)
	var amount decimal.Decimal
	for rows.Next() {
		if err := rows.Scan(&amount); err != nil {
			panic(fmt.Errorf(err.Error()))
		}
	}
	c.JSON(http.StatusOK, gin.H{"amount": amount})
}

// func apiDocsHandler(c *gin.Context) {
// 	log.Println("api")
// 	filename := "API.py"
// 	//Seems this headers needed for some browsers (for example without this headers Chrome will download files as txt)
// 	c.Header("Content-Description", "File Transfer")
// 	c.Header("Content-Transfer-Encoding", "binary")
// 	c.Header("Content-Disposition", "attachment; filename="+filename)
// 	c.Header("Content-Type", "application/octet-stream")
// 	c.File("../files/API.py")
// 	filename = "NUSWAP_APIS.xlsx"
// 	c.Header("Content-Disposition", "attachment; filename="+filename)
// 	c.File("../files/NUSWAP_APIS.xlsx")
// }
