// Package server provides a HTTP server implementation for a Battlesnake game.
//
// This package defines the data structures used in the game. It also provides a
// HTTP server that receives game events and responds with game moves.
//
// For more information about the Battlesnake game API, see https://docs.battlesnake.com/.
package server

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
)

const apiVersion = "1"

// Coord represents a coordinate on the game board.
type Coord struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Battlesnake represents a Battlesnake on the game board.
type Battlesnake struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	Health         int            `json:"health"`
	Body           []Coord        `json:"body"`
	Head           Coord          `json:"head"`
	Length         int            `json:"length"`
	Latency        string         `json:"latency"`
	Shout          string         `json:"shout"`
	Squad          string         `json:"squad"`
	Customizations Customizations `json:"customizations"`
}

// Customizations represents the customizable attributes of a Battlesnake. Namely its color
// as well as head and tail graphics.
type Customizations struct {
	Color string `json:"color"`
	Head  string `json:"head"`
	Tail  string `json:"tail"`
}

// Board represents the game board.
type Board struct {
	Height  int           `json:"height"`
	Width   int           `json:"width"`
	Food    []Coord       `json:"food"`
	Hazards []Coord       `json:"hazards"`
	Snakes  []Battlesnake `json:"snakes"`
}

// GameState represents state of the game at any given turn. It includes a Game instance, the
// current turn, a board instance and the player's battlesnake.
// It is passed with every 'move' request.
type GameState struct {
	Game  Game        `json:"game"`
	Turn  int         `json:"turn"`
	Board Board       `json:"board"`
	You   Battlesnake `json:"you"`
}

// Game represents a Battlesnake game.
type Game struct {
	ID      string  `json:"id"`
	Ruleset Ruleset `json:"ruleset"`
	Map     string  `json:"map"`
	Source  string  `json:"source"`
	Timeout int     `json:"timeout"`
}

// Ruleset describes the rule set of a Battlesnake game.
type Ruleset struct {
	Name     string          `json:"name"`
	Version  string          `json:"version"`
	Settings RulesetSettings `json:"settings"`
}

// RulesetSettings describes all settings for a game rule set.
type RulesetSettings struct {
	FoodSpawnChance     int            `json:"foodSpawnChance"`
	MinimumFood         int            `json:"minimumFood"`
	HazardDamagePerTurn int            `json:"hazardDamagePerTurn"`
	HazardMap           string         `json:"hazardMap"`
	HazardMapAuthor     string         `json:"hazardMapAuthor"`
	Royale              RoyaleSettings `json:"royale"`
	Squad               SquadSettings  `json:"squad"`
}

// RulesetSettings describes all settings specific to a Royale game
type RoyaleSettings struct {
	ShrinkEveryNTurns int `json:"shrinkEveryNTurns"`
}

// SquadSettings describes all settings specific to a Squad game
type SquadSettings struct {
	AllowBodyCollisions bool `json:"allowBodyCollisions"`
	SharedElimination   bool `json:"sharedElimination"`
	SharedHealth        bool `json:"sharedHealth"`
	SharedLength        bool `json:"sharedLength"`
}

// InfoResponse describes the body of the response to 'info' requests.
type InfoResponse struct {
	APIVersion string `json:"apiversion"`
	Author     string `json:"author"`
	Color      string `json:"color"`
	Head       string `json:"head"`
	Tail       string `json:"tail"`
	Version    string `json:"version"`
}

// InfoResponse describes the body of the response to 'move' requests.
type MoveResponse struct {
	Move  string `json:"move"`
	Shout string `json:"shout"`
}

// BattlesnakeServer represents a battlesnake server instance.
type BattlesnakeServer struct {
	http.Handler
	port     string
	info     *InfoResponse
	logger   *Logger
	moveFunc func(*GameState, *Logger) MoveResponse
}

// New creates a new instance of the battlesnake server with the specified port, InfoResponse, logger, loggerOpts, and moveFunc.
// The API version is set to the current version.
//
// Arguments:
// - port: The port number to listen on.
// - info: A pointer to an InfoResponse struct containing information about the battlesnake.
// - logger: A pointer to a log.Logger instance for to be used for logging.
// - loggerOpts: An integer representing the logging level options. Use bitwise or `|` to combine options. e.g. `LWarn|LErr|LDebug`.
// - moveFunc: A function that takes a pointer to a GameState and a pointer to a Logger, and returns a MoveResponse.
//
// Returns:
// - A pointer to a BattlesnakeServer instance.
//
// Example usage:
//
//	info := &server.InfoResponse{
//	    APIVersion: "1",
//	    Author:     "Your Battlesnake Author Name",
//	    Color:      "#888888",
//	    Head:       "default",
//	    Tail:       "default",
//	    Version:    "0.0.1",
//	}
//	server := server.New("8080", info, log.New(os.Stdout, "", 0), 0, moveFunc)
func New(port string, info *InfoResponse, logger *log.Logger, loggerOpts int, moveFunc func(*GameState, *Logger) MoveResponse) *BattlesnakeServer {
	info.APIVersion = apiVersion
	s := &BattlesnakeServer{
		port:     port,
		info:     info,
		logger:   NewLogger(logger, loggerOpts),
		moveFunc: moveFunc,
	}
	r := http.NewServeMux()
	r.HandleFunc("/", s.withRequestLogging(s.indexHandler()))
	r.HandleFunc("/start", s.withRequestLogging(s.startHandler()))
	r.HandleFunc("/end", s.withRequestLogging(s.endHandler()))
	r.HandleFunc("/move", s.withRequestLogging(s.moveHandler()))
	s.Handler = r
	return s
}

func (s *BattlesnakeServer) Start() error {
	ln, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return err
	}
	s.logger.Printf("START server running at %s...\n", ln.Addr())
	s.logger.Debug("request debug logging enabled")
	return http.Serve(ln, s)
}

func (s *BattlesnakeServer) indexHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("content-type", "application/json")
		if err := json.NewEncoder(w).Encode(s.info); err != nil {
			s.logger.Err("Failed to encode index response: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func (s *BattlesnakeServer) startHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := &GameState{}
		if err := json.NewDecoder(r.Body).Decode(&state); err != nil {
			s.logger.Err("Failed to decode start request body: %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		s.logger.Info("Game ID %s [Turn %d] Snake ID %s - Start", state.Game.ID, state.Turn, state.You.ID)
	}
}

func (s *BattlesnakeServer) endHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := &GameState{}
		if err := json.NewDecoder(r.Body).Decode(&state); err != nil {
			s.logger.Err("Failed to decode end request body: %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		s.logger.Info("Game ID %s [Turn %d] Snake ID %s - End", state.Game.ID, state.Turn, state.You.ID)
	}
}

func (s *BattlesnakeServer) moveHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := &GameState{}
		if err := json.NewDecoder(r.Body).Decode(&state); err != nil {
			s.logger.Err("Failed to decode move request body: %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		resp := s.moveFunc(state, s.logger)

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			s.logger.Err("Failed to encode move response, %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		s.logger.Info("Game ID %s [Turn %d] Snake ID %s - Move: %s", state.Game.ID, state.Turn, state.You.ID, resp.Move)
	}
}

func (s *BattlesnakeServer) withRequestLogging(next http.HandlerFunc) http.HandlerFunc {
	if s.logger.Enabled(LDebug) {
		return func(w http.ResponseWriter, r *http.Request) {
			b, err := io.ReadAll(r.Body)
			if err != nil {
				s.logger.Err("Failed to read request body: %s", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(b))
			s.logger.Debug("%s %s %s\n", r.Method, r.URL.Path, string(b))
			next(w, r)
		}
	}
	return next
}
