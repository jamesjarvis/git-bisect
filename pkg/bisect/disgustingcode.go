package bisect

import "encoding/json"

// DAGEntry is the actual DAG part
type DAGEntry struct {
	commit  string
	parents []string
}

func (d *DAGEntry) UnmarshalJSON(data []byte) error {

	var v []interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	d.commit, _ = v[0].(string)

	for _, parent := range v[1].([]interface{}) {
		d.parents = append(d.parents, parent.(string))
	}

	return nil
}

// Problem is the problem section of the json
type Problem struct {
	Name string     `json:"name"`
	Good string     `json:"good"`
	Bad  string     `json:"bad"`
	Dag  []DAGEntry `json:"dag"`
}

// Solution is the solution section of the json
type Solution struct {
	Bug    string   `json:"bug"`
	AllBad []string `json:"all_bad"`
}

// Root is all of the json
type Root struct {
	Problem  Problem
	Solution Solution
}

func (d *Root) UnmarshalJSON(data []byte) error {
	var v []interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	pmap := v[0]
	smap := v[1]

	pbytes, _ := json.Marshal(pmap)
	sbytes, _ := json.Marshal(smap)

	err := json.Unmarshal(pbytes, &d.Problem)

	if err != nil {
		return err
	}

	err = json.Unmarshal(sbytes, &d.Solution)

	if err != nil {
		return err
	}

	return err
}
