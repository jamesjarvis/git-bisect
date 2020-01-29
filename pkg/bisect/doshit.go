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

// GetParents retrieves all parents (and the parents parents) of a commit
func GetParents(p map[string]DAGEntry, c string) []DAGEntry {
	var parents []DAGEntry

	// //TODO: Remove duplicates

	// actualcommit, ok := p[c]
	// if !ok {
	// 	return nil
	// }
	// parents = append(parents, actualcommit)

	// for _, parent := range actualcommit.parents {
	// 	parents = append(parents, GetParents(p, parent)...)
	// }

	// fmt.Println(len(parents))

	return parents
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
