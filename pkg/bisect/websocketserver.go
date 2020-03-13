package bisect

import (
	"encoding/json"
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

// Authentication is the first message sent to the server
type Authentication struct {
	User []string `json:"User"`
}

// Connection is the websocket connection
type Connection struct {
	WS      *websocket.Conn
	Timeout time.Duration
}

// ConnectWebsocket connects to the websocket server, and returns the problem
func ConnectWebsocket(u url.URL, t time.Duration) (*Connection, error) {
	flag.Parse()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}

	return &Connection{c, t}, nil
}

// GetProblemWebsocket simply returns the problem, given an authentication
func (c *Connection) GetProblemWebsocket(a Authentication) (ProblemInstance, error) {
	var prob ProblemInstance

	var repo RepoContainer
	var inst InstanceContainer

	jsona, err := json.Marshal(a)
	if err != nil {
		return prob, err
	}

	// Send off the authentication
	err = c.WS.WriteMessage(websocket.TextMessage, jsona)
	if err != nil {
		return prob, err
	}

	// Retrieve the Initial Repo
	_, message, err := c.WS.ReadMessage()
	if err != nil {
		return prob, err
	}

	// Parse the Repo message
	err = json.Unmarshal(message, &repo)
	if err != nil {
		return prob, err
	}

	// Retrieve the instance message
	_, instanceMessage, err := c.WS.ReadMessage()
	if err != nil {
		return prob, err
	}

	// Parse the instance message
	err = json.Unmarshal(instanceMessage, &inst)
	if err != nil {
		return prob, err
	}

	// If there are no issues, then return the new problem instance
	prob.Repo = repo.Repo
	prob.Instance = inst.Instance

	// log.Print(string(message))

	return prob, nil
}

// AskQuestionWebsocket asks a question about this particular commit to the server
func (c *Connection) AskQuestionWebsocket(q Question) (Answer, error) {
	var ans Answer

	jsonq, err := json.Marshal(q)
	if err != nil {
		return ans, err
	}

	c.WS.SetWriteDeadline(time.Now().Add(c.Timeout))
	c.WS.SetReadDeadline(time.Now().Add(c.Timeout))

	// time.Sleep(time.Second / 4)

	err = c.WS.WriteMessage(websocket.TextMessage, jsonq)
	if err != nil {
		log.Printf("Error writing question")
		return ans, err
	}

	// time.Sleep(time.Second / 100)

	_, message, err := c.WS.ReadMessage()
	if err != nil {
		log.Printf("Error retrieving question answer")
		return ans, err
	}

	err = json.Unmarshal(message, &ans)
	if err != nil {
		return ans, err
	}

	return ans, nil
}

// SubmitSolutionWebsocket is the "endpoint" where you can submit a solution
// It can either return a score, or an instance, or a new repo, which is then followed by an instance.
// Really intuitive and simple?
func (c *Connection) SubmitSolutionWebsocket(attempt Solution, currentProb ProblemInstance) (Score, ProblemInstance, error) {
	var scor Score
	var prob ProblemInstance
	var repo RepoContainer
	var inst InstanceContainer

	c.WS.SetWriteDeadline(time.Now().Add(c.Timeout))
	c.WS.SetReadDeadline(time.Now().Add(c.Timeout))

	// Marshal the submission
	jsonq, err := json.Marshal(attempt)
	if err != nil {
		return scor, prob, err
	}

	// time.Sleep(time.Second / 4)

	// Submit
	err = c.WS.WriteMessage(websocket.TextMessage, jsonq)
	if err != nil {
		log.Printf("Error sending solution")
		return scor, prob, err
	}

	// time.Sleep(time.Second / 4)

	// Retrieve the response
	_, message, err := c.WS.ReadMessage()
	if err != nil {
		log.Printf("Error retrieving solution answer")
		return scor, prob, err
	}

	// Assume it is the score
	err = json.Unmarshal(message, &scor)
	if err != nil || len(scor.Score) == 0 {

		// Attempt to get the instance
		err = json.Unmarshal(message, &inst)
		if err != nil || inst.Instance.Good == "" {

			// Attempt to retrieve repo
			err = json.Unmarshal(message, &repo)
			if err != nil {
				return scor, prob, err
			}

			// time.Sleep(time.Second / 4)

			// Now get the instance
			_, instmessage, err := c.WS.ReadMessage()
			if err != nil {
				log.Printf("Error retrieving new instance")
				return scor, prob, err
			}

			// Attempt to retrieve instance message
			err = json.Unmarshal(instmessage, &inst)
			if err != nil {
				return scor, prob, err
			}

			// Now build the new problem
			prob.Repo = repo.Repo
			prob.Instance = inst.Instance

			log.Printf("Retrieved new PROBLEM: %v", prob.Repo.Name)

			// Return the new problem
			return scor, prob, nil
		}
		prob.Repo = currentProb.Repo
		prob.Instance = inst.Instance

		log.Printf("Retrieved new INSTANCE for: %v", prob.Repo.Name)

		// Return the new instance of the same problem
		return scor, prob, nil
	}

	// Return the final score
	return scor, prob, nil
}
