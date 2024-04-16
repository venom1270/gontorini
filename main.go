package main

import "github.com/venom1270/santorini/game"

func main() {
	var gameState *game.State = game.NewState()
	gameState.SetPlayers(2)
	gameState.PrintState()
	//game.Setup(gameState)
	//game.PrintState(gameState)
}
