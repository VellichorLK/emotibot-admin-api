package sensitive

type sensitiveDao interface {
	GetSensitiveWords() ([]string, error)
}

type sensitiveDAOImpl struct{}

func (dao *sensitiveDAOImpl) GetSensitiveWords() ([]string, error) {
	return []string{
		"收益",
	}, nil
}
