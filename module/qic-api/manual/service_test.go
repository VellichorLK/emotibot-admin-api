package manual

import (
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/test"
	"testing"
)

var mockTask model.InspectTask = model.InspectTask{
	UUID: "defaultID",
	Name: "mockTask",
}

type mockTaskDao struct{}

func (dao *mockTaskDao) Create(task *model.InspectTask, sql model.SqlLike) (int64, error) {
	return int64(0), nil
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
