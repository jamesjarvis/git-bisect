package main

import (
	"flag"
	"log"
	"net/url"

	bisect "github.com/jamesjarvis/git-bisect/pkg/bisect"
	"github.com/jamesjarvis/git-bisect/pkg/dag"
)

func main() {
	var addr = flag.String("addr", "129.12.44.229:1234", "http service address")
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/"}

	log.Printf("Connecting to problem server (%v) 🤖\n", u.String())

	config := dag.ParamConfig{
		Limit:     20000,
		Divisions: 400,
	}

	conn, err := bisect.ConnectWebsocket(u)
	if err != nil {
		log.Print("Could not connect to websocket 🤖😢")
		log.Fatal(err)
	}
	defer conn.WS.Close()

	log.Println("Connected to websocket 🤖✅")

	auth := bisect.Authentication{
		User: []string{"jj333", "30e8e949"},
	}

	problem, err := conn.GetProblemWebsocket(auth)
	if err != nil {
		log.Print("You... Shall... not.... be authorised to connect to this server 😢")
		log.Fatal(err)
	}

	log.Printf("Retrieved problem %v, parsing...", problem.Repo.Name)

	newDag := bisect.DAGMaker(&problem.Repo)

	log.Printf("Problem: %v has %v vertexes (commits) and %v edges with new dag map\n", problem.Repo.Name, newDag.GetOrder(), newDag.GetSize())

	err = newDag.GoodCommit(problem.Instance.Good)
	if err != nil {
		log.Fatal(err)
	}

	// log.Printf("Problem: %v now has %v commits after GOOD (%v)\n", problem.Name, newDag.GetOrder(), problem.Good)

	err = newDag.BadCommit(problem.Instance.Bad)
	if err != nil {
		log.Fatal(err)
	}

	// log.Printf("Problem: %v now has %v commits after BAD (%v)\n", problem.Name, newDag.GetOrder(), problem.Bad)

	score, err := conn.NextMoveWebsocket(newDag, config, problem)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%v", score)
}
