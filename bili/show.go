package bili

import (
	"bilibili-ticket-go/bili/models/api"
	r "bilibili-ticket-go/bili/models/return"
	"bilibili-ticket-go/bili/token"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

const frontVersion = "134" // Stored on "https://s1.hdslb.com/bfs/static/platform/static/js/vendor.4052c4899bf31668a61b.js?277f136a95f6bbe03034" -> var version = "134";

func (c *Client) GetProjectInformation(projectID string) (error, *r.ProjectInformation) {
	res, err := c.http.R().Get(fmt.Sprintf("https://show.bilibili.com/api/ticket/project/getV2?version=%s&id=%s&project_id=%s&requestSource=pc-new", frontVersion, projectID, projectID))
	if err != nil {
		return err, &r.ProjectInformation{
			ProjectID: projectID,
			StartTime: time.Time{},
			EndTime:   time.Time{},
		}
	}
	var data api.MainApiDataRoot[api.TicketProjectInformationStruct]
	err = res.Unmarshal(&data)
	if err != nil {
		return err, &r.ProjectInformation{
			ProjectID: projectID,
			StartTime: time.Time{},
			EndTime:   time.Time{},
		}
	}
	if err = data.CheckValid(); err != nil {
		return err, &r.ProjectInformation{
			ProjectID: projectID,
			StartTime: time.Time{},
			EndTime:   time.Time{},
		}
	}
	return nil, &r.ProjectInformation{
		ProjectID:    projectID,
		StartTime:    time.Unix(data.Data.End, 0),
		EndTime:      time.Unix(data.Data.Start, 0),
		IsHotProject: data.Data.HotProject,
	}
}

func (c *Client) GetTicketSkuIDsByProjectID(projectID string) (error, []r.TicketSkuScreenID) {
	res, err := c.http.R().Get(fmt.Sprintf("https://show.bilibili.com/api/ticket/project/getV2?version=%s&id=%s&project_id=%s&requestSource=pc-new", frontVersion, projectID, projectID))
	if err != nil {
		return err, nil
	}
	var data api.MainApiDataRoot[api.TicketProjectInformationStruct]
	err = res.Unmarshal(&data)
	if err != nil {
		return err, nil
	}
	if err = data.CheckValid(); err != nil {
		return err, nil
	}
	tickets := make([]r.TicketSkuScreenID, 0)
	for _, s := range data.Data.ScreenList {
		for _, t := range s.TicketList {
			ticket := r.TicketSkuScreenID{
				ScreenID: s.ScreenId,
				SkuID:    t.SkuId,
				Name:     t.ScreenName,
				Desc:     t.Desc,
				Price:    t.Price,
				Flags: struct {
					Number      int
					DisplayName string
				}{
					Number:      t.SaleFlag.Number,
					DisplayName: t.SaleFlag.DisplayName,
				},
				SaleStat: struct {
					Start time.Time
					End   time.Time
				}{
					Start: time.Unix(t.SaleStart, 0),
					End:   time.Unix(t.SaleEnd, 0),
				},
			}
			tickets = append(tickets, ticket)
		}
	}
	return nil, tickets
}

func (c *Client) GetRequestTokenAndPToken(tk token.Generator, projectID string, ticket r.TicketSkuScreenID) (error, *r.RequestTokenAndPToken) {
	form := map[string]any{
		"project_id":    projectID,
		"screen_id":     ticket.ScreenID,
		"order_type":    1,
		"count":         1,
		"sku_id":        ticket.SkuID,
		"requestSource": "pc-new",
	}
	if tk.IsHotProject() {
		form["newRisk"] = true
		form["token"] = tk.GenerateTokenPrepareStage()
	}
	req, err := c.http.R().SetBodyJsonMarshal(form).Post("https://show.bilibili.com/api/ticket/order/prepare?project_id=" + projectID)
	if err != nil {
		return err, nil
	}
	var data api.ShowApiDataRoot[api.RequestTokenAndPTokenStruct]
	err = req.Unmarshal(&data)
	if err != nil {
		return err, nil
	}
	if err = data.CheckValid(); err != nil {
		return err, nil
	}
	return nil, &r.RequestTokenAndPToken{
		RequestToken: data.Data.Token,
		PToken:       data.Data.Ptoken,
		GaiaToken:    data.Data.GaData.GriskId,
	}
}

func (c *Client) GetConfirmInformation(tokens *r.RequestTokenAndPToken, projectID string) (error, *api.ConfirmStruct) {
	req, err := c.http.R().SetQueryParams(map[string]string{
		"token":         tokens.RequestToken,
		"ptoken":        tokens.PToken,
		"project_id":    projectID,
		"projectId":     projectID,
		"requestSource": "pc-new",
		"voucher":       "",
	}).Get("https://show.bilibili.com/api/ticket/order/confirmInfo")
	if err != nil {
		return err, nil
	}
	var data api.ShowApiDataRoot[api.ConfirmStruct]
	err = req.Unmarshal(&data)
	if err != nil {
		return err, nil
	}
	if err = data.CheckValid(); err != nil {
		return err, nil
	}
	return nil, &data.Data
}

func (c *Client) SubmitOrder(tk token.Generator, whenGenPToken time.Time, tokens *r.RequestTokenAndPToken, projectID string, ticket r.TicketSkuScreenID, buyer api.BuyerStruct) (error, int, string, *api.TicketOrderStruct) {
	bs, err := json.Marshal([1]api.BuyerStruct{buyer})
	if err != nil {
		return err, -1, "", nil
	}
	form := map[string]string{
		"project_id":    projectID,
		"screen_id":     strconv.FormatInt(ticket.ScreenID, 10),
		"count":         "1",
		"pay_money":     strconv.Itoa(ticket.Price),
		"order_type":    "1",
		"timestamp":     strconv.FormatInt(whenGenPToken.Unix(), 10),
		"deviceId":      c.fingerprint.Buvidfp,
		"buyer_info":    string(bs),
		"click_postion": fmt.Sprintf("{\"x\":948,\"y\":997,\"origin\":%d,\"now\":%d}", whenGenPToken, time.Now().Unix()),
		"sku_id":        strconv.FormatInt(ticket.SkuID, 10),
		"requestSource": "pc-new",
	}
	if tk.IsHotProject() {
		form["newRisk"] = "true"
		form["ctoken"] = tk.GenerateTokenCreateStage(whenGenPToken)
		form["ptoken"] = tokens.PToken
		form["token"] = tokens.RequestToken
		form["orderCreateUrl"] = "https://show.bilibili.com/api/ticket/order/createV2"
	}
	req, err := c.http.R().SetFormData(form).Post("https://show.bilibili.com/api/ticket/order/createV2?project_id=" + projectID)
	var data api.ShowApiDataRoot[*api.TicketOrderStruct]
	err = req.Unmarshal(&data)
	if err != nil {
		return err, -1, "", nil
	}
	return nil, data.GetCode(), data.GetMessage(), data.Data
}
