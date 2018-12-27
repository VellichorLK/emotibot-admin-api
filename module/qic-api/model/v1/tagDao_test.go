package model

import (
	"database/sql"
	"reflect"
	"testing"

	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func newMockedTagDao(t *testing.T) (*TagSQLDao, sqlmock.Sqlmock) {
	db, mocker, err := sqlmock.New()
	if err != nil {
		t.Fatal("sqlmock new failed, ", err)
	}
	mocker.ExpectQuery("SHOW TABLEs LIKE '" + tblTags + "'").WillReturnRows(sqlmock.NewRows([]string{"TABLE"}).AddRow(tblTags))
	dao, err := NewTagSQLDao(db)
	if err != nil {
		t.Fatal("new TagSQLDao failed, ", err)
	}
	return dao, mocker
}

func newTestingTagDao(t *testing.T) *TagSQLDao {
	db, err := sql.Open("mysql", "root:password@tcp(127.0.0.1)/QISYS?parseTime=true&loc=Asia%2FTaipei")
	if err != nil {
		t.Fatal("expect db open success but got error: ", err)
	}
	dao, err := NewTagSQLDao(db)
	if err != nil {
		t.Fatal("epxect new tag sql dao success, but got error: ", err)
	}
	return dao
}

var testTags = []Tag{
	Tag{
		ID:               1,
		UUID:             "94d58cb937f34291be262095ce974f2e",
		Enterprise:       "csbot",
		Name:             "Test1",
		PositiveSentence: `["test1-1", "test1-2", "test1-3"]`,
		NegativeSentence: `[]`,
		IsDeleted:        true,
		Typ:              0,
		CreateTime:       1545901909,
		UpdateTime:       1545901927,
	},
	Tag{
		ID:               2,
		UUID:             "5e46b3ee737c45afb29f2a243c1aae7e",
		Enterprise:       "csbot",
		Name:             "Test2",
		PositiveSentence: `["test2-1"]`,
		NegativeSentence: `["test2-2", "test2-3"]`,
		IsDeleted:        false,
		Typ:              1,
		CreateTime:       1545901951,
		UpdateTime:       1545901959,
	},
}

func TestNewTagSQLDao(t *testing.T) {
	newMockedTagDao(t)
}

func TestTagDaoTags(t *testing.T) {
	dao, mocker := newMockedTagDao(t)
	rows := sqlmock.NewRows(tagSelectColumns)
	expectedTags := testTags[:2]
	for _, tag := range expectedTags {
		rows.AddRow(tag.ID, tag.UUID, tag.IsDeleted, tag.Name,
			tag.Typ, tag.PositiveSentence, tag.NegativeSentence,
			tag.CreateTime, tag.UpdateTime, tag.Enterprise)
	}
	mocker.ExpectQuery("SELECT .+ FROM `" + tblTags + "`").WillReturnRows(rows)
	tags, err := dao.Tags(nil, TagQuery{
		ID: []uint64{1, 2},
	})
	if err != nil {
		t.Fatal("expect ok but got err: ", err)
	}
	if !reflect.DeepEqual(tags, expectedTags) {
		t.Logf("tags: %+v\n expected tags: %+v\n", tags, expectedTags)
		t.Fatal("expected tags to be the same of expected tags")
	}
}

func TestTagDaoTagsIntegration(t *testing.T) {
	skipIntergartion(t)
	dao := newTestingTagDao(t)
	expectedTags := testTags[:2]
	tags, err := dao.Tags(nil, TagQuery{ID: []uint64{1, 2}})
	if err != nil {
		t.Fatal("expect ok, but got error: ", err)
	}
	if !reflect.DeepEqual(tags, expectedTags) {
		t.Logf("tags: %+v\n expected tags: %+v\n", tags, expectedTags)
		t.Error("expect got tags but not the same")
	}
}
