package daltest

import "testing"

// TestMultipleExpect should Mocker can do multiple command without error.
func TestMultipleExpect(t *testing.T) {
	//TODO: implement me.
}

// TestMockerExpectDeleteSimilarQuestions test ExpectDeleteSimilarQuestions behavior
// Not test result yet.
func TestMockerExpectDeleteSimilarQuestions(t *testing.T) {
	client, mocker, _ := New()
	mocker.ExpectDeleteSimilarQuestions("csbot", "test")
	err := client.DeleteSimilarQuestions("csbot", "test")
	if err != nil {
		t.Fatal(err)
	}
}

//TestMockShouldFailAtWrongExpects test mocker should fail if expect were not met.
func TestMockShouldFailAtWrongExpects(t *testing.T) {
	client, mocker, _ := New()
	mocker.ExpectDeleteSimilarQuestions("csbot", "test")
	err := client.DeleteSimilarQuestions("csbot", "test")
	if err != nil {
		t.Fatal(err)
	}
	err = client.DeleteSimilarQuestions("WTF", "YA")
	if err == nil {
		t.Fatal("only one expect called with two actual behavior should produce error but got no one.")
	}
}

func TestMockerExpectIsSimilarQuestion(t *testing.T) {
	client, mocker, _ := New()
	mocker.ExpectIsSimilarQuestion("csbot", "subjectA").WillReturn(nil, []string{"OK"})
	isSimQ, err := client.IsSimilarQuestion("csbot", "subjectA")
	if err != nil {
		t.Fatal("expect no error but got, ", err)
	}
	if isSimQ != true {
		t.Fatal("expect is sim q is true, but got ", isSimQ)
	}
}
