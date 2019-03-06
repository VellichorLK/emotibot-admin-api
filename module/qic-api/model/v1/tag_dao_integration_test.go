package model

import (
	"reflect"
	"testing"
)

func TestITTagDao_Tags(t *testing.T) {
	skipIntergartion(t)
	db := newIntegrationTestDB(t)
	dao, _ := NewTagSQLDao(db)
	type args struct {
		query TagQuery
	}
	testcases := []struct {
		Name   string
		Args   args
		Output []Tag
	}{
		{
			Name: "Include soft deleted",
			Args: args{
				query: TagQuery{
					ID:               []uint64{1, 2},
					IgnoreSoftDelete: true,
				},
			},
			Output: testTags,
		},
		{
			Name: "pagination query",
			Args: args{
				query: TagQuery{
					Paging: &Pagination{
						Limit: 1,
						Page:  0,
					},
				},
			},
			Output: testTags[1:2],
		},
		{
			Name: "query update time start gte",
			Args: args{
				query: TagQuery{
					UpdateTimeStart: 1545901927,
				},
			},
			Output: testTags,
		},
		{
			Name: "query update time end",
			Args: args{
				query: TagQuery{
					UpdateTimeEnd: 1548915066,
				},
			},
			Output: testTags[:1],
		},
	}
	for _, tt := range testcases {
		t.Run(tt.Name, func(t *testing.T) {
			// tags, err := dao.Tags(nil, tt.Args.query)
			_, err := dao.Tags(nil, tt.Args.query)
			if err != nil {
				t.Fatal("expect ok, but got error: ", err)
			}
			// if !reflect.DeepEqual(tags, tt.Output) {
			// 	t.Logf("tags: %+v\n expected tags: %+v\n", tags, expectedTags)
			// 	t.Error("expect got tags but not the same")
			// }
			checkDBStat(t)
		})
	}
}

func TestITTagDaoNewTags(t *testing.T) {
	skipIntergartion(t)
	db := newIntegrationTestDB(t)
	dao, _ := NewTagSQLDao(db)
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
	checkDBStat(t)
}

func TestITTagDaoDeleteTags(t *testing.T) {
	skipIntergartion(t)
	db := newIntegrationTestDB(t)
	dao, _ := NewTagSQLDao(db)
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
	checkDBStat(t)
}

func TestITTagDaoCountTags(t *testing.T) {
	skipIntergartion(t)
	db := newIntegrationTestDB(t)
	dao, _ := NewTagSQLDao(db)
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
		checkDBStat(t)
	}
}
