package ticket

import (
	client "bilibili-ticket-go/bili"
	"bilibili-ticket-go/bili/models/api"
	r "bilibili-ticket-go/bili/models/return"
	"bilibili-ticket-go/bili/token"
	"bilibili-ticket-go/global"
	"bilibili-ticket-go/utils"
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"
)

type TicketRoutine struct {
	mutex     sync.Mutex
	client    client.Client
	buyerID   int64
	ProjectID int64
	skuID     int64
	screenID  int64
	ctx       context.Context
	isRunning bool
	cancel    context.CancelFunc
}

func NewTicketRoutine(client client.Client, buyerID int64, ProjectID int64, skuID int64, screenID int64) *TicketRoutine {
	ctx, cancel := context.WithCancel(context.Background())
	return &TicketRoutine{
		client:    client,
		isRunning: false,
		buyerID:   buyerID,
		ProjectID: ProjectID,
		ctx:       ctx,
		cancel:    cancel,
		skuID:     skuID,
		screenID:  screenID,
	}
}

func (tr *TicketRoutine) Start() {
	tr.mutex.Lock()
	defer tr.mutex.Unlock()

	if tr.isRunning {
		return
	}

	tr.isRunning = true
	tr.cancel = func() {
		tr.isRunning = false
	}
	go run(tr.client, tr.buyerID, tr.ProjectID, tr.skuID, tr.screenID, 500*time.Millisecond, tr.ctx)
}

func (tr *TicketRoutine) Stop() {
	tr.mutex.Lock()
	defer tr.mutex.Unlock()

	if !tr.isRunning {
		return
	}

	tr.cancel()
	tr.isRunning = false
}

func run(client client.Client, buyerID int64, projectID int64, skuID int64, screenID int64, interval time.Duration, ctx context.Context) {
	logger := utils.GetLogger(global.GetLogger(), fmt.Sprintf("%d-%d-%d", projectID, skuID, buyerID), nil)
	err, info := client.GetProjectInformation(strconv.FormatInt(projectID, 10))
	if err != nil {
		logger.Errorf("GetProjectInformation err: %v", err)
		return
	}
	var tokenGen token.Generator
	if info.IsHotProject {
		tokenGen = token.NewCTokenGenerator()
	} else {
		tokenGen = token.NewNormalTokenGenerator()
	}
	err, tickets := client.GetTicketSkuIDsByProjectID(strconv.FormatInt(projectID, 10))
	if err != nil {
		logger.Errorf("GetTicketSkuIDsByProjectID err: %v", err)
		return
	}
	var ticket *r.TicketSkuScreenID
	for _, t := range tickets {
		if t.SkuID == skuID && t.ScreenID == screenID {
			ticket = &t
			break
		}
	}
	if ticket == nil {
		logger.Errorf("Ticket with skuID %d not found in project %d", skuID, projectID)
		return
	}
	whenGenPtoken := time.Now()
	err, tk := client.GetRequestTokenAndPToken(tokenGen, strconv.FormatInt(projectID, 10), *ticket)
	if err != nil {
		logger.Errorf("GetRequestTokenAndPToken err: %v", err)
		return
	}
	var buyer *api.BuyerStruct
	var count uint16 = 0
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if count >= 61 {
				// 该换个新token去骗叔叔了
				whenGenPtoken = time.Now()
				err, tk = client.GetRequestTokenAndPToken(tokenGen, strconv.FormatInt(projectID, 10), *ticket)
				if err != nil {
					logger.Errorf("GetRequestTokenAndPToken err: %v", err)
				}
				goto SLEEP
			}
			if buyer != nil {
				err, confirm := client.GetConfirmInformation(tk, strconv.FormatInt(projectID, 10))
				if err != nil {
					logger.Errorf("GetConfirmInformation err: %v", err)
					goto SLEEP
				}
				for _, b := range confirm.BuyerList.List {
					if b.Id == buyerID {
						buyer = &b
					}
				}
				if buyer == nil {
					logger.Errorf("GetConfirmInformation buyer not found in project %d", projectID)
					return
				}
			}
			err, code, msg, to := client.SubmitOrder(tokenGen, whenGenPtoken, tk, strconv.FormatInt(projectID, 10), *ticket, *buyer)
			if err != nil {
				logger.Errorf("SubmitOrder err: %v", err)
				goto SLEEP
			}
			if to == nil {
				// 被做局了
				whenGenPtoken = time.Now()
				err, tk = client.GetRequestTokenAndPToken(tokenGen, strconv.FormatInt(projectID, 10), *ticket)
				if err != nil {
					logger.Errorf("GetRequestTokenAndPToken err: %v", err)
				}
				goto SLEEP
			}
			if code == 0 || code == 100048 || code == 100079 {
				// 肘击成功
				logger.Infof("SubmitOrder success, orderID: %d", to.OrderId)
				return
			} else if code == 100034 {
				// 价格不对捏
				ticket.Price = to.PayMoney
			}
			logger.Infof("%s (%d)", msg, code)
		SLEEP:
			count++
			time.Sleep(interval)
		}
	}
}
