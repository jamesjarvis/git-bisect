package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	bisect "github.com/jamesjarvis/git-bisect/pkg/bisect"
)

func main() {
	jsonFile, err := os.Open("/Users/jarjames/git/git-bisect/tests/test_react0.json")
	if err != nil {
		panic(err)
	}
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		panic(err)
	}

	var file bisect.Root

	json.Unmarshal([]byte(byteValue), &file)
	if err != nil {
		panic(err)
	}

	dag := bisect.DagMapMaker(&file.Problem)

	fmt.Printf("Problem: %v has %v commits\n", file.Problem.Name, len(dag))

	dag = bisect.GoodCommit(dag, file.Problem.Good)

	fmt.Printf("Problem: %v now has %v commits after GOOD (%v)\n", file.Problem.Name, len(dag), file.Problem.Good)

	dag = bisect.BadCommit(dag, file.Problem.Bad)

	fmt.Printf("Problem: %v now has %v commits after BAD (%v)\n", file.Problem.Name, len(dag), file.Problem.Bad)

	// fmt.Printf("First commit from remaining: %v", dag)
}
