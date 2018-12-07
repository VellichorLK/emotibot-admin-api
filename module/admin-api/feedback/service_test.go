package feedback

import (
	"testing"
)

type mockDao struct {
	Reasons     map[string][]*Reason
	ReasonIDMax int64
}

func (dao mockDao) GetReasons(appid string) ([]*Reason, error) {
	if _, ok := dao.Reasons[appid]; ok {
		return dao.Reasons[appid], nil
	}
	return []*Reason{}, nil
}

func (dao mockDao) AddReason(appid string, content string) (int64, error) {
	if dao.Reasons == nil {
		dao.Reasons = map[string][]*Reason{}
	}
	if _, ok := dao.Reasons[appid]; !ok {
		dao.Reasons[appid] = []*Reason{}
	}

	dao.ReasonIDMax++
	dao.Reasons[appid] = append(dao.Reasons[appid], &Reason{
		ID:      dao.ReasonIDMax,
		Content: content,
	})

	return dao.ReasonIDMax, nil
}

func (dao mockDao) DeleteReason(appid string, id int64) error {
	if _, ok := dao.Reasons[appid]; !ok {
		return nil
	}
	return nil
}

func (dao *mockDao) Init() {
	dao.Reasons = map[string][]*Reason{
		"csbot": []*Reason{
			&Reason{
				ID:      1,
				Content: "reason1",
			},
			&Reason{
				ID:      2,
				Content: "reason2",
			},
		},
		"test": []*Reason{},
	}
	dao.ReasonIDMax = int64(0)
}

func TestServiceGetReasons(t *testing.T) {
	dao := mockDao{}
	dao.Init()
	serviceDao = dao
	t.Parallel()
	t.Run("Test empty reason", func(t *testing.T) {
		reasons, err := GetReasons("test")
		if err != nil {
			t.Error(err.Error())
			return
		}
		if len(reasons) != 0 {
			t.Errorf("Excepted 0 data, get %d", len(reasons))
			return
		}
	})

	t.Run("Test error input", func(t *testing.T) {
		_, err := GetReasons("")
		if err == nil {
			t.Error("Function doesn't check appid")
		}
	})

	t.Run("Test common", func(t *testing.T) {
		reasons, err := GetReasons("csbot")
		if err != nil {
			t.Error(err.Error())
			return
		}
		if len(reasons) != 2 {
			t.Errorf("Excepted 2 data, get %d", len(reasons))
			return
		}
	})
}

func TestServiceAddReason(t *testing.T) {
	dao := mockDao{}
	dao.Init()
	serviceDao = dao
	t.Parallel()
	t.Run("Test error input", func(t *testing.T) {
		_, err := AddReason("", "reason-unknown")
		if err == nil {
			t.Error("Function doesn't check appid")
		}
		_, err = AddReason("test", "")
		if err == nil {
			t.Error("Function doesn't check input content")
		}
	})

	t.Run("Test common", func(t *testing.T) {
		_, err := AddReason("test-add", "reason-test")
		if err != nil {
			t.Error(err.Error())
			return
		}
		reasons, err := GetReasons("test-add")
		if len(reasons) == 0 {
			t.Errorf("Excepted more than 0 data")
			return
		}
	})
}
