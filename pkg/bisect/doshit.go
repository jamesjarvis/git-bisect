package bisect

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/jamesjarvis/git-bisect/pkg/dag"
)

// DAGMaker takes the problem struct and returns the Dag
func DAGMaker(p *Repo) *dag.DAG {
	// initialize a new graph
	d := dag.NewDAG()

	var currentVertex string
	var currentParentVertex string
	var err error

	// Add the vertex and edge's
	for _, current := range p.Dag {
		currentVertex = current.commit

		for _, parent := range current.parents {
			currentParentVertex = parent

			err = d.AddEdge(currentParentVertex, currentVertex)
			if err != nil {
				// log.Printf("Error adding edge of (%v) -> (%v)", currentParentVertex, currentVertex)
			}
		}
	}

	return d
}

// SaveResults saves the scores to a file
func SaveResults(s *Score, start time.Time) error {
	end := time.Now()

	f, err := os.Create("results.txt")
	if err != nil {
		return err
	}

	defer f.Close()

	totalscore := 0
	correctsolutions := 0
	wrongsolutions := 0
	gaveup := 0

	for key, val := range s.Score {
		corr, ok := val.(map[string]interface{})
		if !ok {
			wrong, ok := val.(string)
			if !ok {
				log.Printf("Failed failure analysis with %v, %v", key, val)
			}
			if wrong == "GaveUp" {
				// If you give up, it counts as a score of 30 I believe?
				gaveup++
				_, err := f.WriteString(fmt.Sprintf("😤 %v", key))
				if err != nil {
					return err
				}
			} else if wrong == "Wrong" {
				wrongsolutions++
				_, err := f.WriteString(fmt.Sprintf("❌ %v", key))
				if err != nil {
					return err
				}
			}
		} else {
			finalint, ok := corr["Correct"].(float64)
			if !ok {
				finalstring, ok := corr["Correct"].(string)
				if !ok {
					log.Printf("Failed score analysis with %v, %v", key, val)
				} else {
					intversion, err := strconv.ParseInt(finalstring, 10, 32)
					if err != nil {
						return err
					}
					totalscore += int(intversion)
					correctsolutions++
					_, err = f.WriteString(fmt.Sprintf("✅ %v : %v\n", key, intversion))
					if err != nil {
						return err
					}
				}
			} else {
				totalscore += int(finalint)
				correctsolutions++
				_, err = f.WriteString(fmt.Sprintf("✅ %v : %v\n", key, finalint))
				if err != nil {
					return err
				}
			}

		}
	}

	f.WriteString(fmt.Sprintf("Total problems: %v\n", len(s.Score)))
	f.WriteString(fmt.Sprintf("Correct solutions: %v, Incorrect solutions: %v, GaveUp: %v\n", correctsolutions, wrongsolutions, gaveup))
	f.WriteString(fmt.Sprintf("Total questions asked: %v\n", totalscore))
	f.WriteString(fmt.Sprintf("Average questions per correct problem: %v\n", totalscore/correctsolutions))
	f.WriteString(fmt.Sprintf("Started at: %v\n", start.String()))
	f.WriteString(fmt.Sprintf("Completed at: %v\n", end.String()))
	f.WriteString(fmt.Sprintf("Time taken: %v\n", time.Since(start).String()))

	return f.Sync()
}
