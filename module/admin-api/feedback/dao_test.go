package feedback

import (
	"testing"

	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestDaoGetReasons(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	dao := &feedbackDao{}
	appid := "csbot"

	t.Run("Test dao not set correctly", func(t *testing.T) {
		_, err := dao.GetReasons(appid)
		if err != ErrDBNotInit {
			t.Errorf("Unexcepted error: %s", err)
			return
		}
	})

	dao.db = db
	rows := sqlmock.NewRows([]string{"id", "content"}).
		AddRow(1, "reason1").
		AddRow(2, "reason2")
	mock.ExpectQuery("^SELECT (.+) FROM feedback_reason WHERE appid = (.+)$").WithArgs(appid).WillReturnRows(rows)

	t.Run("Test get reasons", func(t *testing.T) {
		reasons, err := dao.GetReasons(appid)
		if err != nil {
			t.Errorf("Unexcepted error: %s", err)
			return
		}
		if len(reasons) != 2 {
			t.Errorf("Excepted reason length is 2, get %d", len(reasons))
			return
		}
		if reasons[0].ID != 1 || reasons[0].Content != "reason1" ||
			reasons[1].ID != 2 || reasons[1].Content != "reason2" {
			t.Errorf("Excepted reason result: %+v", reasons)
			return
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
			return
		}
	})
}

func TestDaoAddReason(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	appid := "csbot"
	newReason := "reason new"
	dao := &feedbackDao{}

	t.Run("Test dao not set correctly", func(t *testing.T) {
		_, err := dao.AddReason(appid, newReason)
		if err != ErrDBNotInit {
			t.Errorf("Unexcepted error: %s", err)
			return
		}
	})

	t.Run("Test add correctly", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO feedback_reason").WithArgs(appid, newReason, getFixTimestamp()).WillReturnResult(sqlmock.NewResult(10, 1))

		dao.db = db
		timestampHandler = getFixTimestamp

		id, err := dao.AddReason(appid, newReason)
		if err != nil {
			t.Errorf("Unexcepted error: %s", err)
			return
		}
		if id != 10 {
			t.Errorf("Excepted add id is 10, got %d", id)
			return
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
			return
		}
	})
}

func TestDeleteReason(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	appid := "csbot"
	id := int64(10)
	dao := &feedbackDao{}

	t.Run("Test dao not set correctly", func(t *testing.T) {
		err = dao.DeleteReason(appid, id)
		if err != ErrDBNotInit {
			t.Errorf("Unexcepted error: %s", err)
			return
		}
	})

	t.Run("Test delete reason", func(t *testing.T) {
		mock.ExpectExec("DELETE FROM feedback_reason").WithArgs(appid, id).WillReturnResult(sqlmock.NewResult(0, 1))
		dao.db = db

		err = dao.DeleteReason(appid, id)
		if err != nil {
			t.Errorf("Unexcepted error: %s", err)
			return
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
			return
		}
	})
}

func TestUpdateReason(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	appid := "csbot"
	id := int64(10)
	content := "new reason"
	dao := &feedbackDao{}

	t.Run("Test dao not set correctly", func(t *testing.T) {
		err = dao.UpdateReason(appid, id, content)
		if err != ErrDBNotInit {
			t.Errorf("Unexcepted error: %s", err)
			return
		}
	})

	t.Run("Test update reason", func(t *testing.T) {
		mock.ExpectExec("UPDATE feedback_reason").WithArgs(content, appid, id).WillReturnResult(sqlmock.NewResult(0, 1))
		dao.db = db

		err = dao.UpdateReason(appid, id, content)
		if err != nil {
			t.Errorf("Unexcepted error: %s", err)
			return
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
			return
		}
	})
}

func getFixTimestamp() int64 {
	return 1000000
}
