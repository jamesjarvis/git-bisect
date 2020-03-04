package bisect

// import (
// 	"encoding/json"
// 	"fmt"
// 	"io/ioutil"
// 	"os"
// 	"path/filepath"
// )

// // Count is the score essentially
// var Count int

// // TheKnowledge is the actual information the server should know
// var TheKnowledge Solutions

// type attempt struct {
// 	count    int
// 	solution Solutions
// 	problem  Problem
// 	done     bool
// 	score    int
// }

// // ConnectionJSON lol
// type ConnectionJSON struct {
// 	Problems    map[string]attempt
// 	CurrentFile string
// }

// // AskQuestionJSON asks a question about this particular commit to the server, and updates the count
// func (c *ConnectionJSON) AskQuestionJSON(q Question) (Answer, error) {
// 	var ans Answer

// 	att, ok := c.Problems[c.CurrentFile]
// 	if !ok {
// 		return ans, fmt.Errorf("Current file %v not found", c.CurrentFile)
// 	}

// 	att.count++

// 	c.Problems[c.CurrentFile] = att

// 	if contains(att.solution.AllBad, q.Question) {
// 		return Answer{
// 			Answer: "Bad",
// 		}, nil
// 	}

// 	return Answer{
// 		Answer: "Good",
// 	}, nil
// }

// // SubmitSolutionJSON is the "endpoint" where you can submit a solution and retrieve your score
// func (c *ConnectionJSON) SubmitSolutionJSON(attempt Solution) (Score, Problem, error) {
// 	// Get the current attempt
// 	currentAttempt := c.Problems[c.CurrentFile]

// 	// Check the attempt
// 	if contains(currentAttempt.solution.AllBad, attempt.Solution) {
// 		// log.Printf("(%v) is indeed BAD", attempt.Solution)
// 	} else {
// 		// log.Printf("(%v) is actually GOOD", attempt.Solution)
// 	}

// 	// Update the score for that attempt and wipe the memory
// 	if currentAttempt.solution.Bug != attempt.Solution {
// 		currentAttempt.score = -1
// 	} else {
// 		currentAttempt.score = currentAttempt.count
// 	}
// 	currentAttempt.done = true
// 	currentAttempt.solution = Solutions{}
// 	currentAttempt.problem = Problem{}
// 	c.Problems[c.CurrentFile] = currentAttempt

// 	newProblem, err := c.GetProblemJSON()
// 	if err != nil {
// 		scores := Score{
// 			Score: make(map[string]int),
// 		}
// 		for str, prob := range c.Problems {
// 			scores.Score[str] = prob.score
// 		}

// 		return scores, Problem{}, nil
// 	}

// 	return Score{}, newProblem, nil
// }

// func contains(s []string, e string) bool {
// 	for _, a := range s {
// 		if a == e {
// 			return true
// 		}
// 	}
// 	return false
// }

// func filepathwalkdir(root string) ([]string, error) {
// 	var files []string
// 	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
// 		if !info.IsDir() {
// 			files = append(files, path)
// 		}
// 		return nil
// 	})
// 	return files, err
// }

// // ConnectJSON does a "connection" to the "server", but json
// func ConnectJSON(dir string) (ConnectionJSON, error) {

// 	connection := ConnectionJSON{
// 		Problems: make(map[string]attempt),
// 	}

// 	files, err := filepathwalkdir(dir)
// 	if err != nil {
// 		return connection, err
// 	}

// 	for _, file := range files {
// 		connection.Problems[file] = attempt{
// 			done: false,
// 		}
// 	}

// 	return connection, nil
// }

// // GetProblemJSON goes through the available files
// // If it hasnt been done, it loads the file into the connection
// // If it doesnt find a file to choose, then it returns an error
// func (c *ConnectionJSON) GetProblemJSON() (Problem, error) {
// 	var prob Problem

// 	for fileID, att := range c.Problems {
// 		if !att.done {
// 			c.CurrentFile = fileID

// 			jsonFile, err := os.Open(fileID)
// 			if err != nil {
// 				return prob, err
// 			}
// 			byteValue, err := ioutil.ReadAll(jsonFile)
// 			if err != nil {
// 				return prob, err
// 			}
// 			var file Root
// 			err = json.Unmarshal([]byte(byteValue), &file)
// 			if err != nil {
// 				return prob, err
// 			}

// 			att.problem = file.Problem
// 			att.solution = file.Solutions

// 			c.Problems[fileID] = att

// 			return att.problem, nil
// 		}
// 	}
// 	return prob, NoNewProblems{}
// }

// // NoNewProblems is the error type to describe the situation, that a given
// // edge does not exit in the graph.
// type NoNewProblems struct {
// }

// // Implements the error interface.
// func (e NoNewProblems) Error() string {
// 	return "No new problems found"
// }

// // NextMove actually contains the logic
// func (c *ConnectionJSON) NextMove(d *dag.DAG, pc dag.ParamConfig, idealScores map[string]float64) (Score, error) {
// 	var s Score
// 	for {

// 		// IF the length is 0, submit the last "badcommit"
// 		if d.GetOrder() == 0 {
// 			// log.Printf("Submitting MOST RECENT (%v)\n", d.MostRecentBad)
// 			s, p, err := c.SubmitSolutionJSON(Solution{
// 				Solution: d.MostRecentBad,
// 			})
// 			if err != nil {
// 				return Score{}, err
// 			}
// 			if p.Name == "" { // Actually complete
// 				s.IdealScores = CopyMapFloat(idealScores)
// 				// log.Printf("Size of the ideal scores: %v, %v", idealScores, s.IdealScores)
// 				return s, err
// 			}

// 			// Else, restart with the new problem

// 			d = DAGMaker(&p)

// 			log.Printf("Problem: %v has %v vertexes (commits) and %v edges\n", p.Name, d.GetOrder(), d.GetSize())

// 			err = d.GoodCommit(p.Good)
// 			if err != nil {
// 				return s, err
// 			}

// 			// log.Printf("Problem: %v now has %v commits after GOOD (%v)\n", p.Name, d.GetOrder(), p.Good)

// 			err = d.BadCommit(p.Bad)
// 			if err != nil {
// 				return s, err
// 			}

// 			idealScores[p.Name] = math.Log2(float64(d.GetOrder()))
// 			// log.Printf("Stored new ideal score: %v : %v", p.Name, idealScores[p.Name])

// 			// log.Printf("Problem: %v now has %v commits after BAD (%v)\n", p.Name, d.GetOrder(), p.Bad)

// 		}

// 		midpoint, err := d.GetMidPoint(pc)
// 		if err != nil {
// 			return s, err
// 		}

// 		question := Question{
// 			Question: midpoint,
// 		}

// 		// ELSE get midpoint and ask question
// 		answer, err := c.AskQuestionJSON(question)
// 		if err != nil {
// 			return s, err
// 		}

// 		switch answer.Answer {
// 		case "Good":
// 			err := d.GoodCommit(question.Question)
// 			if err != nil {
// 				return s, err
// 			}
// 			// log.Printf("Now %v commits after GOOD (%v)\n", d.GetOrder(), question.Question)
// 		case "Bad":
// 			err := d.BadCommit(question.Question)
// 			if err != nil {
// 				return s, err
// 			}
// 			// log.Printf("Now %v commits after BAD (%v)\n", d.GetOrder(), question.Question)
// 		}
// 	}
// }