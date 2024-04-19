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

	wg sync.WaitGroup
}

const MAX_WATCHERS = 10

var lobbyList []Lobby

func InitializeLobbyListTest() {
	lobbyList = []Lobby{
		{*game.NewState(), "Lobby1", make([]string, 2), 0, make([]string, 10), 5, sync.WaitGroup{}},
		{*game.NewState(), "Lobby2", make([]string, 2), 0, make([]string, 10), 0, sync.WaitGroup{}},
		{*game.NewState(), "Lobby3", make([]string, 2), 0, make([]string, 10), 0, sync.WaitGroup{}},
	}

	for i := 0; i < len(lobbyList); i++ {
		lobbyList[i].wg.Add(lobbyList[i].GameState.GetNumPlayers())
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
