package model

import (
	"encoding/csv"
	"os"
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

func seedTags() []Tag {
	f, err := os.Open("./testdata/seed/Tag.tsv")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	reader := csv.NewReader(f)
	reader.Comma = '\t'
	records, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}
	var tags []Tag
	for i := 1; i < len(records); i++ {
		t := &Tag{}
		Binding(t, records[i])
		tags = append(tags, *t)
	}
	return tags
}

var testTags []Tag

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
