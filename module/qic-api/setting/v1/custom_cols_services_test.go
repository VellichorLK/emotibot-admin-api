package setting

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"time"

	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/general"
	"emotibot.com/emotigo/module/qic-api/util/test"
)

func TestGetCustomCols(t *testing.T) {
	var (
		mockedData = []model.UserKey{
			{
				ID:         1,
				Name:       "testing",
				Enterprise: "csbot",
				InputName:  "t1",
				Type:       model.UserKeyTypString,
				IsDeleted:  false,
				CreateTime: 1549871000,
				UpdateTime: 1549872000,
			},
		}
		expectOutput = []CustomCol{
			{
				ID:        1,
				Name:      mockedData[0].Name,
				InputName: mockedData[0].InputName,
				Type:      model.UserKeyTypString,
			},
		}
		expectPaging = general.Paging{
			Limit: 1,
			Page:  1,
			Total: 10,
		}
	)
	userKeys = func(delegatee model.SqlLike, query model.UserKeyQuery) ([]model.UserKey, error) {
		return mockedData, nil
	}
	countUserKeys = func(delegatee model.SqlLike, query model.UserKeyQuery) (int64, error) {
		return 10, nil
	}
	output, page, err := GetCustomCols(model.UserKeyQuery{
		Paging: &model.Pagination{
			Limit: 1,
			Page:  1,
		},
	})
	if err != nil {
		t.Fatal("expect no error, but got ", err)
	}
	if !reflect.DeepEqual(expectOutput, output) {
		t.Logf("\nexpect: %+v\noutput: %+v\n", expectOutput, output)
		t.Error("output not as expected")
	}
	if !reflect.DeepEqual(expectPaging, page) {
		t.Logf("\nexpect: %+v\npage: %+v", expectPaging, page)
		t.Error("paging not as expected")
	}
}

func mockNewUserKey(mockdata []model.UserKey) func(delegatee model.SqlLike, key model.UserKey) (model.UserKey, error) {
	i := -1
	return func(delegatee model.SqlLike, key model.UserKey) (model.UserKey, error) {
		i++
		if i >= len(mockdata) {
			return model.UserKey{}, fmt.Errorf("only mock %d times, but call exceed", len(mockdata))
		}
		m := mockdata[i]
		// since we dont mock time.Now. we can only make sure time generate is not very different.
		// after check, give key the same time. or the deep equal will fail.
		if math.Abs(float64(key.CreateTime-m.CreateTime)) > 1000 {
			return model.UserKey{}, fmt.Errorf("expect input's create time match mock's")
		}
		key.CreateTime = m.CreateTime
		if math.Abs(float64(key.UpdateTime-m.UpdateTime)) > 1000 {
			return model.UserKey{}, fmt.Errorf("expect input's update time match mock's")
		}
		key.UpdateTime = m.UpdateTime
		key.ID = m.ID
		if !reflect.DeepEqual(m, key) {
			return model.UserKey{}, fmt.Errorf("expect input: %+v, input: %+v", m, key)
		}
		return m, nil
	}
}
func TestNewCustomCols(t *testing.T) {
	type args struct {
		requests []NewUKRequest
	}
	tests := []struct {
		name    string
		args    args
		mock    func() []model.UserKey
		wantErr bool
	}{
		{
			name: "insert one time",
			args: args{
				requests: []NewUKRequest{
					NewUKRequest{
						Enterprise: "csbot",
						Name:       "你好",
						InputName:  "HelloWorld",
						Type:       model.UserKeyTypNumber,
					},
				},
			},
			mock: func() []model.UserKey {
				now := time.Now().Unix()
				mockdata := []model.UserKey{
					model.UserKey{
						ID:         1,
						Name:       "你好",
						Enterprise: "csbot",
						InputName:  "HelloWorld",
						Type:       model.UserKeyTypNumber,
						CreateTime: now,
						UpdateTime: now,
					},
				}
				newUserKey = mockNewUserKey(mockdata)
				userKeys = func(delegatee model.SqlLike, query model.UserKeyQuery) ([]model.UserKey, error) {
					return []model.UserKey{}, nil
				}
				return mockdata
			},
		},
		{
			name: "insert the same twice",
			args: args{
				requests: []NewUKRequest{
					NewUKRequest{
						InputName:  "same",
						Enterprise: "csbot",
					},
					NewUKRequest{
						InputName:  "same",
						Enterprise: "csbot",
					},
				},
			},
			mock: func() []model.UserKey {
				now := time.Now().Unix()
				mockdata := []model.UserKey{
					model.UserKey{
						ID:         1,
						Enterprise: "csbot",
						InputName:  "same",
						CreateTime: now,
						UpdateTime: now,
					},
					model.UserKey{
						ID:         2,
						Enterprise: "csbot",
						InputName:  "same",
						CreateTime: now,
						UpdateTime: now,
					},
				}
				newUserKey = mockNewUserKey(mockdata)
				i := 0
				userKeys = func(delegatee model.SqlLike, query model.UserKeyQuery) ([]model.UserKey, error) {
					if i > 0 {
						return []model.UserKey{model.UserKey{}}, nil
					}
					i++
					return []model.UserKey{}, nil
				}
				return nil
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db = &test.MockDBLike{}
			want := tt.mock()
			got, err := NewCustomCols(tt.args.requests)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewCustomCols() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("NewCustomCols() = %#v, want %#v", got, want)
			}
		})
	}
}

func TestDeleteCustomCols(t *testing.T) {
	type args struct {
		enterprise string
		inputnames []string
	}
	tests := []struct {
		name    string
		args    args
		mock    func(args) int64
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				enterprise: "csbot",
				inputnames: []string{"test"},
			},
			mock: func(args args) int64 {
				var mockTotal int64 = 1
				deleteUserKey = func(delegatee model.SqlLike, query model.UserKeyQuery) (int64, error) {
					return mockTotal, nil
				}
				return mockTotal
			},
		},
		{
			name: "check empty inputname",
			args: args{
				enterprise: "csbot",
				inputnames: []string{},
			},
			mock: func(args args) int64 {
				return 0
			},
			wantErr: true,
		},
		{
			name: "check empty enterprise",
			args: args{
				enterprise: "",
				inputnames: []string{"1"},
			},
			mock: func(args args) int64 {
				return 0
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := tt.mock(tt.args)
			got, err := DeleteCustomCols(tt.args.enterprise, tt.args.inputnames...)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteCustomCols() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != want {
				t.Errorf("DeleteCustomCols() = %v, want %v", got, want)
			}
		})
	}
}
