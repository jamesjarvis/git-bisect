package bisect

import "log"

// LastBadCommit is the last known bad commit
var LastBadCommit string

// GoodCommit changes the map with response to being a good commit
// New search space is it and it's ancestors
func GoodCommit(p map[string]DAGEntry, c string) map[string]DAGEntry {
	newsearchspace := StartGetParents(p, c)
	newsearchspace[c] = p[c]

	newsearchspace = RemoveMapFromMap(p, newsearchspace)

	log.Printf("Now has %v commits after GOOD (%v)\n", len(newsearchspace), c)

	return newsearchspace
}

// BadCommit changes the map with response to being a bad commit
// New search space is the old search space - it and it's ancestors
func BadCommit(p map[string]DAGEntry, c string) map[string]DAGEntry {
	LastBadCommit = c
	newsearchspace := StartGetParents(p, c)
	newsearchspace[c] = p[c]

	log.Printf("Now has %v commits after BAD (%v)\n", len(newsearchspace), c)

	return newsearchspace
}

// MidPoint gets a midpoint from the map
func MidPoint(p map[string]DAGEntry) string {
	for {
		temp := GetFirstElementFromMap(p)
		if temp != LastBadCommit && temp != "" {
			return temp
		}
	}
}

// NextMove actually contains the logic
func NextMove(m map[string]DAGEntry) Score {
	for {
		// IF the length is 0, submit the last "badcommit"
		if len(m) == 0 {
			return SubmitSolution(Solution{
				Solution: LastBadCommit,
			})
		}

		// IF the length is 1, submit the only one there
		if len(m) == 1 {
			return SubmitSolution(Solution{
				Solution: GetFirstElementFromMap(m),
			})
		}

		midpoint := MidPoint(m)

		// ELSE get midpoint and ask question
		answer := AskQuestion(Question{
			Question: midpoint,
		})

		switch answer.Answer {
		case "Good":
			m = GoodCommit(m, midpoint)
		case "Bad":
			m = BadCommit(m, midpoint)
		}
	}
}

// AskQuestion asks a question about this particular commit to the server, and updates the count
func AskQuestion(q Question) Answer {
	answer := AskQuestionJSON(TheKnowledge, q)

	log.Printf("QUESTION: (%v) is (%v)", q.Question, answer.Answer)

	return answer
}

// SubmitSolution is the "endpoint" where you can submit a solution and retrieve your score
func SubmitSolution(attempt Solution) Score {
	return SubmitSolutionJSON(TheKnowledge, attempt)
}

// Connect connects to the server
func Connect() Problem {
	return ConnectJSON()
}
