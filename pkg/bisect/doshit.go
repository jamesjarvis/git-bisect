package bisect

import (
	"log"

	"github.com/jamesjarvis/git-bisect/pkg/dag"
)

// DagMapMaker takes the problem struct and returns the Dag Map
func DagMapMaker(p *Problem) map[string]DAGEntry {
	var m map[string]DAGEntry
	m = make(map[string]DAGEntry)

	for _, element := range p.Dag {
		m[element.commit] = element
	}

	return m
}

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

// InitialiseVisited returns a map with the same keys as the input map, but false as all of its values
func InitialiseVisited(m map[string]DAGEntry) map[string]bool {
	tempVisited := make(map[string]bool)
	for key := range m {
		tempVisited[key] = false
	}
	return tempVisited
}

// VisitedStatus returns true or false, depending on if it has been visited. Defaults to true if it does not exist.
func VisitedStatus(v map[string]bool, commit string) bool {
	value, ok := v[commit]
	if !ok {
		return true
	}
	return value
}

// AppendMaps literally just concatenates two maps together
func AppendMaps(a map[string]DAGEntry, b map[string]DAGEntry) map[string]DAGEntry {
	for k, v := range b {
		a[k] = v
	}
	return a
}

// RemoveMapFromMap removes map b from map a
func RemoveMapFromMap(a map[string]DAGEntry, b map[string]DAGEntry) map[string]DAGEntry {
	for k := range b {
		delete(a, k)
	}
	return a
}

// ExistInMap checks if a DAGEntry exists in a map
func ExistInMap(m map[string]DAGEntry, c string) bool {
	_, ok := m[c]
	return ok
}

// RemoveFromMap literally just removes a list from the map
func RemoveFromMap(m map[string]DAGEntry, e []DAGEntry) map[string]DAGEntry {
	for _, element := range e {
		delete(m, element.commit)
	}
	return m
}

// GenerateMap takes a list of DAGEntry's and returns a map of them (thereby removing duplicates)
func GenerateMap(entries []DAGEntry) map[string]DAGEntry {
	var mappeh map[string]DAGEntry
	for _, element := range entries {
		mappeh[element.commit] = element
	}

	return mappeh
}

// GetFirstElementFromMap gets the first element from the map
func GetFirstElementFromMap(m map[string]bool) string {
	var temp string
	for key := range m {
		return key
	}
	return temp
}

// EstimateMaxAncestry returns an estimate of the max ancestry length
func EstimateMaxAncestry(m map[string]DAGEntry) int {
	return len(m)
}
