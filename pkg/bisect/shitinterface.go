package bisect

import (
	"log"

	"github.com/jamesjarvis/git-bisect/pkg/dag"
)

// CopyMapFloat copies a string, float map to avoid any potential issues?
func CopyMapFloat(in map[string]float64) map[string]float64 {
	out := make(map[string]float64)
	for key, value := range in {
		out[key] = value
	}
	return out
}

// NextMoveWebsocket actually contains the logic
func (c *Connection) NextMoveWebsocket(d *dag.DAG, pc dag.ParamConfig, problemInstance ProblemInstance) (Score, error) {
	var s Score
	for {

		// IF the length is 0, submit the last "badcommit"
		if d.GetOrder() == 0 {
			// log.Printf("Submitting MOST RECENT (%v)\n", d.MostRecentBad)
			var err error
			s, problemInstance, err = c.SubmitSolutionWebsocket(Solution{
				Solution: d.MostRecentBad,
			}, problemInstance)
			if err != nil {
				return Score{}, err
			}
			if problemInstance.Repo.Name == "" {
				return s, err
			}

			// Else, restart with the new problem

			d = DAGMaker(&problemInstance.Repo)

			log.Printf("Problem: %v has %v vertexes (commits) and %v edges\n", problemInstance.Repo.Name, d.GetOrder(), d.GetSize())
			log.Printf("Instance's GOOD: %v, BAD: %v", problemInstance.Instance.Good, problemInstance.Instance.Bad)

			err = d.GoodCommit(problemInstance.Instance.Good)
			if err != nil {
				return s, err
			}

			// log.Printf("Problem: %v now has %v commits after GOOD (%v)\n", p.Name, d.GetOrder(), p.Good)

			err = d.BadCommit(problemInstance.Instance.Bad)
			if err != nil {
				return s, err
			}

			// log.Printf("Problem: %v now has %v commits after BAD (%v)\n", p.Name, d.GetOrder(), p.Bad)

		}

		midpoint, err := d.GetMidPoint(pc)
		if err != nil {
			return s, err
		}

		question := Question{
			Question: midpoint,
		}

		// ELSE get midpoint and ask question
		answer, err := c.AskQuestionWebsocket(question)
		if err != nil {
			return s, err
		}

		switch answer.Answer {
		case "Good":
			err := d.GoodCommit(question.Question)
			if err != nil {
				return s, err
			}
			// log.Printf("Now %v commits after GOOD (%v)\n", d.GetOrder(), question.Question)
		case "Bad":
			err := d.BadCommit(question.Question)
			if err != nil {
				return s, err
			}
			// log.Printf("Now %v commits after BAD (%v)\n", d.GetOrder(), question.Question)
		}
	}
}
