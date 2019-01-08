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

var mockUsers map[string]string = map[string]string{
	"aabbccdd": "passed",
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

func (DAO *mockTaskDao) Users(uids []string, sql model.SqlLike) (map[string]string, error) {
	return mockUsers, nil
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

	uuid, err := CreateTask(&mockTask)
	if err != nil {
		t.Error(err)
		return
	}

	if uuid != mockTask.UUID {
		t.Errorf("create task failed, expect uuid: %s, but got: %s", uuid, mockTask.UUID)
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
	if task.Creator != mockUsers[mockTask.Creator] {
		t.Errorf("wrong creator, expect: %s, but got: %s", mockUsers[mockTask.Creator], task.Creator)
	}
}
