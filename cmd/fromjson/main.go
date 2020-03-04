// package main

// import (
// 	"log"
// 	"math"
// 	"os"

// 	bisect "github.com/jamesjarvis/git-bisect/pkg/bisect"
// 	"github.com/jamesjarvis/git-bisect/pkg/dag"
// )

// func main() {
// 	log.Printf("Connecting to problem server")

// 	config := dag.ParamConfig{
// 		Limit:     20000,
// 		Divisions: 200,
// 	}

// 	// IdealScores is just a global variable to store the "ideal" (log2(n)) scores for each problem
// 	var IdealScores map[string]float64

// 	// Get directory of examples
// 	dirname := os.Args[1]

// 	conn, err := bisect.ConnectJSON(dirname)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	problem, err := conn.GetProblemJSON()
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	log.Printf("Retrieved problem %v, parsing...", problem.Name)

// 	newDag := bisect.DAGMaker(&problem)

// 	log.Printf("Problem: %v has %v vertexes (commits) and %v edges\n", problem.Name, newDag.GetOrder(), newDag.GetSize())

// 	err = newDag.GoodCommit(problem.Good)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// log.Printf("Problem: %v now has %v commits after GOOD (%v)\n", problem.Name, newDag.GetOrder(), problem.Good)

// 	err = newDag.BadCommit(problem.Bad)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	IdealScores = make(map[string]float64)
// 	IdealScores[problem.Name] = math.Log2(float64(newDag.GetOrder()))
// 	// log.Printf("ideal score for %v : %v", problem.Name, math.Log2(float64(newDag.GetOrder())))
// 	// log.Printf("stored ideal score for %v: %v", problem.Name, IdealScores[problem.Name])

// 	// log.Printf("Problem: %v now has %v commits after BAD (%v)\n", problem.Name, newDag.GetOrder(), problem.Bad)

// 	score, err := conn.NextMove(newDag, config, IdealScores)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	IdealScores = bisect.CopyMapFloat(score.IdealScores)

// 	log.Print("SCORES ON THE DOORS")
// 	log.Printf("CONFIG: %v LIMIT, %v DIVISIONS", config.Limit, config.Divisions)

// 	totalScore := 0
// 	var totalIdealScore float64
// 	for name, scor := range score.Score {
// 		totalScore += scor
// 		log.Printf("%v has score: %v", name, scor)
// 	}

// 	for _, idealScor := range score.IdealScores {
// 		totalIdealScore += idealScor
// 	}
// 	log.Printf("Total score: %v", totalScore)
// 	log.Printf("Total ideal score: %v", totalIdealScore)

// }
