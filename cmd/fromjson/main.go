package main

import (
	"log"

	bisect "github.com/jamesjarvis/git-bisect/pkg/bisect"
)

func main() {
	problem := bisect.Connect()

	dag := bisect.DagMapMaker(&problem)

	log.Printf("Problem: %v has %v commits\n", problem.Name, len(dag))

	dag = bisect.GoodCommit(dag, problem.Good)

	// fmt.Printf("Problem: %v now has %v commits after GOOD (%v)\n", problem.Name, len(dag), problem.Good)

	dag = bisect.BadCommit(dag, problem.Bad)

	// fmt.Printf("Problem: %v now has %v commits after BAD (%v)\n", problem.Name, len(dag), problem.Bad)

	score := bisect.NextMove(dag)

	log.Printf("Score for %v: %v\n", problem.Name, score.Score)
}
