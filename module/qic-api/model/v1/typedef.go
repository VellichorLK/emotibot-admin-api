package model

type GroupFilter struct {
	FileName      string
	Deal          int
	Series        string
	CallStart     int64
	CallEnd       int64
	StaffID       string
	StaffName     string
	Extension     string
	Department    string
	CustomerID    string
	CustomerName  string
	CustomerPhone string
	Page          int
	Limit         int
}
