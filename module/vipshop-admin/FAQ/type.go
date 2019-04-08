package FAQ

import (
	"fmt"
	"strings"

	"emotibot.com/emotigo/module/vipshop-admin/imagesManager"
	"emotibot.com/emotigo/module/vipshop-admin/util"
)

const (
	FOREVER_BEGIN = "1970-01-01 00:00:00"
	FOREVER_END   = "2999-12-31 23:59:00"
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
	Id      string `json:sqid`
}

type SimilarQuestionReqBody struct {
	SimilarQuestions []SimilarQuestion `json:similarQuestions`
}

//StdQuestion is a Standard Question in FAQ Table
type StdQuestion struct {
	QuestionID int    `json:"questionId"`
	Content    string `json:"content"`
	CategoryID int    `json:"categoryId"`
	Type       int    `json:"type"`
}

//Category represents sql table vipshop_category
type Category struct {
	ID       int
	Name     string
	ParentID int
	Children []int
}

// TODO: 這裡有兩種question 和 answer的 struct 是因為
//       php版本的 request和response 結構不一致
//       這次修改沒有時間統一 (要改的code 太多)
//       所以留下 TODO by Ken

type Question struct {
	QuestionId      int      `json:"questionId"`
	SQuestionConunt int      `json:"sQuesCount"`
	Content         string   `json:"questionContent"`
	CategoryName    string   `json:"categoryName"`
	CategoryId      int      `json:"categoryId"`
	Answers         []Answer `json:"answerItem"`
	User            string   `json:"createuser"`
	AppID           string
	Status          int `json:"status"`
}

func (q *Question) FetchAnswers() (err error) {
	if q.QuestionId == 0 || q.AppID == "" {
		err = fmt.Errorf("Has no QuestionId or AppID")
		return
	}

	answerDAO := Answer{
		QuestionId: q.QuestionId,
	}
	answerDAOs := []Answer{answerDAO}

	q.Answers, err = FindAnswers(q.AppID, answerDAOs)
	return
}

type Answer struct {
	QuestionId       int      `json:"Question_Id"`
	AnswerId         int      `json:"Answer_Id"`
	Content          string   `json:"Content_String"`
	RelatedQuestion  string   `json:"RelatedQuestion"`
	DynamicMenu      string   `json:"DynamicMenu"`
	NotShow          int      `json:"Not_Show_In_Relative_Q"`
	BeginTime        string   `json:"Begin_Time"`
	EndTime          string   `json:"End_Time"`
	AnswerCmd        string   `json:"Answer_CMD"`
	AnswerCmdMsg     string   `json:"Answer_CMD_Msg"`
	Dimension        []string `json:"dimension"`
	DimensionIDs     []int
	RelatedQuestions []string `json:"relatedQ"`
	DynamicMenus     []string `json:"dynamicMenu"`
	AppID            string
	Images           []int `json:"images"`
}

func (a *Answer) Fetch() (err error) {
	// fetch dimension, related questions, dynamic menu
	if a.AnswerId == 0 || a.AppID == "" {
		err = fmt.Errorf("Has no answer id or appid")
		return
	}

	db := util.GetMainDB()
	if db == nil {
		err = fmt.Errorf("main db connection pool is nil")
		return
	}

	targetLabel := AnswerLabelDAO{
		AnswerId: a.AnswerId,
	}
	targetLabels := []AnswerLabelDAO{targetLabel}

	// fetch related questions
	labels, err := FindAnswerLabels(a.AppID, RelatedQuestion, targetLabels)
	if err != nil {
		return
	}
	for _, label := range labels {
		a.RelatedQuestions = append(a.RelatedQuestions, label.Content)
	}

	// fetch dynamic menu
	labels, err = FindAnswerLabels(a.AppID, DynamicMenu, targetLabels)
	if err != nil {
		return
	}
	for _, label := range labels {
		a.DynamicMenus = append(a.DynamicMenus, label.Content)
	}

	// fetch dimension
	sqlStr := fmt.Sprintf(`SELECT vipshop_tag.Tag_Id, Tag_Name FROM 
	(
		select Tag_Id From vipshop_answertag where Answer_Id = %d
	) as atag
	left join vipshop_tag on atag.Tag_Id = vipshop_tag.Tag_Id
	left join vipshop_tag_type on vipshop_tag.Tag_Type = vipshop_tag_type.Type_id`, a.AnswerId)

	rows, err := db.Query(sqlStr)
	if err != nil {
		return
	}

	for rows.Next() {
		var tagID int
		var tagName string

		rows.Scan(&tagID, &tagName)
		a.DimensionIDs = append(a.DimensionIDs, tagID)
		a.Dimension = append(a.Dimension, strings.Replace(tagName, "#", "", -1))
	}

	// fetch images
	var answerID []interface{} = []interface{}{uint64(a.AnswerId)}
	relations, err := imagesManager.GetImageByAnswerID(answerID)
	if err != nil {
		return
	}
	for _, relation := range relations {
		infos := relation.Info
		for _, info := range infos {
			a.Images = append(a.Images, int(info.ImageID))
		}
	}

	return
}

type AnswerJson struct {
	ID               int `json:"id"`
	QuestionID       int
	Content          string   `json:"answer"`
	DynamicMenu      []string `json:"dynamicMenu"`
	RelatedQuestions []string `json:"relatedQ"`
	AnswerCMD        string   `json:"answerCMD"`
	AnswerCMDMsg     string   `json:"answerCMDMsg"`
	NotShow          bool     `json:"not_show_in_relative_q"`
	Dimension        []int    `json:"dimension"`
	BeginTime        string   `json:"begin_time"`
	EndTime          string   `json:"end_time"`
	Images           []int    `json:"images"`
}

type QuestionJson struct {
	Content          string       `json:"content"`
	CategoryID       int          `json:"categoryid"`
	SimilarQuestions []string     `json:"similarQuestions"`
	Answers          []AnswerJson `json:"answer_json"`
	User             string       `json:"createuser"`
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

type SimilarQuestionDAO struct {
	Qid     int
	Content string
	Sid     int
	Status  int
}

// this dao is used for both dynamic menu and relative question
type AnswerLabelDAO struct {
	Id       int
	AnswerId int
	Content  string
}

type Tag struct {
	Type    int
	Content string
}

//SubCat will recursivily retrive the sub Category of the Category
func (c Category) SubCats() ([]Category, error) {
	db := util.GetMainDB()
	if db == nil {
		return nil, fmt.Errorf("main db connection pool is nil")
	}
	rawQuery := "SELECT CategoryId, CategoryName FROM vipshop_categories WHERE ParentId = ? AND Status = 1"
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
		subSubCats, err := subCat.SubCats() //子類別的子類別
		if err != nil {
			return nil, fmt.Errorf("sub category %s query failed, %v", subCat.Name, err)
		}
		categories = append(categories, subSubCats...)
	}

	return categories, nil
}

// FullName will return complete name of category.
// the start prefix and seperator is slash
// ex: a->b->c, Category c's FullName will be a/b/c
func (c Category) FullName() (string, error) {
	db := util.GetMainDB()
	if db == nil {
		return "", fmt.Errorf("main db connection pool is nil")
	}
	rows, err := db.Query("SELECT CategoryId, CategoryName, ParentId FROM vipshop_categories")
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
	var fullPath string
	c, ok := categories[c.ID]
	if !ok {
		return "", fmt.Errorf("Cant find category id %d in db", c.ID)
	}
	switch c.ParentID {
	case 0:
		fallthrough
	case -1:
		return c.Name, nil
	default:
		fullPath = c.Name
	}
	for c, ok = categories[c.ParentID]; ok; c, ok = categories[c.ParentID] {
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
