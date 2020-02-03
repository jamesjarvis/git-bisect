package bisect

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

// Count is the score essentially
var Count int

// TheKnowledge is the actual information the server should know
var TheKnowledge Solutions

// AskQuestionJSON asks a question about this particular commit to the server, and updates the count
func AskQuestionJSON(s Solutions, q Question) Answer {
	Count++

	if contains(s.AllBad, q.Question) {
		return Answer{
			Answer: "Bad",
		}
	}

	return Answer{
		Answer: "Good",
	}
}

// SubmitSolutionJSON is the "endpoint" where you can submit a solution and retrieve your score
func SubmitSolutionJSON(s Solutions, attempt Solution) Score {
	if contains(s.AllBad, attempt.Solution) {
		log.Printf("(%v) is indeed BAD", attempt.Solution)
	} else {
		log.Printf("(%v) is actually GOOD", attempt.Solution)
	}

	if s.Bug != attempt.Solution {
		return Score{
			Score: -1,
		}
	}

	return Score{
		Score: Count,
	}
}

// GetBug returns the bug
func GetBug() string {
	return TheKnowledge.Bug
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// ConnectJSON does a "connection" to the "server", but json
func ConnectJSON() (Problem, error) {
	// filetotest := "/Users/jarjames/git/git-bisect/tests/test_linux0.json"
	// filetotest := "/Users/jarjames/git/git-bisect/tests/test_bootstrap0.json"
	filetotest := "/Users/jarjames/git/git-bisect/tests/test_bootstrap10.json"
	// filetotest := "/Users/jarjames/git/git-bisect/tests/test_react0.json"

	var prob Problem

	jsonFile, err := os.Open(filetotest)
	if err != nil {
		return prob, err
	}
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return prob, err
	}

	var file Root

	err = json.Unmarshal([]byte(byteValue), &file)
	if err != nil {
		return prob, err
	}

	TheKnowledge = file.Solutions

	return file.Problem, nil
}
