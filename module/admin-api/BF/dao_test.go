package BF

import (
	"testing"

	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestGetSSMCategory(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	useDB = db
	appid := "csbot"

	rows := sqlmock.NewRows([]string{"id", "pid", "name"}).
		AddRow(1, 0, "root").
		AddRow(2, 1, "level1-1").
		AddRow(3, 1, "level1-2").
		AddRow(4, 3, "level1-2-1")
	mock.ExpectQuery("^SELECT (.+) FROM tbl_sq_category WHERE app_id = (.+) AND is_del = 0$").WithArgs(appid).WillReturnRows(rows)

	t.Run("Test get category without deleted", func(t *testing.T) {
		root, err := getSSMCategories(appid, false)
		if err != nil {
			t.Errorf("Unexcepted error: %s", err)
			return
		}

		if root.Name != "root" {
			t.Errorf("Unexcepted root")
		}
		if root.Children == nil || len(root.Children) != 2 {
			t.Errorf("Unexcepted root children number")
		}
		if root.Children != nil {
			for _, child := range root.Children {
				switch child.ID {
				case 2:
					if child.Children == nil || len(child.Children) != 0 {
						t.Errorf("Unexcepted children number of child id = 2")
					}
				case 3:
					if child.Children == nil || len(child.Children) != 1 {
						t.Errorf("Unexcepted children number of child id = 3")
					}
				}
			}
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
			return
		}
	})
}
