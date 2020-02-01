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

var addr = flag.String("addr", "129.12.44.229:1234", "http service address")

// Authentication is the first message sent to the server
type Authentication struct {
	User string `json:"User"`
}

// ConnectWebsocket connects to the websocket server, and returns the problem
func ConnectWebsocket() (*websocket.Conn, error) {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "wss", Host: *addr, Path: "/"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	return c, nil
}

// GetProblemWebsocket simply returns the problem, given an authentication
func GetProblemWebsocket(s *websocket.Conn, a Authentication) (Problem, error) {
	var prob Problem

	jsona, err := json.Marshal(a)
	if err != nil {
		return prob, err
	}

	err = s.WriteMessage(websocket.TextMessage, jsona)
	if err != nil {
		return prob, err
	}

	_, message, err := s.ReadMessage()
	if err != nil {
		return prob, err
	}

	err = json.Unmarshal(message, &prob)
	if err != nil {
		return prob, err
	}

	return prob, nil
}

// AskQuestionWebsocket asks a question about this particular commit to the server, and updates the count
func AskQuestionWebsocket(s *websocket.Conn, q Question) (Answer, error) {
	var ans Answer

	jsonq, err := json.Marshal(q)
	if err != nil {
		return ans, err
	}

	err = s.WriteMessage(websocket.TextMessage, jsonq)
	if err != nil {
		return ans, err
	}

	_, message, err := s.ReadMessage()
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
func SubmitSolutionWebsocket(s *websocket.Conn, attempt Solution) (Score, Problem, error) {
	var scor Score
	var prob Problem

	jsonq, err := json.Marshal(attempt)
	if err != nil {
		return scor, prob, err
	}

	err = s.WriteMessage(websocket.TextMessage, jsonq)
	if err != nil {
		return scor, prob, err
	}

	_, message, err := s.ReadMessage()
	if err != nil {
		return scor, prob, err
	}

	// Assume it is the score
	err = json.Unmarshal(message, &scor)
	if err != nil {

		// Attempt to get the problem
		err = json.Unmarshal(message, &prob)
		if err != nil {
			return scor, prob, err
		}
		return scor, prob, nil
	}

	return scor, prob, nil
}
