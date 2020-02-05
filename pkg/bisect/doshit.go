package bisect

import (
	"log"

	"github.com/jamesjarvis/git-bisect/pkg/dag"
)

// DAGMaker takes the problem struct and returns the Dag
func DAGMaker(p *Problem) *dag.DAG {
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
				log.Printf("Error adding edge of (%v) -> (%v)", currentParentVertex, currentVertex)
				// log.Fatal(err)
			}
		}
	}

	return d
}
