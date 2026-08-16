package handlers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hoshinonyaruko/gensokyo/callapi"
	"github.com/hoshinonyaruko/gensokyo/idmap"
	"github.com/hoshinonyaruko/gensokyo/mylog"
	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/openapi"
)

func init() {
	callapi.RegisterHandler("set_group_add_request", HandleSetGroupAddRequest)
}

// HandleSetGroupAddRequest 处理加群请求(入群申请审批)
// onebot v11 标准参数: group_id, user_id, flag(join_request_id), approve/refuse, reason
// 官方接口: POST /v2/groups/{group_openid}/approval_join_request/{member_openid}
// 说明: 机器人需拥有群管理员身份才能审批
func HandleSetGroupAddRequest(client callapi.Client, api openapi.OpenAPI, apiv2 openapi.OpenAPI, message callapi.ActionMessage) (string, error) {
	groupID := message.Params.GroupID.(string)
	userID := message.Params.UserID.(string)
	flag := message.Params.Flag

	// 反查真实群 openid
	realGroupID, err := idmap.RetrieveRowByIDv2(groupID)
	if err != nil || realGroupID == "" {
		mylog.Printf("set_group_add_request: 无法反查群openid group[%v] err[%v]", groupID, err)
		return "", nil
	}
	// 反查真实成员 openid
	realUserID, err := idmap.RetrieveRowByIDv2(userID)
	if err != nil || realUserID == "" {
		mylog.Printf("set_group_add_request: 无法反查用户openid user[%v] err[%v]", userID, err)
		return "", nil
	}

	req := &dto.ApproveJoinRequestToCreate{
		JoinRequestID: flag,
		RejectReason:  message.Params.Reason,
	}
	if message.Params.Approve {
		req.Op = "approve"
	} else if message.Params.Refuse {
		req.Op = "decline"
	} else {
		// OneBot 规范要求 approve/refuse 至少一个为 true
		mylog.Printf("set_group_add_request: 缺少 approve/refuse 参数")
		return "", nil
	}

	mylog.Printf("set_group_add_request: 审批申请 group[%v] user[%v] op[%v] flag[%v]", realGroupID, realUserID, req.Op, flag)
	if err := apiv2.ApproveJoinRequest(context.TODO(), realGroupID, realUserID, req); err != nil {
		mylog.Printf("set_group_add_request: 审批失败: %v", err)
		return "", nil
	}

	t := time.Now()
	response := map[string]interface{}{
		"data":    map[string]interface{}{},
		"message": "success",
		"retcode": 0,
		"status":  "ok",
		"time":    t.Unix(),
	}
	if message.Echo != nil && message.Echo != "" {
		response["echo"] = message.Echo
	}
	if err := client.SendMessage(response); err != nil {
		mylog.Printf("set_group_add_request: 发送响应失败: %v", err)
	}
	result, err := json.Marshal(response)
	if err != nil {
		return "", nil
	}
	return string(result), nil
}