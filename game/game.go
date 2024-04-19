package game

import (
	"errors"
	"fmt"
	"log"
)

// Board: 0 (empty) -> 1 (lvl1) -> 2 (lvl2) -> 3 (lvl3) -> 4 (lvl4) -> capped)
const (
	EMPTY = 0
	LVL1  = 1
	LVL2  = 2
	LVL3  = 3
	LVL4  = 4
)

type State struct {
	board           [25]int
	PlayerPositions [8]int // We might not need this, optimization only
	numPlayers      int
	currentPlayer   int
}

type Position struct {
	Row, Col int
}

func (s *State) SetPlayers(players int) bool {
	if players >= 2 && players <= 4 {
		s.numPlayers = players
		return true
	} else {
		return false
	}
}

func NewState() *State {
	var state = State{}
	state.currentPlayer = 0
	state.numPlayers = 2
	for i := 0; i < 8; i++ {
		state.PlayerPositions[i] = -100
	}
	return &state
}

func (s *State) Setup() {
	s.PlacePlayers()
	fmt.Println("Setup complete!")
}

func (s *State) SetupQuick() {
	s.SetPlayers(2)
	s.SetWorkerPosition(0, 0, Position{1, 1})
	s.SetWorkerPosition(0, 1, Position{3, 3})
	s.SetWorkerPosition(1, 0, Position{3, 1})
	s.SetWorkerPosition(1, 1, Position{1, 3})
	s.board[0] = LVL3
	s.board[6] = LVL2
	s.currentPlayer = 0
	fmt.Println("Quick setup done!")
}

func (s *State) PlacePlayers() {
	for i := 0; i < s.numPlayers; i++ {
		for !s.tryPlacePlayer(i, 0) {
		}
		for !s.tryPlacePlayer(i, 1) {
		}
	}
}

func (s *State) tryPlacePlayer(player int, worker int) bool {
	fmt.Printf("Place worker %d for player %d\n", worker+1, player+1)
	var pos Position
	fmt.Printf("i: ")
	pos.Row = readNumber(1, 5) - 1
	fmt.Printf("j: ")
	pos.Col = readNumber(1, 5) - 1
	var _, err = s.SetWorkerPosition(player, worker, pos)
	if err != nil {
		log.Fatal(err)
	}
	return true
}

func (s *State) SetWorkerPosition(player, worker int, pos Position) (*State, error) {
	if s == nil {
		return nil, errors.New("invalid state")
	}
	if player < 0 || player >= s.numPlayers {
		return s, errors.New("invalid player")
	}
	if worker < 0 || worker > 1 {
		return s, errors.New("invalid worker")
	}
	if !s.isValidPosition(pos) {
		return s, errors.New("invalid position")
	}
	var _, position = s.getBoardLocation(pos)
	if !s.isFieldEmpty(pos) {
		return s, errors.New("position is not empty")
	}
	s.PlayerPositions[(player*2)+worker] = position
	return s, nil
}

func readNumber(low, high int) int {
	var i int = -1
	first := true
	for i < low || i > high {
		if first {
			first = false
		} else {
			fmt.Println("Invalid input!")
		}
		fmt.Scanln(&i)
	}
	return i
}

// Value, location/position
func (s *State) getBoardLocation(pos Position) (int, int) {
	return s.board[pos.Row*5+pos.Col], pos.Row*5 + pos.Col
}

// i, j
func getBoardCoordianates(x int) Position {
	return Position{x % 5, x / 5}
}

func (s *State) getFieldLevel(pos Position) int {
	return s.board[pos.Row*5+pos.Col]
}

func (s *State) isFieldEmpty(pos Position) bool {
	var _, position = s.getBoardLocation(pos)
	for ii := 0; ii < s.numPlayers*2; ii++ {
		if s.PlayerPositions[ii] == position {
			return false
		}
	}
	return true
}

func (s *State) PrintState() {
	fmt.Println("Board")
	fmt.Print("\n \t")
	for i := 0; i < 5; i++ {
		fmt.Print(i)
		fmt.Print(" ")
	}
	fmt.Println()
	fmt.Println()
	for i := 0; i < 5; i++ {
		fmt.Print(i)
		fmt.Print("\t")
		for j := 0; j < 5; j++ {
			fmt.Print(s.board[i*5+j])
			fmt.Print(" ")
		}
		fmt.Println()
	}
	fmt.Println("\nWorkers")

	for i := 0; i < 5; i++ {
		//fmt.Print(i)
		fmt.Print("\t")
		for j := 0; j < 5; j++ {
			_, index := s.getBoardLocation(Position{i, j})
			for p := 0; p < s.numPlayers*2; p++ {
				if s.PlayerPositions[p] == index {
					fmt.Print(p + 1)
					fmt.Print(" ")
					index = -1
					break
				}
			}
			if index != -1 {
				fmt.Print(". ")
			}
		}
		fmt.Println()
	}

	fmt.Println()

	for i := 0; i < s.numPlayers; i++ {
		var worker1Pos = getBoardCoordianates(s.PlayerPositions[i*2])
		var worker2Pos = getBoardCoordianates(s.PlayerPositions[i*2+1])
		fmt.Printf(" Player %d, worker %d: %v (%d)\n", i+1, 1, worker1Pos, s.PlayerPositions[i*2])
		fmt.Printf(" Player %d, worker %d: %v (%d)\n", i+1, 2, worker2Pos, s.PlayerPositions[i*2+1])
	}
}

func (s *State) CheckGameOver() (int, error) {
	if s == nil {
		return -1, errors.New("invalid state")
	}
	// Check which fields bring victory
	var victoryFields []int
	for i := 0; i < 25; i++ {
		if s.board[i] == LVL3 {
			victoryFields = append(victoryFields, i)
		}
	}
	// Check if any player won
	for i := 0; i < s.numPlayers*2; i++ {
		for _, victoryField := range victoryFields {
			if s.PlayerPositions[i] == victoryField {
				return i / 2, nil
			}
		}
	}
	return -1, nil
}

func (s *State) GetWorkerPosition(player, worker int) (int, error) {
	if s == nil {
		return -1, errors.New("invalid state")
	}
	if player < 0 || player >= s.numPlayers {
		return -1, errors.New("invalid player")
	}
	if worker < 0 || worker > 1 {
		return -1, errors.New("invalid worker")
	}
	return s.PlayerPositions[(player*2)+worker], nil
}

func posToPosition(x int) Position {
	return Position{x / 5, x % 5}
}

func positionToPos(pos Position) int {
	return pos.Row*5 + pos.Col
}

func (s *State) GetValidMoveLocations(player, worker int) ([]Position, error) {
	var workerPosition, err = s.GetWorkerPosition(player, worker)
	if err != nil {
		return nil, err
	}
	var workerHeight = s.board[workerPosition]
	var position = posToPosition(workerPosition)
	var valuesToCheck = []Position{
		{position.Row + 1, position.Col + 1},
		{position.Row + 1, position.Col - 1},
		{position.Row - 1, position.Col + 1},
		{position.Row - 1, position.Col - 1},
		{position.Row + 1, position.Col},
		{position.Row - 1, position.Col},
		{position.Row, position.Col + 1},
		{position.Row, position.Col - 1},
	}
	var validPositions = []Position{}
	for _, pos := range valuesToCheck {
		if !s.isValidPosition(pos) {
			continue
		}
		var p = positionToPos(pos)
		var b = s.board[p]
		if b != LVL4 && b <= workerHeight+1 && s.isFieldEmpty(pos) {
			validPositions = append(validPositions, pos)
		}
	}
	return validPositions, nil
}

func (s *State) isValidPosition(pos Position) bool {
	return pos.Col >= 0 && pos.Col < 5 && pos.Row >= 0 && pos.Row < 5
}

func (s *State) GetValidBuildLocations(player, worker int) ([]Position, error) {
	var workerPosition, err = s.GetWorkerPosition(player, worker)
	if err != nil {
		return nil, err
	}
	var position = posToPosition(workerPosition)
	var valuesToCheck = []Position{
		{position.Row + 1, position.Col + 1},
		{position.Row + 1, position.Col - 1},
		{position.Row - 1, position.Col + 1},
		{position.Row - 1, position.Col - 1},
		{position.Row + 1, position.Col},
		{position.Row - 1, position.Col},
		{position.Row, position.Col + 1},
		{position.Row, position.Col - 1},
	}
	var validPositions = []Position{}
	for _, pos := range valuesToCheck {
		if !s.isValidPosition(pos) {
			continue
		}
		var p = positionToPos(pos)
		var b = s.board[p]
		if b != LVL4 && s.isFieldEmpty(pos) {
			validPositions = append(validPositions, pos)
		}
	}
	return validPositions, nil
}

func (s *State) MoveWorker(player, worker int, pos Position) (bool, error) {
	if player != s.currentPlayer {
		return false, errors.New("invalid player")
	}
	if !s.isValidPosition(pos) {
		return false, errors.New("invalid position")
	}

	var validPositons, err = s.GetValidMoveLocations(player, worker)
	if err != nil {
		return false, err
	}
	if !IsInArray(validPositons, pos) {
		return false, nil
	}

	s.SetWorkerPosition(player, worker, pos)
	//s.currentPlayer = (s.currentPlayer + 1) % s.numPlayers
	return true, nil
}

func (s *State) Build(player, worker int, pos Position) (bool, error) {
	if player != s.currentPlayer {
		return false, errors.New("invalid player, not current")
	}
	if !s.isValidPosition(pos) {
		return false, errors.New("invalid position")
	}

	var validPositons, err = s.GetValidBuildLocations(player, worker)
	if err != nil {
		return false, err
	}
	if !IsInArray(validPositons, pos) {
		return false, nil
	}

	_, index := s.getBoardLocation(pos)
	// Build
	s.board[index]++
	// Next player turn
	s.currentPlayer = (s.currentPlayer + 1) % s.numPlayers
	return true, nil
}

func (s *State) GetCurrentPlayer() int {
	return s.currentPlayer
}

func (s *State) GetNumPlayers() int {
	return s.numPlayers
}

func IsInArray[T comparable](arr []T, el T) bool {
	for _, v := range arr {
		if v == el {
			return true
		}
	}
	return false
}
