package bisect

import (
	"fmt"
	"os"

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
func SaveResults(s *Score) error {
	f, err := os.Create("results.txt")
	if err != nil {
		return err
	}

	defer f.Close()

	for key, val := range s.Score {
		_, err := f.WriteString(fmt.Sprintf("Problem: %v, %v\n", key, val))
		if err != nil {
			return err
		}
	}

	return f.Sync()
}
