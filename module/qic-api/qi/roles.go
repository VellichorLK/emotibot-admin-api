package qi

import "emotibot.com/emotigo/module/qic-api/model/v1"

const (
	CallStaffRoleName    = "staff"
	CallCustomerRoleName = "customer"
)

var callTypeDict = map[string]int8{
	CallStaffRoleName:    model.CallChanStaff,
	CallCustomerRoleName: model.CallChanCustomer,
}

func callRoleTyp(role string) int8 {
	value, found := callTypeDict[role]
	if !found {
		return model.CallChanDefault
	}
	return value
}
func callRoleTypStr(typ int8) string {
	for key, val := range callTypeDict {
		if val == typ {
			return key
		}
	}
	return "default"
}
func RoleMatcherTyp(name string) int {
	typ, exist := roleMapping[name]
	if !exist {
		return -1
	}
	return typ
}
func RoleMatcherString(typ int) string {
	name, exist := roleCodeMap[typ]
	if !exist {
		return ""
	}
	return name
}

// TODO: It should be matched with the model.GroupConditionRole
var roleMapping map[string]int = map[string]int{
	"staff":    0,
	"customer": 1,
	"any":      2,
}

var roleCodeMap map[int]string = map[int]string{
	0: "staff",
	1: "customer",
	2: "any",
}
