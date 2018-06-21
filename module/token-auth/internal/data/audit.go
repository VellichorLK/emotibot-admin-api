package data

const (
	AuditContentUserAdd      = "新增企业用户"
	AuditContentUserUpdate   = "更新企业用户"
    AuditContentUserDelete   = "删除企业用户"
    AuditContentAppAdd       = "新增机器人"
	AuditContentAppUpdate    = "更新机器人"
	AuditContentAppDelete    = "删除机器人"
	AuditContentGroupAdd     = "新增机器人群组"
	AuditContentGroupUpdate  = "更新机器人群组"
	AuditContentGroupDelete  = "删除机器人群组"
	AuditContentRoleAdd      = "新增权限"
	AuditContentRoleUpdate   = "更新权限"
	AuditContentRoleDelete   = "删除权限"
	AuditLogin               = "用户登入"
)

type AuditLog struct {
	AppID     string
	UserID    string
	UserIP    string
	Module    string
	Operation string
	Content   string
	Result    int
}
