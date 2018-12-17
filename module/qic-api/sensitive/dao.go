package sensitive

type sensitiveDao interface {
	GetSensitiveWords() ([]string, error)
}

type sensitiveDAOImpl struct{}

func (dao *sensitiveDAOImpl) GetSensitiveWords() ([]string, error){
	return []string{
		"本金",
		"利息",
		"存款", 
		"取本",
		"存",
		"取",
		"保本保息",
		"收益",
		"理财",
		"理财功能的保险", 
		"理财型保险",
	 }, nil
}