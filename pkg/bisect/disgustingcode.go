package bisect

import (
	"encoding/json"
)

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

// ProblemInstance is just a container for the problem
type ProblemInstance struct {
	Repo     Repo
	Instance Instance
}

// RepoContainer is just a way to get around Radu's json formatting
type RepoContainer struct {
	Repo Repo `json:"Repo"`
}

// Repo is the repo messager
type Repo struct {
	Name          string     `json:"name"`
	InstanceCount int        `json:"instance_count"`
	Dag           []DAGEntry `json:"dag"`
}

// InstanceContainer just contains the instance lol
type InstanceContainer struct {
	Instance Instance `json:"Instance"`
}

// Instance is the problem instance, with the good and bad commits
// It is assumed that the repo this instance refers to is simply the last repo mentioned before this message.
type Instance struct {
	Good string `json:"good"`
	Bad  string `json:"bad"`
}

// Question is the question json interface
type Question struct {
	Question string `json:"Question"`
}

// Answer is the answer json interface
type Answer struct {
	Answer string `json:"Answer"`
}

// Solution is the solution json interface
type Solution struct {
	Solution string `json:"Solution"`
}

// Score is the score json interface (should change for the websockets)
type Score struct {
	Score map[string]interface{} `json:"Score"`
}
