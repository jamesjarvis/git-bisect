package bisect

// DagMapMaker takes the problem struct and returns the Dag Map
func DagMapMaker(p *Problem) map[string]DAGEntry {
	var m map[string]DAGEntry
	m = make(map[string]DAGEntry)

	for _, element := range p.Dag {
		m[element.commit] = element
	}

	return m
}

var Visited map[string]bool

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

// StartGetParents initialises the "visited" map before starting the recursion.
func StartGetParents(m map[string]DAGEntry, commit string) map[string]DAGEntry {
	Visited = InitialiseVisited(m)
	currentParent := m[commit]

	return GetParents(m, currentParent)
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

// GetParents is the recursive get parents method for retrieving ancestory
func GetParents(m map[string]DAGEntry, d DAGEntry) map[string]DAGEntry {

	tempAncestors := make(map[string]DAGEntry)

	// If it doesnt exist any more, ignore.
	if !ExistInMap(m, d.commit) {
		return tempAncestors
	}

	// If it has been visited, ignore
	if VisitedStatus(Visited, d.commit) {
		return tempAncestors
	}

	// Then add itself to visited
	Visited[d.commit] = true

	// currentCommit := m[commit]
	// Ancestors[commit] = currentCommit

	// Then repeat the process with the children...
	for _, parent := range d.parents {
		currentParent := m[parent]
		tempAncestors[parent] = currentParent
		results := GetParents(m, currentParent)
		tempAncestors = AppendMaps(tempAncestors, results)
	}

	// // Remove parents that no longer exist
	// var tempParents []string
	// for _, parent := range currentCommit.parents {
	// 	if ExistInMap(m, parent) {
	// 		tempParents = append(tempParents, parent)
	// 	}
	// }
	// currentCommit.parents = tempParents

	return tempAncestors
}

// // GetParents retrieves all parents (and the parents parents) of a commit
// func GetParents(p map[string]DAGEntry, c string) []DAGEntry {
// 	var parents []DAGEntry

// 	// //TODO: Remove duplicates

// 	actualcommit, ok := p[c]
// 	if !ok {
// 		return nil
// 	}
// 	// parents = append(parents, actualcommit)

// 	for _, parent := range actualcommit.parents {
// 		parents = append(parents, GetParents(p, parent)...)
// 	}

// 	for i := range parents {
// 		if parents[i].commit == c {
// 			return nil
// 		}
// 	}

// 	fmt.Println(len(parents))

// 	return append(parents, actualcommit)
// }

// func GetAllParents(p map[string]DAGEntry, c string) map[string]DAGEntry {
// 	var parents map[string]DAGEntry

// 	// First, get the initial commit, and generate an empty map, and a mutex.
// 	// Then, send each parent to a pool of workers and
// }

// // Create a datastore with the hashamp
// // Use Channels to send requests to this hashmap.

// // Worker would ask whether to include this particular commit (does it exist in the total map, and does it not already exist in the new parents?)
// // Then the datastore would return with the DAGEntry if it has been added, or nil if it was not added
// // Then the worker would either die if nil, or spawn new workers with the parents from the DAGEntry
// // Then finally, to kill the datastore and signal that the parents have all been found, the "done" channel should be empty??

// func parentWorker(commit string, wg *sync.WaitGroup) {

// }

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
func GetFirstElementFromMap(m map[string]DAGEntry) string {
	var temp string
	for key := range m {
		return key
	}
	return temp
}
