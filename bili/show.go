package bili

import (
	"bilibili-ticket-go/bili/models/response"
	"fmt"
)

const frontVersion = "134" // Stored on "https://s1.hdslb.com/bfs/static/platform/static/js/vendor.4052c4899bf31668a61b.js?277f136a95f6bbe03034" -> var version = "134";

func (c *Client) GetTicketSkuIDByProjectID(projectID string) error {
	res, err := c.http.R().Get(fmt.Sprintf("https://show.bilibili.com/api/ticket/project/getV2?version=%s&id=%s&project_id=%s&requestSource=pc-new", frontVersion, projectID, projectID))
	if err != nil {
		return err
	}
	var r response.DataRoot[response.TicketProjectInformationStruct]
	err = res.Unmarshal(&r)
	if err != nil {
		return err
	}
	if err = r.CheckValid(); err != nil {
		return err
	}
	return nil
}
