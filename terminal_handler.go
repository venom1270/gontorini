package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/venom1270/santorini/customerrors"
	"github.com/venom1270/santorini/game"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/term"
)

type SessionState int

const (
	STATE_NAVIGATION = iota
	STATE_LOBBY_WAITING
	STATE_LOBBY_READY_WAIT
	STATE_LOBBY_READY
	STATE_LOBBY_PLAYING
)

type ConnSession struct {
	ClientId    string
	conn        net.Conn
	State       SessionState
	Lobby       *Lobby
	PlayerIndex int
}

var activeSessions []*ConnSession

func findActiveSession(clientId string) *ConnSession {
	for i := 0; i < len(activeSessions); i++ {
		if activeSessions[i].ClientId == clientId {
			return activeSessions[i]
		}
	}
	return nil
}

func handleTerminal(channel *ssh.Channel, connSession ConnSession, wg *sync.WaitGroup) {

	activeSession := findActiveSession(connSession.ClientId)
	if activeSession != nil {
		fmt.Printf("Active session exists for client '%s'! Restoring session...\n", connSession.ClientId)
		connSession = *activeSession
	} else {
		fmt.Println("Creating session TODO: only if joined lobby... i guess")
		activeSessions = append(activeSessions, &connSession)
	}

	log.Println("Opening terminal...")
	log.Printf("ClientId: %s", connSession.ClientId)
	term := terminal.NewTerminal(*channel, "> ")

	wg.Add(1)
	go func() {
		defer func() {
			(*channel).Close()
			wg.Done()
		}()

		welcome, _ := handleInput("welcome", &connSession)
		term.Write([]byte(welcome))

		for {

			if connSession.Lobby == nil {
				// In menu navigation

				line, err := term.ReadLine()
				if err != nil {
					log.Printf("Error, closing stream: %v", err)
					break
				}

				out, err := handleInput(line, &connSession)
				if err != nil {
					out = handleError(err)
					if out == "EXIT" {
						return
					}
				}
				term.Write([]byte(out))
				term.Write([]byte("\n"))
			} else {
				// In lobby (game)

				switch connSession.State {
				default:
					term.Write([]byte("Not implemented....... fin\n"))
				case STATE_LOBBY_WAITING:
					term.Write([]byte("Waiting in lobby...\n"))
					connSession.Lobby.wg.Done()
					connSession.Lobby.wg.Wait()
					connSession.State = STATE_LOBBY_READY_WAIT
				case STATE_LOBBY_READY_WAIT:
					term.Write([]byte("All players joined, press any button to ready up!\n"))
					_, err := term.ReadLine()
					var out = handleError(err)
					if err != nil {
						if out == "EXIT" {
							return
						}
					}
					connSession.Lobby.readyWaitGrup.Done()
					term.Write([]byte("READY! Waiting other players...\n"))
					connSession.Lobby.readyWaitGrup.Wait()
					connSession.State = STATE_LOBBY_PLAYING

					//case STATE_LOBBY_READY:
				//	term.ReadLine()

				case STATE_LOBBY_PLAYING:
					// Simple 2 player game with preset worker locations, TODO more players
					gameStr := connSession.Lobby.GameState.GetStateStr()
					myIndex := connSession.PlayerIndex
					term.Write([]byte(gameStr))
					if connSession.Lobby.GameState.GetCurrentPlayer() == connSession.PlayerIndex {
						term.Write([]byte("YOUR TURN!\n"))
						worker, movePos, buildPos, err := readMoveWorkerInput(term)
						if err != nil {
							var out = handleError(err)
							if out == "EXIT" {
								return
							} else {
								term.Write([]byte(fmt.Sprintf("Error reading input: %v\n", err)))
								break
							}
						}
						fmt.Println(worker - myIndex*2 - 1)
						success, err := connSession.Lobby.GameState.MoveWorkerAndBuild(myIndex, worker-myIndex*2-1, movePos, buildPos)
						if err != nil {
							term.Write([]byte(fmt.Sprintf("**** Movement error: %v\n", err)))
							break
						}
						if !success {
							term.Write([]byte("**** Movement unsuccessful, try again\n"))
							break
						}
						completeTurn(&connSession)
					} else {
						term.Write([]byte(fmt.Sprintf("Player %d turn...\n", connSession.Lobby.GameState.GetCurrentPlayer()+1)))
						connSession.Lobby.turnWaitGroups[myIndex].Wait()
					}

					// TODO: GAME END
				}

			}

		}

		// If terminal waiting for wg, this does not execute... because waiting...
		// Research what happens in different scenarios
		fmt.Println("ENDING SESSION!!!!!")
	}()
}

func completeTurn(conn *ConnSession) {
	for i := 0; i < conn.Lobby.NumPlayers; i++ {
		if i == conn.PlayerIndex {
			conn.Lobby.turnWaitGroups[i].Add(conn.Lobby.NumPlayers - 1)
		} else {
			conn.Lobby.turnWaitGroups[i].Done()
		}
	}
}

func readMoveWorkerInput(term *term.Terminal) (int, game.Position, game.Position, error) {
	var worker int
	var movePos game.Position
	var buildPos game.Position

	term.Write([]byte(" Select worker: "))
	line, err := term.ReadLine()
	if err != nil {
		return -1, game.Position{}, game.Position{}, err
	}
	worker, err = strconv.Atoi(line)
	if err != nil {
		return -1, game.Position{}, game.Position{}, err
	}
	term.Write([]byte(" Move location\n"))
	term.Write([]byte("  Column: "))
	line, err = term.ReadLine()
	if err != nil {
		return -1, game.Position{}, game.Position{}, err
	}
	movePos.Col, err = strconv.Atoi(line)
	if err != nil {
		return -1, game.Position{}, game.Position{}, err
	}
	term.Write([]byte("  Row: "))
	line, err = term.ReadLine()
	if err != nil {
		return -1, game.Position{}, game.Position{}, err
	}
	movePos.Row, err = strconv.Atoi(line)
	if err != nil {
		return -1, game.Position{}, game.Position{}, err
	}
	term.Write([]byte(" Build location\n"))
	term.Write([]byte("  Column: "))
	line, err = term.ReadLine()
	if err != nil {
		return -1, game.Position{}, game.Position{}, err
	}
	buildPos.Col, err = strconv.Atoi(line)
	if err != nil {
		return -1, game.Position{}, game.Position{}, err
	}
	term.Write([]byte("  Row: "))
	line, err = term.ReadLine()
	if err != nil {
		return -1, game.Position{}, game.Position{}, err
	}
	buildPos.Row, err = strconv.Atoi(line)
	if err != nil {
		return -1, game.Position{}, game.Position{}, err
	}

	return worker, movePos, buildPos, nil
}

func handleInput(input string, session *ConnSession) (string, error) {
	switch {
	default:
		return fmt.Sprintf("Unrecognized command: '%s'", input), nil
	case input == "exit":
		return "", errors.New("client sent 'exit' command")
	case input == "help":
		return HELP_STRING, nil
	case input == "welcome":
		return fmt.Sprintf(WELCOME_STRING, session.ClientId), nil

	case input == "list":
		return LobbyListStr(), nil
	case strings.HasPrefix(input, "joinw ") && strings.Count(input, " ") == 1 && len(input) > 6:
		_, err := LobbyJoin(session, strings.Split(input, " ")[1], true)
		if err != nil {
			return "", err
		}
		return "Lobby joined as watcher!", nil
	case strings.HasPrefix(input, "join ") && strings.Count(input, " ") == 1 && len(input) > 5:
		_, err := LobbyJoin(session, strings.Split(input, " ")[1], false)
		if err != nil {
			return "", err
		}
		return "Lobby joined as player!", nil
	}
}

func handleError(err error) string {
	if _, ok := err.(customerrors.InfoError); ok {
		log.Printf("Error handling input, InfoError: %v", err)
		return err.Error()
	} else {
		log.Printf("Error handling input, closing stream: %v", err)
		return "EXIT"
	}
}

const WELCOME_STRING = `
*** Welcome to simple Gontorini server! ***

It allows hosting and playing simple Gontorini (some refer to it as Santorini) simply over SSH!
Type 'help' for more info on how to interact with this terminal.
You can type when you see the '>' symbol.

Your id: %s

`

const HELP_STRING = `
You can use these commands to navigate:
 - 'exit' to quit the session
 - 'list' to show currently open lobbies
 - 'join <lobbyId>' to join a lobby
 - 'joinw <lobbyId>' to join as a watcher
`
