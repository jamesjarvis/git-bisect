package bisect

import (
	"log"

	"github.com/jamesjarvis/git-bisect/pkg/dag"
)

// // LastBadCommit is the last known bad commit
// var LastBadCommit string

// // GoodCommit changes the map with response to being a good commit
// // New search space is the old search space - it and it's ancestors
// func GoodCommit(dag *dag.DAG, c string) *dag.DAG {

// 	ances, err := dag.GetOrderedAncestors(c)
// 	if err != nil {
// 		log.Fatalf("Failed to get ancestors of (%v)", c)
// 	}

// 	log.Printf("Size of ordered ancestors for (%v): %v", c, len(ances))

// 	newsearchspace := dag.GoodSearchSpace(ances, c)

// 	log.Printf("Now has %v commits after GOOD (%v)\n", newsearchspace.GetOrder(), c)

// 	return newsearchspace
// }

// // BadCommit changes the map with response to being a bad commit
// // New search space is it and it's ancestors
// func BadCommit(dag *dag.DAG, c string) *dag.DAG {
// 	LastBadCommit = c
// 	ances, err := dag.GetOrderedAncestors(c)
// 	if err != nil {
// 		log.Fatalf("Failed to get ancestors of (%v)", c)
// 	}

// 	newsearchspace := dag.BadSearchSpace(ances, c)

// 	log.Printf("Now has %v commits after BAD (%v)\n", newsearchspace.GetOrder(), c)

// 	return newsearchspace
// }

// // MidPoint gets a midpoint from the map
// func MidPoint(p map[string]DAGEntry) string {
// 	maxlen := EstimateMaxAncestry(p)
// 	maxattemps := 100
// 	var temp string
// 	for i := 0; i < maxattemps; i++ {
// 		temp = GetFirstElementFromMap(p)
// 		if temp != LastBadCommit && temp != "" {
// 			lenParents := StartGetLengthParents(p, temp)
// 			if lenParents > 0 {
// 				similarity := (float64(maxlen) / 2.0) / float64(StartGetLengthParents(p, temp))
// 				if similarity <= 1.25 && similarity >= 0.75 {
// 					return temp
// 				}
// 			}
// 		}
// 	}
// 	return temp
// }

// NextMove actually contains the logic
func NextMove(d *dag.DAG) Score {
	for {
		// IF the length is 0, submit the last "badcommit"
		if d.GetOrder() == 0 {
			log.Printf("Submitting MOST RECENT (%v)\n", d.MostRecentBad)
			return SubmitSolution(Solution{
				Solution: d.MostRecentBad,
			})
		}

		// IF the length is 1, submit the only one there
		if d.GetOrder() == 1 {
			submission := GetFirstElementFromMap(d.GetVertices())
			log.Printf("Submitting LAST LEFT (%v)\n", submission)
			return SubmitSolution(Solution{
				Solution: submission,
			})
		}

		midpoint := d.MidPoint

		// ELSE get midpoint and ask question
		answer := AskQuestion(Question{
			Question: midpoint,
		})

		switch answer.Answer {
		case "Good":
			err := d.GoodCommit(midpoint)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Now %v commits after GOOD (%v)\n", d.GetOrder(), midpoint)
		case "Bad":
			err := d.BadCommit(midpoint)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Now %v commits after BAD (%v)\n", d.GetOrder(), midpoint)
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
