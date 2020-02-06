package main

import (
	"flag"
	"log"
	"net/url"

	bisect "github.com/jamesjarvis/git-bisect/pkg/bisect"
)

func main() {
	var addr = flag.String("addr", "129.12.44.229:1234", "http service address")
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/"}

	log.Printf("Connecting to problem server (%v) 🤖\n", u.String())

	conn, err := bisect.ConnectWebsocket(u)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.WS.Close()

	log.Println("Connected to websocket 🤖✅")

	auth := bisect.Authentication{
		User: "jj333",
	}

	problem, err := conn.GetProblemWebsocket(auth)
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

	// log.Printf("Problem: %v now has %v commits after GOOD (%v)\n", problem.Name, newDag.GetOrder(), problem.Good)

	err = newDag.BadCommit(problem.Bad)
	if err != nil {
		log.Fatal(err)
	}

	// log.Printf("Problem: %v now has %v commits after BAD (%v)\n", problem.Name, newDag.GetOrder(), problem.Bad)

	score, err := conn.NextMoveWebsocket(newDag)
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
