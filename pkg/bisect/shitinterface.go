package bisect

// GoodCommit changes the map with response to being a good commit
func GoodCommit(p map[string]DAGEntry, c string) map[string]DAGEntry {
	newsearchspace := GetParents(p, c)
	newsearchspace = append(newsearchspace, p[c])

	return GenerateMap(newsearchspace)
}

// BadCommit changes the map with response to being a bad commit
func BadCommit(p map[string]DAGEntry, c string) map[string]DAGEntry {
	newsearchspace := GetParents(p, c)
	newsearchspace = append(newsearchspace, p[c])

	return RemoveFromMap(p, newsearchspace)
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
