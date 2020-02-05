package bisect

import (
	"encoding/json"
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
)

// Authentication is the first message sent to the server
type Authentication struct {
	User string `json:"User"`
}

// Connection is the websocket connection
type Connection struct {
	WS *websocket.Conn
}

// ConnectWebsocket connects to the websocket server, and returns the problem
func ConnectWebsocket(u url.URL) (*Connection, error) {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}

	return &Connection{c}, nil
}

// GetProblemWebsocket simply returns the problem, given an authentication
func (c *Connection) GetProblemWebsocket(a Authentication) (Problem, error) {
	var prob Problemcontainer

	jsona, err := json.Marshal(a)
	if err != nil {
		return prob.Problem, err
	}

	err = c.WS.WriteMessage(websocket.TextMessage, jsona)
	if err != nil {
		return prob.Problem, err
	}

	_, message, err := c.WS.ReadMessage()
	if err != nil {
		return prob.Problem, err
	}

	err = json.Unmarshal(message, &prob)
	if err != nil {
		return prob.Problem, err
	}

	// log.Print(string(message))

	return prob.Problem, nil
}

// AskQuestionWebsocket asks a question about this particular commit to the server, and updates the count
func (c *Connection) AskQuestionWebsocket(q Question) (Answer, error) {
	var ans Answer

	jsonq, err := json.Marshal(q)
	if err != nil {
		return ans, err
	}

	err = c.WS.WriteMessage(websocket.TextMessage, jsonq)
	if err != nil {
		return ans, err
	}

	_, message, err := c.WS.ReadMessage()
	if err != nil {
		return ans, err
	}

	err = json.Unmarshal(message, &ans)
	if err != nil {
		return ans, err
	}

	return ans, nil
}

// SubmitSolutionWebsocket is the "endpoint" where you can submit a solution
// It can either return a score, or
func (c *Connection) SubmitSolutionWebsocket(attempt Solution) (Score, Problem, error) {
	var scor Score
	var prob Problemcontainer

	jsonq, err := json.Marshal(attempt)
	if err != nil {
		return scor, prob.Problem, err
	}

	err = c.WS.WriteMessage(websocket.TextMessage, jsonq)
	if err != nil {
		return scor, prob.Problem, err
	}

	_, message, err := c.WS.ReadMessage()
	if err != nil {
		return scor, prob.Problem, err
	}

	// Assume it is the score
	err = json.Unmarshal(message, &scor)
	if err != nil || len(scor.Score) == 0 {

		// Attempt to get the problem
		err = json.Unmarshal(message, &prob)
		if err != nil {
			return scor, prob.Problem, err
		}
		// Return the new problem
		return scor, prob.Problem, nil
	}

	// Return the final score
	return scor, prob.Problem, nil
}
