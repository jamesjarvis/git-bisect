package main

import (
	"log"

	bisect "github.com/jamesjarvis/git-bisect/pkg/bisect"
)

func main() {
	log.Printf("Connecting to problem server")

	problem := bisect.Connect()

	log.Printf("Retrieved problem %v, parsing...", problem.Name)

	// dag := bisect.DagMapMaker(&problem)

	// log.Printf("Problem: %v has %v commits with original map\n", problem.Name, len(dag))

	newDag := bisect.DAGMaker(&problem)

	log.Printf("Problem: %v has %v vertexes (commits) and %v edges with new dag map\n", problem.Name, newDag.GetOrder(), newDag.GetSize())

	bisect.GoodCommitNew(newDag, problem.Good)

	// dag = bisect.GoodCommit(dag, problem.Good)

	// // fmt.Printf("Problem: %v now has %v commits after GOOD (%v)\n", problem.Name, len(dag), problem.Good)

	// dag = bisect.BadCommit(dag, problem.Bad)

	// // fmt.Printf("Problem: %v now has %v commits after BAD (%v)\n", problem.Name, len(dag), problem.Bad)

	// score := bisect.NextMove(dag)

	// log.Printf("Score for %v: %v\n", problem.Name, score.Score)
}
