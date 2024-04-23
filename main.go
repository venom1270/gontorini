package main

import (
	"fmt"

	"github.com/venom1270/santorini/game"
)

func main() {
	//startGame()
	InitializeLobbyListTest()
	StartServer()
}

func startGame() {
	var gameState *game.State = game.NewState(2)
	gameState.SetupQuick()

	for over, _ := gameState.CheckGameOver(); over == -1; over, _ = gameState.CheckGameOver() {
		gameState.PrintState()
		var cp = gameState.GetCurrentPlayer() + 1
		fmt.Printf("Player %d turn\n", cp)
		worker, movePos, buildPos := readInput()
		var success, err = gameState.MoveWorker(gameState.GetCurrentPlayer(), (worker/cp - 1), movePos)
		if err != nil || !success {
			fmt.Printf("ERROR: %v\n", err)
			continue
		}
		success, err = gameState.Build(gameState.GetCurrentPlayer(), (worker/cp - 1), buildPos)
		if err != nil || !success {
			fmt.Printf("ERROR: %v\n", err)
			continue
		}
		fmt.Println("Success")
	}

	gameState.PrintState()

	fmt.Println("GAME OVER!")
	// HARDCODED TODO FIX
	if gameState.GetCurrentPlayer() == 0 {
		fmt.Println("Player 2 won!")
	} else {
		fmt.Println("Player 1 won!")
	}
}

func readInput() (int, game.Position, game.Position) {
	var worker int
	var movePos game.Position
	var buildPos game.Position

	fmt.Print(" Select worker: ")
	fmt.Scan(&worker)
	fmt.Println(" Move location")
	fmt.Print("  Column: ")
	fmt.Scan(&movePos.Col)
	fmt.Print("  Row: ")
	fmt.Scan(&movePos.Row)
	fmt.Println(" Build location")
	fmt.Print("  Column: ")
	fmt.Scan(&buildPos.Col)
	fmt.Print("  Row: ")
	fmt.Scan(&buildPos.Row)

	return worker, movePos, buildPos
}
