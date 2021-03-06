package FAQ

import (
	"fmt"
	"time"

	"emotibot.com/emotigo/module/admin-api/util"
)

var (
	statusNormal  = 0
	statusDelete  = -1
	statusUpdated = 1
)

type APICategory struct {
	ParentID int            `json:"fid"`
	ID       int            `json:"id"`
	Level    int            `json:"level"`
	Path     string         `json:"filepath"`
	Name     string         `json:"text"`
	Children []*APICategory `json:"children"`
}

type SimilarQuestion struct {
	Content string `json:"content"`
	Id      string `json:"sqid"`
}

type SimilarQuestionReqBody struct {
	SimilarQuestions []SimilarQuestion `json:"similarQuestions"`
}

//StdQuestion is a Standard Question in FAQ Table
type StdQuestion struct {
	QuestionID int    `json:"questionId"`
	Content    string `json:"content"`
	CategoryID int    `json:"categoryId"`
}

//Category represents sql table <appid>_category
type Category struct {
	ID       int
	Name     string
	ParentID int
	Children []int
}

type Question struct {
	QuestionId      int      `json:"questionId"`
	SQuestionConunt int      `json:"sQuesCount"`
	Content         string   `json:"questionContent"`
	CategoryName    string   `json:"categoryName"`
	CategoryId      int      `json:"categoryId"`
	Answers         []Answer `json:"answerItem"`
}

type Answer struct {
	QuestionId      int            `json:"Question_Id"`
	AnswerId        int            `json:"Answer_Id"`
	Content         string         `json:"Content_String"`
	RelatedQuestion string         `json:"RelatedQuestion"`
	DynamicMenu     string         `json:"DynamicMenu"`
	NotShow         int            `json:"Not_Show_In_Relative_Q"`
	BeginTime       string         `json:"Begin_Time"`
	EndTime         string         `json:"End_Time"`
	AnswerCmd       string         `json:"Answer_CMD"`
	AnswerCmdMsg    string         `json:"Answer_CMD_Msg"`
	Dimension       []string       `json:"dimension"`
	DimensionMap    map[int]string `json:"dimension_map"`
	Label           string         `json:"label"`
}

type QueryCondition struct {
	TimeSet                bool
	BeginTime              string
	EndTime                string
	Keyword                string
	SearchQuestion         bool
	SearchAnswer           bool
	SearchDynamicMenu      bool
	SearchRelativeQuestion bool
	SearchAll              bool
	NotShow                bool
	Dimension              []DimensionGroup
	CategoryId             int
	Limit                  int
	CurPage                int
}

type DimensionGroup struct {
	TypeId  int    `json:"typeId"`
	Content string `json:"tagContent"`
}

type Parameter interface {
	FormValue(name string) string
}

// Tag means dimension in UI
type Tag struct {
	Type    int
	Content string
}

// Label means activity tag in UI
type Label struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	RuleCount int    `json:"rule_count"`
}

//RFQuestion is removed Feedback question(移除解決未解決的問題)
type RFQuestion struct {
	ID         int    `json:"id"`
	Content    string `json:"content"`
	CategoryID int    `json:"categoryId"`
	IsValid    bool   `json:"isValid"`
}

//UpdateRFQUestionsArgs are Post API JSON arguments
type UpdateRFQuestionsArgs struct {
	Contents []string `json:"contents"`
}

//SubCat will recursivily retrive the sub Category of the Category
func (c Category) SubCats(appid string) ([]Category, error) {
	db := util.GetMainDB()
	if db == nil {
		return nil, fmt.Errorf("main db connection pool is nil")
	}
	rawQuery := fmt.Sprintf(`
		SELECT CategoryId, CategoryName
		FROM %s_categories
		WHERE ParentId = ? AND Status = 1`, appid)
	rows, err := db.Query(rawQuery, c.ID)
	if err != nil {
		return nil, fmt.Errorf("sql query %s failed %v", rawQuery, err)
	}
	defer rows.Close()
	var categories []Category
	for rows.Next() {
		var subCat Category
		subCat.ParentID = c.ID
		if err := rows.Scan(&subCat.ID, &subCat.Name); err != nil {
			return nil, fmt.Errorf("scan failed, %v", err)
		}
		categories = append(categories, subCat)
		subSubCats, err := subCat.SubCats(appid) //子類別的子類別
		if err != nil {
			return nil, fmt.Errorf("sub category %s query failed, %v", subCat.Name, err)
		}
		categories = append(categories, subSubCats...)
	}

	return categories, nil
}

// FullName will return complete name of category.
// the start prefix and seperator is slash
// ex: a->b->c, Category c's FullName will be /a/b/c
func (c Category) FullName(appid string) (string, error) {
	db := util.GetMainDB()
	if db == nil {
		return "", fmt.Errorf("main db connection pool is nil")
	}
	rows, err := db.Query(fmt.Sprintf("SELECT CategoryId, CategoryName, ParentId FROM %s_categories", appid))
	if err != nil {
		return "", fmt.Errorf("query category table failed, %v", err)
	}
	defer rows.Close()
	var categories = make(map[int]Category)

	for rows.Next() {
		var c Category
		rows.Scan(&c.ID, &c.Name, &c.ParentID)
		categories[c.ID] = c
	}
	if err = rows.Err(); err != nil {
		return "", fmt.Errorf("Rows scaning failed, %v", err)
	}

	if c, ok := categories[c.ID]; ok {
		switch c.ParentID {
		case 0:
			fallthrough
		case -1:
			return c.Name, nil
		}
		var fullPath string
		for ; ok; c, ok = categories[c.ParentID] {
			fullPath = c.Name + "/" + fullPath
			if c.ParentID == 0 {
				break
			}
		}
		if !ok {
			return "", fmt.Errorf("category id %d has invalid parentID %d", c.ID, c.ParentID)
		}
		return fullPath, nil
	}
	return "", fmt.Errorf("Cant find category id %d in db", c.ID)
}

type RuleTarget int

const (
	TargetStandardQuestion RuleTarget = iota
	TargetAnswer
	TargetMax
)

func (RuleTarget) Max() int {
	return int(TargetMax) - 1
}

type ResponseType int

const (
	TypeReplace ResponseType = iota
	TypeAppendFront
	TypeAppendBehind
	TypeAppendMax
)

func (ResponseType) Max() int {
	return int(TypeAppendMax) - 1
}

type RuleContent struct {
	// Type only allow keyword or regex
	Type  string   `json:"type"`
	Value []string `json:"value"`
}

type Rule struct {
	ID        int            `json:"id"`
	Name      string         `json:"name"`
	Target    RuleTarget     `json:"target"`
	Rule      []*RuleContent `json:"rule"`
	Answer    string         `json:"answer"`
	Type      ResponseType   `json:"response_type"`
	Status    bool           `json:"status"`
	Begin     *time.Time     `json:"begin_time"`
	End       *time.Time     `json:"end_time"`
	LinkLabel []int          `json:"labels"`
}

func (r RuleContent) IsValid() bool {
	return (r.Type == "keyword" || r.Type == "regex") && len(r.Value) > 0
}

type TagType struct {
	ID     int         `json:"id"`
	Name   string      `json:"name"`
	Code   string      `json:"code"`
	Values []*TagValue `json:"values"`
}

type TagValue struct {
	ID    int    `json:"id"`
	Value string `json:"value"`
	Code  string `json:"code"`
}
