package main

import (
	"fmt"
	"strings"
	"sync"

	"github.com/venom1270/santorini/customerrors"
	"github.com/venom1270/santorini/game"
)

type Lobby struct {
	GameState   game.State
	LobbyId     string
	Players     []string
	NumPlayers  int
	Watchers    []string
	NumWatchers int

	wg             sync.WaitGroup   // Join waitgroup
	readyWaitGrup  sync.WaitGroup   // Join waitgroup
	turnWaitGroups []sync.WaitGroup // Game, turn waitgroup
}

const MAX_WATCHERS = 10

var lobbyList []Lobby

func InitializeLobbyListTest() {
	lobbyList = []Lobby{
		{*game.NewState(), "Lobby1", make([]string, 2), 0, make([]string, 10), 5, sync.WaitGroup{}, sync.WaitGroup{}, make([]sync.WaitGroup, 2)},
		{*game.NewState(), "Lobby2", make([]string, 2), 0, make([]string, 10), 0, sync.WaitGroup{}, sync.WaitGroup{}, make([]sync.WaitGroup, 2)},
		{*game.NewState(), "Lobby3", make([]string, 2), 0, make([]string, 10), 0, sync.WaitGroup{}, sync.WaitGroup{}, make([]sync.WaitGroup, 2)},
	}

	for i := 0; i < len(lobbyList); i++ {
		lobbyList[i].GameState.SetupQuick() // TODO: currently only a simple 2 player game with preset workers
		lobbyList[i].wg.Add(lobbyList[i].GameState.GetNumPlayers())
		lobbyList[i].readyWaitGrup.Add(lobbyList[i].GameState.GetNumPlayers())
		for j := 0; j < len(lobbyList[i].turnWaitGroups); j++ {
			lobbyList[i].turnWaitGroups[j].Add(j)
		}
	}
}

func LobbyListStr() string {
	var sb strings.Builder
	sb.WriteString("\nLobby list:\n")
	for _, lobby := range lobbyList {
		sb.WriteString(fmt.Sprintf("- %s (%d/%d players, %d/%d watchers)\n", lobby.LobbyId, lobby.NumPlayers, lobby.GameState.GetNumPlayers(), lobby.NumWatchers, MAX_WATCHERS))
	}
	return sb.String()
}

func LobbyJoin(connSession *ConnSession, lobbyId string, watcher bool) (*Lobby, error) {
	// TODO: error handling

	fmt.Println(lobbyList)

	lobby, err := findLobby(lobbyId)
	if err != nil {
		return &Lobby{}, err
	}

	if !watcher {
		// join
		if lobby.NumPlayers >= lobby.GameState.GetNumPlayers() {
			return &Lobby{}, customerrors.NewInfoError("lobby is full")
		}
		if game.IsInArray(lobby.Players, connSession.ClientId) {
			return &Lobby{}, customerrors.NewInfoError("already in lobby")
		}
		lobby.Players[lobby.NumPlayers] = connSession.ClientId
		connSession.PlayerIndex = lobby.NumPlayers
		lobby.NumPlayers++
	} else {
		// joinw
		if lobby.NumWatchers >= MAX_WATCHERS {
			return &Lobby{}, customerrors.NewInfoError("no more watchers allowed")
		}
		if game.IsInArray(lobby.Watchers, connSession.ClientId) {
			return &Lobby{}, customerrors.NewInfoError("already in lobby")
		}
		lobby.Watchers[lobby.NumWatchers] = connSession.ClientId
		lobby.NumWatchers++
	}

	connSession.State = STATE_LOBBY_WAITING
	connSession.Lobby = lobby

	return lobby, nil
}

func findLobby(lobbyId string) (*Lobby, error) {
	//for _, lobby := range lobbyList { IM LEAVING THIS GOOF HERE FOR FUTURE ME TO LAUGH AT MYSELF
	for i := 0; i < len(lobbyList); i++ {
		lobby := &lobbyList[i]
		if lobby.LobbyId == lobbyId {
			return lobby, nil
		}
	}
	return &Lobby{}, customerrors.NewInfoError("invalid lobby")
}
