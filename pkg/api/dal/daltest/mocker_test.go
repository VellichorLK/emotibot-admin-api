package daltest

import (
	"testing"
)

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

// TestMockerExpectIsSimilarQuestion test mocker should be able inject result into ExpectResult
func TestMockerExpectIsSimilarQuestion(t *testing.T) {
	client, mocker, _ := New()
	mocker.ExpectIsSimilarQuestion("csbot", "subjectA").WillReturn(true)
	isSimQ, err := client.IsSimilarQuestion("csbot", "subjectA")
	if err != nil {
		t.Fatal("expect no error but got, ", err)
	}
	if isSimQ != true {
		t.Fatal("expect value to be true, but got false")
	}
	mocker.ExpectIsSimilarQuestion("csbot", "subjectA").WillReturn(false)
	isSimQ, err = client.IsSimilarQuestion("csbot", "subjectA")
	if err != nil {
		t.Fatal("expect no error but got ", err)
	}
	if isSimQ {
		t.Fatal("expect value to be false, but got true")
	}
	mocker.ExpectIsSimilarQuestion("csbot", "subjectA").WillFail()
	_, err = client.IsSimilarQuestion("csbot", "subjectA")
	if err == nil {
		t.Fatal("expect to failed but got err == nil")
	}
}

func TestMockerExpectIsStandardQuestion(t *testing.T) {
	client, mocker, _ := New()
	mocker.ExpectIsStandardQuestion("csbot", "Test").WillReturn(true)
	isStdQ, err := client.IsStandardQuestion("csbot", "Test")
	if err != nil {
		t.Fatal("expect no error but got ", err)
	}
	if !isStdQ {
		t.Fatal("expect value to be true but got false")
	}
	mocker.ExpectIsStandardQuestion("csbot", "Test2").WillReturn(false)
	isStdQ, err = client.IsStandardQuestion("csbot", "Test2")
	if err != nil {
		t.Fatal("expect no error but got ", err)
	}
	if isStdQ {
		t.Fatal("expect value to be false but got true")
	}
}

func TestMockerExpectQuestions(t *testing.T) {
	client, mocker, _ := New()
	var output = []interface{}{"A", "B", "C"}
	result := mocker.ExpectQuestions("csbot")
	result.WillReturn(output...)
	questions, err := client.Questions("csbot")
	if err != nil {
		t.Fatal(err)
	}
	for i, q := range questions {
		if output[i] != q {
			t.Fatal("expect mock output ", output[i], ", but got", q)
		}
	}
}

func TestMockerExpectQuestion(t *testing.T) {
	client, mocker, _ := New()
	var output = "A"
	mocker.ExpectQuestion("csbot", "test").WillReturn(output)
	q, err := client.Question("csbot", "test")
	if err != nil {
		t.Fatal(err)
	}
	if q != output {
		t.Fatal("expect output to be", output, ", but got ", q)
	}
	mocker.ExpectQuestion("csbot", "test2").WillFail()
	_, err = client.Question("csbot", "test")
	if err == nil {
		t.Fatal("expect error but got nil")
	}
}

func TestMockerExpectSimilarQuestion(t *testing.T) {
	client, mocker, _ := New()
	var output = []interface{}{"lqA", "lqB"}
	mocker.ExpectSimilarQuestions("csbot", "stdQ").WillReturn(output...)
	results, err := client.SimilarQuestions("csbot", "stdQ")
	if err != nil {
		t.Fatal("got error: ", err)
	}
	for i, r := range results {
		if output[i] != r {
			t.Fatal("expect output ", output[i], ", but got ", r)
		}
	}
	mocker.ExpectSimilarQuestions("csbot", "stdQ").WillFail()
	_, err = client.SimilarQuestions("csbot", "stdQ")
	if err == nil {
		t.Fatal("expect error but got nil")
	}

}
