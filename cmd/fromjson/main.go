package main

import (
	"log"
	"os"

	bisect "github.com/jamesjarvis/git-bisect/pkg/bisect"
)

func main() {
	log.Printf("Connecting to problem server")

	// Get directory of examples
	dirname := os.Args[1]

	conn, err := bisect.ConnectJSON(dirname)
	if err != nil {
		log.Fatal(err)
	}

	problem, err := conn.GetProblemJSON()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Retrieved problem %v, parsing...", problem.Name)

	newDag := bisect.DAGMaker(&problem)

	log.Printf("Problem: %v has %v vertexes (commits) and %v edges with new dag map\n", problem.Name, newDag.GetOrder(), newDag.GetSize())

	err = newDag.GoodCommit(problem.Good)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Problem: %v now has %v commits after GOOD (%v)\n", problem.Name, newDag.GetOrder(), problem.Good)

	err = newDag.BadCommit(problem.Bad)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Problem: %v now has %v commits after BAD (%v)\n", problem.Name, newDag.GetOrder(), problem.Bad)

	score, err := conn.NextMove(newDag)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("SCORES ON THE DOORS")

	totalScore := 0
	for name, scor := range score.Score {
		totalScore += scor
		log.Printf("%v had a score of %v", name, scor)
	}
	log.Printf("Total score: %v", totalScore)

}
