package bisect

import (
	"log"

	"github.com/jamesjarvis/git-bisect/pkg/dag"
)

// NextMove actually contains the logic
func NextMove(d *dag.DAG, actualbug string) Score {
	for {

		// IF the length is 0, submit the last "badcommit"
		if d.GetOrder() == 0 {
			log.Printf("Submitting MOST RECENT (%v)\n", d.MostRecentBad)
			return SubmitSolution(Solution{
				Solution: d.MostRecentBad,
			})
		}

		midpoint, err := d.GetMidPoint()
		if err != nil {
			log.Fatal(err)
		}

		question := Question{
			Question: midpoint,
		}

		// ELSE get midpoint and ask question
		answer := AskQuestion(question)

		switch answer.Answer {
		case "Good":
			err := d.GoodCommit(question.Question)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Now %v commits after GOOD (%v)\n", d.GetOrder(), question.Question)
		case "Bad":
			err := d.BadCommit(question.Question)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Now %v commits after BAD (%v)\n", d.GetOrder(), question.Question)
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
func Connect() (Problem, error) {
	return ConnectJSON()
}
