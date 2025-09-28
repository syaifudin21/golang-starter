package websocket

import (
	"encoding/json"
	"exam/internal/dtos"
	"exam/internal/model"
	"exam/internal/service"
	"fmt"
	"log"
	"time"
)

const (
	StateWaiting    = "waiting"
	StateInProgress = "in_progress"
	StateFinished   = "finished"
)

// InboundMessage is a message from a client to the room.
type InboundMessage struct {
	Client  *Client
	Type    string
	Payload json.RawMessage
}

// Room maintains the set of active clients and manages the game state.
type Room struct {
	QuizID                   string
	quizSessionID            uint // New field to store the ID of the current quiz session
	State                    string
	Clients                  map[*Client]bool
	clientsByUserID          map[uint]*Client // New map to track clients by UserID
	Register                 chan *Client
	Unregister               chan *Client
	Inbound                  chan *InboundMessage
	quizService              *service.QuizService
	quiz                     *model.Quiz
	scores                   map[uint]*dtos.PlayerScore
	currentQuestionIndex     int
	answeredPlayers          map[uint]bool // Players who have answered the current question
	isQuestionAnsweredCorrectly bool          // Flag to track if the current question has been answered correctly by anyone
}

func NewRoom(quizID string, quizService *service.QuizService) *Room {
	return &Room{
		QuizID:                   quizID,
		quizSessionID:            0, // Initialize with 0, will be set by startGame
		State:                    StateWaiting,
		Clients:                  make(map[*Client]bool),
		clientsByUserID:          make(map[uint]*Client), // Initialize the new map
		Register:                 make(chan *Client),
		Unregister:               make(chan *Client),
		Inbound:                  make(chan *InboundMessage),
		quizService:              quizService,
		scores:                   make(map[uint]*dtos.PlayerScore),
		currentQuestionIndex:     -1,
		answeredPlayers:          make(map[uint]bool),
		isQuestionAnsweredCorrectly: false,
	}
}

func (r *Room) Run() {
	log.Printf("Room %s is running", r.QuizID)
	for {
		select {
		case client := <-r.Register:
			r.handleClientRegister(client)
		case client := <-r.Unregister:
			r.handleClientUnregister(client)
		case msg := <-r.Inbound:
			r.handleInboundMessage(msg)
		}
	}
}

func (r *Room) handleInboundMessage(msg *InboundMessage) {
	switch msg.Type {
	case "start_game":
		if msg.Client != nil { // Client-triggered start (deprecated)
			log.Printf("Warning: start_game message received from client %d. Use API to start quiz.", msg.Client.UserID)
			r.startGame(0) // Pass 0 as session ID for client-triggered start
		} else { // API-triggered start
			var payload struct { SessionID uint `json:"session_id"` }
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				log.Printf("Error unmarshalling start_game payload: %v", err)
				return
			}
			r.startGame(payload.SessionID)
		}
	case "submit_answer":
		var payload dtos.SubmitAnswerPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			log.Printf("Error unmarshalling submit_answer payload: %v", err)
			return
		}
		r.handleSubmitAnswer(msg.Client, payload)
	}
}

func (r *Room) startGame(sessionID uint) {
	if r.State != StateWaiting {
		return
	}

	r.quizSessionID = sessionID // Store the session ID

	quiz, err := r.quizService.GetQuizWithQuestions(r.QuizID)
	if err != nil {
		log.Printf("Error loading quiz: %v", err)
		return
	}
	r.quiz = quiz
	r.State = StateInProgress

	log.Printf("Starting game for quiz: %s (Session ID: %d)", r.quiz.Title, r.quizSessionID)
	r.broadcastMessage("game_starting", nil, nil)

	time.AfterFunc(3*time.Second, r.sendNextQuestion)
}

func (r *Room) sendNextQuestion() {
	r.currentQuestionIndex++
	log.Printf("sendNextQuestion called. Current question index: %d, Total questions: %d for quiz %s", r.currentQuestionIndex, len(r.quiz.Questions), r.QuizID)

	if r.currentQuestionIndex >= len(r.quiz.Questions) {
		r.endGame()
		return
	}

	// Reset trackers for the new question
	r.answeredPlayers = make(map[uint]bool)
	r.isQuestionAnsweredCorrectly = false

	currentQuestion := r.quiz.Questions[r.currentQuestionIndex]
	questionDTO := dtos.QuizQuestionDTO{
		ID:      currentQuestion.ID,
		Content: json.RawMessage(currentQuestion.Content),
		Options: json.RawMessage(currentQuestion.Options),
	}

	log.Printf("Sending question %d", r.currentQuestionIndex+1)
	r.broadcastMessage("next_question", questionDTO, nil)
}

func (r *Room) handleSubmitAnswer(client *Client, payload dtos.SubmitAnswerPayload) {
	if r.State != StateInProgress || r.answeredPlayers[client.UserID] {
		return
	}

	r.answeredPlayers[client.UserID] = true
	log.Printf("User %d submitted answer for question %d. Answered players: %d, Total clients: %d", client.UserID, payload.QuestionID, len(r.answeredPlayers), len(r.Clients))

	currentQuestion := r.quiz.Questions[r.currentQuestionIndex]

	isCorrect := payload.Answer == currentQuestion.CorrectAnswer
	wasFirstCorrectAnswer := false

	if isCorrect {
		// Check if this is the first correct answer for this question
		if !r.isQuestionAnsweredCorrectly {
			r.isQuestionAnsweredCorrectly = true
			wasFirstCorrectAnswer = true
			r.scores[client.UserID].Score += 10 // Award points only to the first
		}
	}

	// Record the answer
	if r.quizSessionID != 0 {
		answerRecord := &model.QuizAnswer{
			QuizSessionID: r.quizSessionID,
			QuestionID:    currentQuestion.ID,
			UserID:        client.UserID,
			Answer:        payload.Answer,
			IsCorrect:     isCorrect,
			SubmittedAt:   time.Now(),
		}
		if err := r.quizService.RecordQuizAnswer(answerRecord); err != nil {
			log.Printf("Error recording quiz answer for session %d: %v", r.quizSessionID, err)
		}
	}

	resultPayload := dtos.AnswerResultPayload{
		QuestionID:    currentQuestion.ID,
		IsCorrect:     isCorrect,
		PlayerID:      client.UserID,
		PlayerName:    r.scores[client.UserID].UserName,
		IsFirstAnswer: wasFirstCorrectAnswer,
	}
	r.broadcastMessage("answer_result", resultPayload, nil)

	// Send score update only if a score changed
	if wasFirstCorrectAnswer {
		var scoreList []dtos.PlayerScore
		for _, s := range r.scores {
			scoreList = append(scoreList, *s)
		}
		r.broadcastMessage("score_update", dtos.ScoreUpdatePayload{Scores: scoreList}, nil)
	}

	if len(r.answeredPlayers) == len(r.Clients) {
		log.Println("All players have answered. Moving to next question.")
		time.AfterFunc(2*time.Second, r.sendNextQuestion)
	}
}

func (r *Room) endGame() {
	r.State = StateFinished
	log.Printf("Game %s finished. Current question index: %d, Total questions: %d", r.QuizID, r.currentQuestionIndex, len(r.quiz.Questions))

	var winner dtos.PlayerScore
	// Initialize winner with a score that ensures any actual player score will be higher
	winner.Score = -1 // Assuming scores are non-negative

	for _, score := range r.scores {
		if score.Score > winner.Score {
			winner = *score
		}
	}

	var scoreList []dtos.PlayerScore
	for _, s := range r.scores {
		scoreList = append(scoreList, *s)
	}

	gameOverPayload := dtos.GameOverPayload{
		Winner: winner,
		Scores: scoreList,
	}
	r.broadcastMessage("game_over", gameOverPayload, nil)
	log.Printf("Game over message broadcast for quiz %s. Winner: %s (Score: %d)", r.QuizID, winner.UserName, winner.Score)

	// Record final scores and end time in the quiz session
	if r.quizSessionID != 0 {
		if err := r.quizService.EndQuizSession(r.quizSessionID, scoreList); err != nil {
			log.Printf("Error ending quiz session %d: %v", r.quizSessionID, err)
		}
	}
}

// --- Helper methods ---

func (r *Room) handleClientRegister(client *Client) {
	// If a client with this UserID is already connected, disconnect the old one
	if oldClient, ok := r.clientsByUserID[client.UserID]; ok {
		log.Printf("User %d already connected to room %s. Disconnecting old client.", client.UserID, r.QuizID)
		r.handleClientUnregister(oldClient)
	}

	r.Clients[client] = true
	r.clientsByUserID[client.UserID] = client // Add new client to map
	log.Printf("Client %d registered to room %s", client.UserID, r.QuizID)

	if _, ok := r.scores[client.UserID]; !ok {
		userName := "Player " + fmt.Sprintf("%d", client.UserID)
		r.scores[client.UserID] = &dtos.PlayerScore{UserID: client.UserID, UserName: userName, Score: 0}
	}

	playerInfo := &dtos.PlayerInfoPayload{UserID: client.UserID, UserName: r.scores[client.UserID].UserName}
	r.broadcastMessage("player_joined", playerInfo, client)
}

func (r *Room) handleClientUnregister(client *Client) {
	if _, ok := r.Clients[client]; ok {
		delete(r.Clients, client)
		delete(r.clientsByUserID, client.UserID) // Remove from clientsByUserID map
		close(client.Send)
		log.Printf("Client %d unregistered from room %s", client.UserID, r.QuizID)

		playerInfo := &dtos.PlayerInfoPayload{UserID: client.UserID, UserName: r.scores[client.UserID].UserName}
		r.broadcastMessage("player_left", playerInfo, nil)
	}
}

func (r *Room) broadcastMessage(msgType string, payload interface{}, exclude *Client) {
	payloadBytes, _ := json.Marshal(payload)
	msg := dtos.WebsocketMessage{Type: msgType, Payload: payloadBytes}
	msgBytes, _ := json.Marshal(msg)

	for client := range r.Clients {
		if client != exclude {
			select {
			case client.Send <- msgBytes:
			default:
				close(client.Send)
				delete(r.Clients, client)
			}
		}
	}
}