/*
Package daltest give support for mocking dal package without using httptest package.
It should eliminate the need for developer to understand the dal networking or api endpoint.
List of supported commands:
	- DeleteSimilarQuestions(appID string, lq ...string) error
	- IsSimilarQuestion(appID, lq string) (bool, error)
	- IsStandardQuestion(appID, content string) (bool, error)
	- Question(appID, lq string) (string, error)
	- Questions(appID string) ([]string, error)
	- SetSimilarQuestion(appID, sq string, lq ...string) error
	- SimilarQuestions(appID string, sq string) ([]string, error)
*/
package daltest

import (
	"fmt"
	"net/http"
	"sync"

	"emotibot.com/emotigo/pkg/api/dal/v1"
)

// Mocker is the mocking command center, which can be called to prepare the query.
// Due to internal struct, It need serialized access for command, or the result may not be correct.
// It haven't support random access command and inspect arguments yet.
type Mocker struct {
	expectCommands []mockCommand
}

// New create the mocker and the injected dal Client, or the somehow error return from dal package.
func New() (*dal.Client, *Mocker, error) {
	var m = &Mocker{
		expectCommands: make([]mockCommand, 0),
	}
	dalClient, err := dal.NewClientWithHTTPClient("http://127.0.0.1", m.newHTTPClient())
	return dalClient, m, err
}

func (m *Mocker) newHTTPClient() *mockClient {
	return &mockClient{
		mu:     sync.Mutex{},
		mocker: m,
	}
}

//mockClient is the client implements HTTPClient interface, which can be inject into dal client.
type mockClient struct {
	mu sync.Mutex
	//To access the mocked command
	mocker *Mocker
}

//Do is intend for mocking http Client, It use mutex lock to insure possible .
func (m *mockClient) Do(*http.Request) (*http.Response, error) {
	expectCommands := m.mocker.expectCommands
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(expectCommands) == 0 {
		return nil, fmt.Errorf("All Expect is satisfied, but still got called")
	} else if len(expectCommands) == 1 {
		command := expectCommands[0]
		m.mocker.expectCommands = make([]mockCommand, 0)
		return command.NewResponse(), nil
	}
	command := expectCommands[0]
	//Pop one command
	m.mocker.expectCommands = expectCommands[1 : len(expectCommands)-1]
	return command.NewResponse(), nil
}

// ExpectDeleteSimilarQuestions expect delete similar question(相似問、擴寫) from dal client.
// Since we can not verify the result yet, the input is not used.
func (m *Mocker) ExpectDeleteSimilarQuestions(appID string, lq ...string) *ExpectResult {
	var result = &ExpectResult{}
	var test = &deleteSimilarQuestionsCmd{
		execCommand: &execCommand{result: result},
		appID:       appID,
		lq:          lq,
	}
	m.expectCommands = append(m.expectCommands, test)
	return result
}

// ExpectIsSimilarQuestion expect matching similar question from dal client
// It will return a result for injecting match result.
// Result should only contains one bool value, anymore will be ignored.
// If result is empty or not bool, result will be false.
func (m *Mocker) ExpectIsSimilarQuestion(appID, lq string) (result *ExpectResult) {
	result = &ExpectResult{}
	m.expectCommands = append(m.expectCommands, &matchCommand{result: result})
	return result
}

// ExpectIsStandardQuestion expect matching content with standard questions from dal client.
// It will return result for injecting match result.
// Result should only contains one bool value, anymore will be ignored.
// If result is empty or not bool, result will be false.
func (m *Mocker) ExpectIsStandardQuestion(appID, content string) (result *ExpectResult) {
	result = &ExpectResult{}
	m.expectCommands = append(m.expectCommands, &matchCommand{result: result})
	return result
}

// ExpectQuestion expect query question based on lq(similar question).
// Result should be the content of expect question.
func (m *Mocker) ExpectQuestion(appID, lq string) (result *ExpectResult) {
	result = &ExpectResult{}
	m.expectCommands = append(m.expectCommands, &questionByLQCommand{result: result})
	return result
}

// ExpectQuestions expect query of all questions based on app ID.
// Result should be all the content of expect questions.
func (m *Mocker) ExpectQuestions(appID string) (result *ExpectResult) {
	result = &ExpectResult{}
	m.expectCommands = append(m.expectCommands, &queryQuestionsCommand{result: result})
	return result
}

// ExpectSimilarQuestions expect query of lq based on sq(standard question)
func (m *Mocker) ExpectSimilarQuestions(appID string, sq string) (result *ExpectResult) {
	result = &ExpectResult{}
	m.expectCommands = append(m.expectCommands, &queryQuestionsCommand{result: result})
	return result
}

// ExpectSetSimilarQuestion expect create lq based on sq & app ID.
// Result should be a single true of false.
func (m *Mocker) ExpectSetSimilarQuestion(appID, sq string, lq ...string) (result *ExpectResult) {
	result = &ExpectResult{}
	m.expectCommands = append(m.expectCommands, &execCommand{result: result})
	return result
}
