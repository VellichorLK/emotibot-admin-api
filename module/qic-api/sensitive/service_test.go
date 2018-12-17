package sensitive

 import "testing"

 type mockDAO struct{}

 func (dao *mockDAO) GetSensitiveWords() ([]string, error) {
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

 var mockdao sensitiveDao = &mockDAO{}

 func TestIsSensitive(t *testing.T) {
	 sen1 := "我要保本保息"
	 sen2 := "一個安全的句子"
	 sen3 := "要不要理财型保险"

	 sen1Result, _ := IsSensitive(sen1)
	 sen2Result, _ := IsSensitive(sen2)
	 sen3Result, _ := IsSensitive(sen3)

	 if len(sen1Result) == 0 || len(sen2Result) > 0 || len(sen3Result) == 0 {
		 t.Error("check sensitive words fail")
	 }
}

func TestStringsToRunes(t *testing.T) {
	ss, _ := mockdao.GetSensitiveWords()
	words := stringsToRunes(ss)

	if len(words) != len(ss) {
		t.Error("tranforms strings to runes failed")
	}
}