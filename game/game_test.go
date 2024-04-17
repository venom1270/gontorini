package game

import (
	"testing"
)

func TestSetWorkerPosition(t *testing.T) {
	var state = NewState()
	state.SetPlayers(3)

	state1, err := state.SetWorkerPosition(0, 0, Position{0, 0})
	if err != nil || state1.PlayerPositions[0] != 0 {
		t.Fatalf(`SetWorkerPosition(0, 0, 0, 0) = %d, %v, want 0, _`, state1.PlayerPositions[0], err)
	}

	state2, err := state.SetWorkerPosition(2, 1, Position{4, 2})
	if err != nil || state2.PlayerPositions[5] != 22 {
		t.Fatalf(`SetWorkerPosition(2, 1, 4, 2) = %d, %v, want 22, _`, state2.PlayerPositions[5], err)
	}

	state3, err := state.SetWorkerPosition(4, 4, Position{6, 7})
	if err == nil {
		t.Fatalf(`SetWorkerPosition(4, 4, 6, 7) = %d, %v, want _, error`, state3.PlayerPositions[5], err)
	}
}

func TestCheckGameOver(t *testing.T) {
	var state = NewState()

	player, err := state.CheckGameOver()
	expected := -1
	if player != expected || err != nil {
		t.Fatalf(`CheckGameOver() = %d, %v, want %d, error`, player, err, expected)

	}

	state.board[0] = LVL3
	state.PlayerPositions[3] = 0
	player, err = state.CheckGameOver()
	expected = 1
	if player != expected || err != nil {
		t.Fatalf(`CheckGameOver() = %d, %v, want %d, error`, player, err, expected)
	}
}

func TestGetValidMoveLocations(t *testing.T) {
	var state = NewState()

	state.PlayerPositions[0] = 11

	validPositions, err := state.GetValidMoveLocations(0, 0)
	expected := 8
	if len(validPositions) != expected || err != nil {
		t.Fatalf(`TestGetValidMoveLocations(0, 0) = %d, %v, want %d, error`, len(validPositions), err, expected)
	}

	state.PlayerPositions[2] = 12

	validPositions, err = state.GetValidMoveLocations(0, 0)
	expected = 7
	if len(validPositions) != expected || err != nil {
		t.Fatalf(`TestGetValidMoveLocations(0, 0) = %d, %v, want %d, error`, len(validPositions), err, expected)
	}

	validPositions, err = state.GetValidMoveLocations(1, 0)
	expected = 7
	if len(validPositions) != expected || err != nil {
		t.Fatalf(`TestGetValidMoveLocations(1, 0) = %d, %v, want %d, error`, len(validPositions), err, expected)
	}

	state.board[13] = LVL2
	validPositions, err = state.GetValidMoveLocations(1, 0)
	expected = 6
	if len(validPositions) != expected || err != nil {
		t.Fatalf(`TestGetValidMoveLocations(1, 0) = %d, %v, want %d, error`, len(validPositions), err, expected)
	}

	state.board[12] = LVL3
	validPositions, err = state.GetValidMoveLocations(1, 0)
	expected = 7
	if len(validPositions) != expected || err != nil {
		t.Fatalf(`TestGetValidMoveLocations(1, 0) = %d, %v, want %d, error`, len(validPositions), err, expected)
	}

	//state.PrintState()

}

func TestGetValidBuildLocations(t *testing.T) {
	var state = NewState()

	state.SetWorkerPosition(0, 0, Position{2, 2})
	state.SetWorkerPosition(0, 1, Position{1, 2})
	state.SetWorkerPosition(1, 0, Position{2, 1})
	state.board[13] = LVL4
	validPositions, err := state.GetValidBuildLocations(0, 0)
	expected := 5
	if len(validPositions) != expected || err != nil {
		t.Fatalf(`GetValidBuildLocations(0, 0) = %d, %v, want %d, error`, len(validPositions), err, expected)
	}

	state.SetWorkerPosition(0, 1, Position{4, 4})
	state.board[23] = LVL4
	state.board[19] = LVL4
	validPositions, err = state.GetValidBuildLocations(0, 1)
	expected = 1
	if len(validPositions) != expected || err != nil {
		t.Fatalf(`GetValidBuildLocations(0, 1) = %d, %v, want %d, error`, len(validPositions), err, expected)
	}

	state.SetWorkerPosition(0, 0, Position{3, 3})
	validPositions, err = state.GetValidBuildLocations(0, 1)
	expected = 0
	if len(validPositions) != expected || err != nil {
		t.Fatalf(`GetValidBuildLocations(0, 1) = %d, %v, want %d, error`, len(validPositions), err, expected)
	}
}

func TestBuild(t *testing.T) {
	var state = NewState()
	state.currentPlayer = 0
	state.SetWorkerPosition(0, 0, Position{3, 3})
	state.SetWorkerPosition(0, 1, Position{4, 3})
	state.SetWorkerPosition(1, 0, Position{3, 4})
	state.SetWorkerPosition(1, 1, Position{4, 4})

	success, err := state.Build(0, 0, Position{3, 3})
	expected := false
	if success != expected || err != nil {
		t.Fatalf(`Build(0, 0, Position{3, 3}) = %t, %v, want %t, error`, success, err, expected)
	}

	state.currentPlayer = 1
	success, err = state.Build(1, 1, Position{3, 3})
	expected = false
	if success != expected || err != nil {
		t.Fatalf(`Build(1, 1, Position{3, 3}) = %t, %v, want %t, error`, success, err, expected)
	}

	success, err = state.Build(1, 1, Position{5, 5})
	expected = false
	if success != expected || err == nil {
		t.Fatalf(`Build(0, 0, Position{3, 3}) = %t, %v, want %t, error`, success, err, expected)
	}

	state.currentPlayer = 0
	success, err = state.Build(0, 0, Position{2, 2})
	state.board[12] = LVL3
	expected = true
	if success != expected || err != nil {
		t.Fatalf(`Build(0, 0, Position{2, 2} = %t, %v, want %t, error`, success, err, expected)
	}

}
