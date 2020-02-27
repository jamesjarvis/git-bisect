package bisect

import (
	"log"
	"math"

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

// NextMove actually contains the logic
func (c *ConnectionJSON) NextMove(d *dag.DAG, pc dag.ParamConfig, idealScores map[string]float64) (Score, error) {
	var s Score
	for {

		// IF the length is 0, submit the last "badcommit"
		if d.GetOrder() == 0 {
			// log.Printf("Submitting MOST RECENT (%v)\n", d.MostRecentBad)
			s, p, err := c.SubmitSolutionJSON(Solution{
				Solution: d.MostRecentBad,
			})
			if err != nil {
				return Score{}, err
			}
			if p.Name == "" { // Actually complete
				s.IdealScores = CopyMapFloat(idealScores)
				// log.Printf("Size of the ideal scores: %v, %v", idealScores, s.IdealScores)
				return s, err
			}

			// Else, restart with the new problem

			d = DAGMaker(&p)

			log.Printf("Problem: %v has %v vertexes (commits) and %v edges\n", p.Name, d.GetOrder(), d.GetSize())

			err = d.GoodCommit(p.Good)
			if err != nil {
				return s, err
			}

			// log.Printf("Problem: %v now has %v commits after GOOD (%v)\n", p.Name, d.GetOrder(), p.Good)

			err = d.BadCommit(p.Bad)
			if err != nil {
				return s, err
			}

			idealScores[p.Name] = math.Log2(float64(d.GetOrder()))
			// log.Printf("Stored new ideal score: %v : %v", p.Name, idealScores[p.Name])

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
		answer, err := c.AskQuestionJSON(question)
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

// NextMoveWebsocket actually contains the logic
func (c *Connection) NextMoveWebsocket(d *dag.DAG, pc dag.ParamConfig) (Score, error) {
	var s Score
	for {

		// IF the length is 0, submit the last "badcommit"
		if d.GetOrder() == 0 {
			// log.Printf("Submitting MOST RECENT (%v)\n", d.MostRecentBad)
			s, p, err := c.SubmitSolutionWebsocket(Solution{
				Solution: d.MostRecentBad,
			})
			if err != nil {
				return Score{}, err
			}
			if p.Name == "" {
				return s, err
			}

			// Else, restart with the new problem

			d = DAGMaker(&p)

			log.Printf("Problem: %v has %v vertexes (commits) and %v edges\n", p.Name, d.GetOrder(), d.GetSize())

			err = d.GoodCommit(p.Good)
			if err != nil {
				return s, err
			}

			// log.Printf("Problem: %v now has %v commits after GOOD (%v)\n", p.Name, d.GetOrder(), p.Good)

			err = d.BadCommit(p.Bad)
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
