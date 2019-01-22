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
		Paging: &Pagination{
			Limit: 5,
		},
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
	query := TagQuery{
		ID:               []uint64{1, 2},
		IgnoreSoftDelete: true,
	}
	tags, err := dao.Tags(nil, query)
	if err != nil {
		t.Fatal("expect ok, but got error: ", err)
	}
	if !reflect.DeepEqual(tags, expectedTags) {
		t.Logf("tags: %+v\n expected tags: %+v\n", tags, expectedTags)
		t.Error("expect got tags but not the same")
	}
}

func TestTagDaoNewTagsIntegration(t *testing.T) {
	skipIntergartion(t)
	dao := newTestingTagDao(t)
	tags := testTags[:2]
	results, err := dao.NewTags(nil, tags)
	if err == ErrAutoIDDisabled {
		//Need to Get the correct result back
	} else if err != nil {
		t.Fatal("expect ok, but got error: ", err)
	}
	if reflect.DeepEqual(tags, results) {
		t.Fatal("expect result should replace tags id, but it is the same")
	}
	if len(results) != len(tags) {
		t.Fatal("expect results length to be the same of input, but got ", len(results))
	}
	if results[0].ID != 3 {
		t.Error("expect new tag id incremental to 3, but no.")
	}
	if results[1].ID != 4 {
		t.Error("expect new tag id incremental to 4, but no.")
	}
}

func TestTagDaoDeleteTagsIntegration(t *testing.T) {
	skipIntergartion(t)
	dao := newTestingTagDao(t)
	query := TagQuery{
		ID: []uint64{2},
	}
	count, err := dao.DeleteTags(nil, query)
	if err != nil {
		t.Fatal("expect delete tags to be ok, but got error: ", err)
	}
	if count != 1 {
		t.Error("expect delete 1 tag, but got ", count)
	}
	tags, err := dao.Tags(nil, query)
	if err != nil {
		t.Fatal("expect get tags to be ok, but got error: ", err)
	}
	if len(tags) != 0 {
		t.Error("expect get empty tags, but got ", len(tags))
	}
}

func TestTagDaoCountTagsIntegration(t *testing.T) {
	skipIntergartion(t)
	dao := newTestingTagDao(t)
	var enterprise = "csbot"
	type args struct {
		query TagQuery
	}
	testcases := []struct {
		Name     string
		args     args
		output   uint
		hasError bool
	}{
		{
			"enterprise only",
			args{
				query: TagQuery{
					Enterprise: &enterprise,
				},
			},
			uint(len(testTags)),
			false,
		},
		{
			"should ignore Paging",
			args{
				query: TagQuery{
					Enterprise: &enterprise,
					Paging: &Pagination{
						Limit: 100,
						Page:  1,
					},
				},
			},
			uint(len(testTags)),
			false,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.Name, func(t *testing.T) {
			count, err := dao.CountTags(nil, tt.args.query)
			if tt.hasError && err == nil {
				t.Fatal("expect it is a bad case, but no error returned")
			}
			if !tt.hasError && err != nil {
				t.Fatal("expect count to be ok, but got error: ", err)
			}
			if tt.output != count {
				t.Error("expect count to be ", tt.output, ", but got ", count)
			}
		})
	}

}
