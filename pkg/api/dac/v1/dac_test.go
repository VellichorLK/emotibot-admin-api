package v1

import (
	"net/http"
	"testing"
)

var client, _ = NewClientWithHTTPClient("http://172.16.101.47:8686/", http.DefaultClient)


func TestIsStandardQuestion(t *testing.T) {
	isStdQ, err := client.IsStandardQuestion("csbot", "上门自取需要什么证件")
	if err != nil {
		t.Fatal(err)
	}
	if !isStdQ {
		t.Fatalf("expect this question is standard question, but got false")
	}
}

func TestIsSimilarQuestion(t *testing.T) {
	isLQ, err := client.IsSimilarQuestion("csbot", "丰巢件是否可以申请电子发票1111122222")
	if err != nil {
		t.Fatal(err)
	}
	if !isLQ {
		t.Fatalf("expect this question is similar question, but got false")
	}
}

func TestGetQuestions(t *testing.T) {
	questions, err := client.Questions("csbot")
	if err != nil {
		t.Fatal(err)
	}
	if len(questions) == 0 {
		t.Fatal("expect questions be greater than zero.")
	}
	// t.Fatal(questions)
}

func TestSimilarQuestions(t *testing.T) {
	sq := "丰巢件是否可以申请电子发票"
	lqGroup, err := client.SimilarQuestions("csbot", sq)
	if err != nil {
		t.Fatal(err)
	}
	for _, lq := range lqGroup {
		if lq == sq {
			t.Fatalf("expect sq %s not equal to lq %s", sq, lq)
		}
	}
}

func TestSetSimilarQuestions(t *testing.T) {
	err := client.SetSimilarQuestion("csbot", "香港件体积重量有什么要求", "香港件体积重量有什么要求122222")
	if err != nil {
		if detail, ok := err.(*DetailError); ok {
			t.Fatalf("got detail error, %s, results: %v", detail.Error(), detail.Results)
		}
		t.Fatal(err)
	}
}

func TestQuestion(t *testing.T) {
	question := "上门自取需要什么证件"
	sq, err := client.Question("csbot", "上门自取需要什么证件")
	if err != nil {
		t.Fatal(err)
	}
	if sq != question {
		t.Fatalf("expect sq to be %s but got %s", question, sq)
	}
}

func TestDeleteSimilarQuestions(t *testing.T) {
	err := client.DeleteSimilarQuestions("csbot", "香港件体积重量有什么要求122222")
	if err != nil {
		if detail, ok := err.(*DetailError); ok {
			t.Fatalf("got detail error, %s, results: %v", detail.Error(), detail.Results)
		}
		t.Fatal(err)
	}
}

