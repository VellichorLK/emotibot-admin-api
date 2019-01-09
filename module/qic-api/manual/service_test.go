package manual

import (
	"testing"

	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/test"
)

var mockTask model.InspectTask = model.InspectTask{
	UUID:    "defaultID",
	Name:    "mockTask",
	Creator: "aabbccdd",
}

var mockUsers map[string]*model.Staff = map[string]*model.Staff{
	"aabbccdd": &model.Staff{
		UUID: "aabbccdd",
		Name: "passed",
	},
	"55690": &model.Staff{
		UUID: "55690",
		Name: "aabbccdddd",
	},
}

var mockStaffTaskInfo1 model.StaffTaskInfo = model.StaffTaskInfo{
	TaskID:    int64(0),
	StaffID:   "55688",
	StaffName: "aabbccdd",
	CallID:    int64(555),
	Status:    0,
	Type:      0,
}

var mockStaffTaskInfo2 model.StaffTaskInfo = model.StaffTaskInfo{
	TaskID:    int64(0),
	StaffID:   "55689",
	StaffName: "aabbccddee",
	CallID:    int64(556),
	Status:    1,
	Type:      0,
}

var mockStaffTaskInfo3 model.StaffTaskInfo = model.StaffTaskInfo{
	TaskID:    int64(0),
	StaffID:   "55690",
	StaffName: "aabbccdddd",
	CallID:    int64(555),
	Status:    0,
	Type:      1,
}

var mockTaskInfos map[int64]*[]model.StaffTaskInfo = map[int64]*[]model.StaffTaskInfo{
	int64(0): &[]model.StaffTaskInfo{
		mockStaffTaskInfo1,
		mockStaffTaskInfo2,
		mockStaffTaskInfo3,
	},
}

type mockTaskDao struct{}

func (dao *mockTaskDao) Create(task *model.InspectTask, sql model.SqlLike) (int64, error) {
	return int64(0), nil
}

func (dao *mockTaskDao) CountBy(filter *model.InspectTaskFilter, sql model.SqlLike) (int64, error) {
	return int64(1), nil
}

func (dao *mockTaskDao) GetBy(filter *model.InspectTaskFilter, sql model.SqlLike) ([]model.InspectTask, error) {
	t := model.InspectTask{
		UUID:    mockTask.UUID,
		Name:    mockTask.Name,
		Creator: mockTask.Creator,
	}

	mockTasks := []model.InspectTask{
		t,
	}
	return mockTasks, nil
}

func (dao *mockTaskDao) Users(uids []string, sql model.SqlLike) (map[string]*model.Staff, error) {
	return mockUsers, nil
}

func (dao *mockTaskDao) CountTaskInfoBy(filter *model.StaffTaskFilter, sql model.SqlLike) (int64, error) {
	return int64(len(mockTaskInfos)), nil
}

func (dao *mockTaskDao) GetTasksInfoBy(filter *model.StaffTaskFilter, sql model.SqlLike) (map[int64]*[]model.StaffTaskInfo, error) {
	return mockTaskInfos, nil
}

func (dao *mockTaskDao) Update(taskID int64, task *model.InspectTask, sql model.SqlLike) error {
	return nil
}

func (dao *mockTaskDao) AssignInspectTasks(assigns []model.StaffTaskInfo, sql model.SqlLike) error {
	return nil
}

func setupManualTest() (model.DBLike, model.DBLike, model.InspectTaskDao) {
	oriManualDB := manualDB
	oriAuthDB := authDB
	oriTaskDao := taskDao

	mockDB := &test.MockDBLike{}
	manualDB = mockDB
	authDB = mockDB

	mockDao := &mockTaskDao{}
	taskDao = mockDao

	return oriManualDB, oriAuthDB, oriTaskDao
}

func restoreManualTest(oriManualDB, oriAuthDB model.DBLike, oriTaskDao model.InspectTaskDao) {
	manualDB = oriManualDB
	authDB = oriAuthDB
	taskDao = oriTaskDao
}

func TestCreateTask(t *testing.T) {
	oriManualDB, oriAuthDB, oriTaskDao := setupManualTest()
	defer restoreManualTest(oriManualDB, oriAuthDB, oriTaskDao)

	id, err := CreateTask(&mockTask)
	if err != nil {
		t.Error(err)
		return
	}

	if id != mockTask.ID {
		t.Errorf("create task failed, expect id: %d, but got: %d", id, mockTask.ID)
	}
}

func TestGetTask(t *testing.T) {
	oriManualDB, oriAuthDB, oriTaskDao := setupManualTest()
	defer restoreManualTest(oriManualDB, oriAuthDB, oriTaskDao)

	filter := &model.InspectTaskFilter{}
	total, tasks, err := GetTasks(filter)
	if err != nil {
		t.Error(err)
		return
	}

	if total != int64(1) {
		t.Errorf("get total number of tasks failed, expect: %d, but got: %d", 1, total)
		return
	}

	task := tasks[0]
	creator := mockUsers[mockTask.Creator]
	if task.Creator != creator.Name {
		t.Errorf("wrong creator, expect: %s, but got: %s", creator.Name, task.Creator)
	}

	if task.InspectNum != 2 {
		t.Errorf("wrong inspect number, expect 2, but got: %d", task.InspectNum)
		return
	}

	if task.InspectCount != 1 {
		t.Errorf("wrong inspect count, expect 1, but got: %d", task.InspectNum)
		return
	}

	if task.InspectTotal != 2 {
		t.Errorf("wrong inspect count, expect 2, but got: %d", task.InspectNum)
		return
	}

	if task.Reviewer != mockStaffTaskInfo3.StaffName {
		t.Errorf("wrong reviewer name, expect %s, but got: %s", mockStaffTaskInfo3.StaffName, task.Reviewer)
		return
	}

	if task.ReviewNum != 0 {
		t.Errorf("wrong reviewer count, expect 0, but got: %d", task.ReviewNum)
		return
	}

	if task.ReviewTotal != 1 {
		t.Errorf("wrong reviewer count, expect 1, but got: %d", task.ReviewTotal)
		return
	}
}
