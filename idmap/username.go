package idmap

import "sync"

// 用户昵称缓存: 官方群事件里携带的 username (如加群申请的申请人昵称) 暂存在内存,
// 供 get_stranger_info 等需要展示昵称的 onebot action 反查使用。键为官方 openid。
// 说明: QQ官方机器人API没有"按openid查用户资料"的接口, 因此昵称只能靠事件顺带缓存,
// 重启后清空, 对应 action 会返回空昵称(回复仍是合法 JSON)。
var usernameCache sync.Map

// StoreUsernameV2 缓存 openid 对应的用户昵称
func StoreUsernameV2(openid, username string) {
	if openid == "" || username == "" {
		return
	}
	usernameCache.Store(openid, username)
}

// RetrieveUsernameByOpenID 读取缓存的用户昵称, 无则返回空串
func RetrieveUsernameByOpenID(openid string) string {
	if openid == "" {
		return ""
	}
	v, ok := usernameCache.Load(openid)
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

// 入群申请缓存: join_request_id -> {group_openid, member_openid}
// 原因: SparkBridge groupRequest 插件审批时只传 flag(join_request_id)/approve/reason,
// 不带 group_id/user_id, 而官方审批接口需要 group_openid + member_openid 两个路径参数,
// 因此事件到达时先缓存映射, 审批时按 flag 反查。
var joinRequestCache sync.Map

// StoreJoinRequestV2 缓存入群申请 id 对应的群/成员 openid
func StoreJoinRequestV2(joinRequestID, groupOpenID, memberOpenID string) {
	if joinRequestID == "" || groupOpenID == "" || memberOpenID == "" {
		return
	}
	joinRequestCache.Store(joinRequestID, [2]string{groupOpenID, memberOpenID})
}

// RetrieveJoinRequestV2 按入群申请 id 反查群/成员 openid, 无则返回空串
func RetrieveJoinRequestV2(joinRequestID string) (string, string) {
	if joinRequestID == "" {
		return "", ""
	}
	v, ok := joinRequestCache.Load(joinRequestID)
	if !ok {
		return "", ""
	}
	pair, _ := v.([2]string)
	return pair[0], pair[1]
}
