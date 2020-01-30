package bisect

// GoodCommit changes the map with response to being a good commit
// New search space is it and it's ancestors
func GoodCommit(p map[string]DAGEntry, c string) map[string]DAGEntry {
	newsearchspace := StartGetParents(p, c)
	newsearchspace[c] = p[c]

	return RemoveMapFromMap(p, newsearchspace)
}

// BadCommit changes the map with response to being a bad commit
// New search space is the old search space - it and it's ancestors
func BadCommit(p map[string]DAGEntry, c string) map[string]DAGEntry {
	newsearchspace := StartGetParents(p, c)
	newsearchspace[c] = p[c]

	return newsearchspace
}

// MidPoint gets a midpoint from the map
func MidPoint(p map[string]DAGEntry) DAGEntry {
	length := len(p)
	count := length / 2

	for _, value := range p {
		count--
		if count == 0 {
			return value
		}
	}
	return DAGEntry{}
}
