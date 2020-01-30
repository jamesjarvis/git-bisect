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

// StartGetParents initialises the "visited" map before starting the recursion.
func StartGetParents(m map[string]DAGEntry, commit string) map[string]DAGEntry {
	Visited = InitialiseVisited(m)
	currentParent := m[commit]

	return GetParents(m, currentParent)
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

	// Then repeat the process with the parents...
	for _, parent := range d.parents {
		currentParent := m[parent]
		tempAncestors[parent] = currentParent
		results := GetParents(m, currentParent)
		tempAncestors = AppendMaps(tempAncestors, results)
	}

	return tempAncestors
}

// StartGetLengthParents initialises the "visited" map before starting the recursion.
func StartGetLengthParents(m map[string]DAGEntry, commit string) int {
	Visited = InitialiseVisited(m)
	currentParent := m[commit]

	return GetLengthParents(m, currentParent)
}

// GetLengthParents is the recursive get parents method for retrieving ancestory
func GetLengthParents(m map[string]DAGEntry, d DAGEntry) int {

	lenAncestors := 0

	// If it doesnt exist any more, ignore.
	if !ExistInMap(m, d.commit) {
		return lenAncestors
	}

	// If it has been visited, ignore
	if VisitedStatus(Visited, d.commit) {
		return lenAncestors
	}

	// Then add itself to visited
	Visited[d.commit] = true

	// Then repeat the process with the parents...
	for _, parent := range d.parents {
		currentParent := m[parent]
		lenAncestors++
		results := GetLengthParents(m, currentParent)
		lenAncestors = lenAncestors + results
	}

	return lenAncestors
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
func GetFirstElementFromMap(m map[string]DAGEntry) string {
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
