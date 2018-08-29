package dal

import (
	"net/http"
	"testing"
)

var client, _ = NewClientWithHTTPClient("http://172.16.101.98:8885/", http.DefaultClient)

func TestNewClientWithHTTPClient(t *testing.T) {

	c, err := NewClientWithHTTPClient("http://127.0.0.1:8080/", http.DefaultClient)
	if err != nil {
		t.Fatal(err)
	}
	if c.address != "http://127.0.0.1:8080/dal" {
		t.Fatalf("expect address to be http://127.0.0.1:8080/dal but got %s", c.address)
	}
}
func TestSimilarQuestions(t *testing.T) {
	sq := "APP无法付款怎么办?"
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

func TestIsStandardQuestion(t *testing.T) {
	isStdQ, err := client.IsStandardQuestion("csbot", "APP无法付款怎么办?")
	if err != nil {
		t.Fatal(err)
	}
	if !isStdQ {
		t.Fatalf("expect this question is standard question, but got false")
	}
}

func TestSetSimilarQuestions(t *testing.T) {
	err := client.SetSimilarQuestion("csbot", "e家保理赔", "APP无法付款怎么办?")
	if err != nil {
		if detail, ok := err.(*DetailError); ok {
			t.Fatalf("got detail error, %s, results: %v", detail.Error(), detail.Results)
		}
		t.Fatal(err)
	}
}

func TestDeleteSimilarQuestion(t *testing.T) {
	err := client.DeleteSimilarQuestion("csbot", "APP无法付款怎么办?")
	if err != nil {
		if detail, ok := err.(*DetailError); ok {
			t.Fatalf("got detail error, %s, results: %v", detail.Error(), detail.Results)
		}
		t.Fatal(err)
	}

}
