package bisect

import (
	"log"

	"github.com/jamesjarvis/git-bisect/pkg/dag"
)

// NextMove actually contains the logic
func (c *ConnectionJSON) NextMove(d *dag.DAG) (Score, error) {
	var s Score
	for {

		// IF the length is 0, submit the last "badcommit"
		if d.GetOrder() == 0 {
			log.Printf("Submitting MOST RECENT (%v)\n", d.MostRecentBad)
			s, p, err := c.SubmitSolutionJSON(Solution{
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

			log.Printf("Problem: %v has %v vertexes (commits) and %v edges with new dag map\n", p.Name, d.GetOrder(), d.GetSize())

			err = d.GoodCommit(p.Good)
			if err != nil {
				return s, err
			}

			log.Printf("Problem: %v now has %v commits after GOOD (%v)\n", p.Name, d.GetOrder(), p.Good)

			err = d.BadCommit(p.Bad)
			if err != nil {
				return s, err
			}

			log.Printf("Problem: %v now has %v commits after BAD (%v)\n", p.Name, d.GetOrder(), p.Bad)

		}

		midpoint, err := d.GetMidPoint()
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
			log.Printf("Now %v commits after GOOD (%v)\n", d.GetOrder(), question.Question)
		case "Bad":
			err := d.BadCommit(question.Question)
			if err != nil {
				return s, err
			}
			log.Printf("Now %v commits after BAD (%v)\n", d.GetOrder(), question.Question)
		}
	}
}

// NextMoveWebsocket actually contains the logic
func (c *Connection) NextMoveWebsocket(d *dag.DAG) (Score, error) {
	var s Score
	for {

		// IF the length is 0, submit the last "badcommit"
		if d.GetOrder() == 0 {
			log.Printf("Submitting MOST RECENT (%v)\n", d.MostRecentBad)
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

			log.Printf("Problem: %v has %v vertexes (commits) and %v edges with new dag map\n", p.Name, d.GetOrder(), d.GetSize())

			err = d.GoodCommit(p.Good)
			if err != nil {
				return s, err
			}

			log.Printf("Problem: %v now has %v commits after GOOD (%v)\n", p.Name, d.GetOrder(), p.Good)

			err = d.BadCommit(p.Bad)
			if err != nil {
				return s, err
			}

			log.Printf("Problem: %v now has %v commits after BAD (%v)\n", p.Name, d.GetOrder(), p.Bad)

		}

		midpoint, err := d.GetMidPoint()
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
			log.Printf("Now %v commits after GOOD (%v)\n", d.GetOrder(), question.Question)
		case "Bad":
			err := d.BadCommit(question.Question)
			if err != nil {
				return s, err
			}
			log.Printf("Now %v commits after BAD (%v)\n", d.GetOrder(), question.Question)
		}
	}
}
