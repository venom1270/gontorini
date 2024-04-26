package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/venom1270/santorini/customerrors"
	"github.com/venom1270/santorini/game"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/term"
)

type SessionState int

// Just an identifier cuz no enums good
const WATCHER = 555

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

const (
	PROMPT_INPUT = "> "
	PROMPT_WAIT  = ""
)

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
	term := terminal.NewTerminal(*channel, PROMPT_INPUT)

	wg.Add(1)
	go func() {
		defer func() {
			(*channel).Close()
			wg.Done()
		}()

		welcome, _ := handleInput("welcome", &connSession)
		term.Write([]byte(welcome))

		// This is an async reading channel - to register inputs while state is waiting
		c := make(chan string)
		go func() {
			for {
				line, err := term.ReadLine()
				if err != nil {
					// out := handleError(err)
					log.Printf("Error, closing stream: %v", err)
					(*channel).Close()
					// TODO: log that session disconnected
					break
				}

				select {
				case c <- line:
					fmt.Println("Line read!")
				default:
					fmt.Println("Line not read")
				}

			}
		}()

		for {

			switch connSession.State {
			default:
				term.Write([]byte("Not implemented....... fin\n"))
			case STATE_NAVIGATION:
				term.SetPrompt(PROMPT_INPUT)
				// In menu navigation

				/*line, err := term.ReadLine()
				if err != nil {
					log.Printf("Error, closing stream: %v", err)
					return // TODO: if I do this then the final printline does not exexute
				}*/
				line := <-c

				out, err := handleInput(line, &connSession)
				if err != nil {
					out = handleError(err)
					if out == "EXIT" {
						return
					}
				}
				term.Write([]byte(out))
				term.Write([]byte("\n"))

			case STATE_LOBBY_WAITING:
				term.SetPrompt(PROMPT_WAIT)
				if connSession.PlayerIndex == WATCHER {
					term.Write([]byte("Waiting for players to join and game to begin...\n"))
					connSession.Lobby.readyWaitGrup.Wait()
					connSession.State = STATE_LOBBY_PLAYING // TODO: set this on Lobby maybe??
				} else {
					// Player
					term.Write([]byte("Waiting in lobby...\n"))
					connSession.Lobby.wg.Done()
					connSession.Lobby.wg.Wait()
					connSession.State = STATE_LOBBY_READY_WAIT
				}

			case STATE_LOBBY_READY_WAIT:
				//term.SetPrompt(PROMPT_INPUT)
				term.Write([]byte("All players joined, press any button to ready up!\n"))
				<-c
				/*_, err := term.ReadLine()
				if err != nil {
					var out = handleError(err)
					if out == "EXIT" {
						return
					}
				}*/
				connSession.Lobby.readyWaitGrup.Done()
				term.Write([]byte("READY! Waiting other players...\n"))
				connSession.Lobby.readyWaitGrup.Wait()
				connSession.State = STATE_LOBBY_PLAYING

				//case STATE_LOBBY_READY:
			//	term.ReadLine()

			case STATE_LOBBY_PLAYING:
				term.SetPrompt(PROMPT_WAIT)
				// Simple 2 player game with preset worker locations, TODO more players
				gameStr := connSession.Lobby.GameState.GetStateStr()
				myIndex := connSession.PlayerIndex
				currentPlayer := connSession.Lobby.GameState.GetCurrentPlayer()
				term.Write([]byte(gameStr))
				if connSession.PlayerIndex == WATCHER {
					term.Write([]byte(fmt.Sprintf("Player %d turn.\n", currentPlayer+1)))
					connSession.Lobby.watcherWg.Add(1)
					connSession.Lobby.watcherWg.Wait()

					if connSession.Lobby.GameState.Victor > -1 {
						// Somebdoy won
						gameStr := connSession.Lobby.GameState.GetStateStr()
						term.Write([]byte(gameStr))
						term.Write([]byte(fmt.Sprintf("Player %d won the game!\n", connSession.Lobby.GameState.Victor+1)))
						connSession.State = STATE_NAVIGATION
					}

				} else if currentPlayer == connSession.PlayerIndex {
					term.SetPrompt(PROMPT_INPUT)
					term.Write([]byte("YOUR TURN!\n"))
					worker, movePos, buildPos, err := readMoveWorkerInput(term, &c)
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

					if connSession.Lobby.GameState.Victor != -1 {
						connSession.State = STATE_NAVIGATION
						term.Write([]byte("You won! Gained social score.\n\n"))
						if connSession.Lobby.NumWatchers > 0 {
							fmt.Printf("Num watchers: %d\n", connSession.Lobby.NumWatchers)
							connSession.Lobby.watcherWg.Done()
						}
						completeTurn(&connSession)
						time.AfterFunc(time.Second*2, func() { LobbyClose(connSession.Lobby.LobbyId) })
					} else {
						if connSession.Lobby.NumWatchers > 0 {
							connSession.Lobby.watcherWg.Done() // This has to be in both branches. LobbyClose messe with the waitgroup
						}
						completeTurn(&connSession)
					}

				} else {
					term.Write([]byte(fmt.Sprintf("Player %d turn...\n", currentPlayer+1)))
					connSession.Lobby.turnWaitGroups[myIndex].Wait()

					if connSession.Lobby.GameState.Victor != -1 {
						gameStr := connSession.Lobby.GameState.GetStateStr()
						term.Write([]byte(gameStr))
						term.Write([]byte(fmt.Sprintf("Sorry, you lost! Player %d won. Social score lost.\n\n", connSession.Lobby.GameState.Victor+1)))
						term.Write([]byte(PROMPT_INPUT)) // Workaround...
						connSession.State = STATE_NAVIGATION
					}
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

func readMoveWorkerInput(term *term.Terminal, c *chan string) (int, game.Position, game.Position, error) {
	var worker int
	var movePos game.Position
	var buildPos game.Position

	term.Write([]byte(" Select worker: "))
	/*line, err := term.ReadLine()
	if err != nil {
		return -1, game.Position{}, game.Position{}, customerrors.NewInfoError(err.Error())
	}*/
	line := <-*c
	worker, err := strconv.Atoi(line)
	if err != nil {
		return -1, game.Position{}, game.Position{}, customerrors.NewInfoError(err.Error())
	}
	term.Write([]byte(" Move location\n"))
	term.Write([]byte("  Column: "))
	line = <-*c
	/*line, err = term.ReadLine()
	if err != nil {
		return -1, game.Position{}, game.Position{}, customerrors.NewInfoError(err.Error())
	}*/
	movePos.Col, err = strconv.Atoi(line)
	if err != nil {
		return -1, game.Position{}, game.Position{}, customerrors.NewInfoError(err.Error())
	}
	term.Write([]byte("  Row: "))
	line = <-*c
	/*line, err = term.ReadLine()
	if err != nil {
		return -1, game.Position{}, game.Position{}, customerrors.NewInfoError(err.Error())
	}*/
	movePos.Row, err = strconv.Atoi(line)
	if err != nil {
		return -1, game.Position{}, game.Position{}, customerrors.NewInfoError(err.Error())
	}
	term.Write([]byte(" Build location\n"))
	term.Write([]byte("  Column: "))
	line = <-*c
	/*line, err = term.ReadLine()
	if err != nil {
		return -1, game.Position{}, game.Position{}, customerrors.NewInfoError(err.Error())
	}*/
	buildPos.Col, err = strconv.Atoi(line)
	if err != nil {
		return -1, game.Position{}, game.Position{}, customerrors.NewInfoError(err.Error())
	}
	term.Write([]byte("  Row: "))
	line = <-*c
	/*line, err = term.ReadLine()
	if err != nil {
		return -1, game.Position{}, game.Position{}, customerrors.NewInfoError(err.Error())
	}*/
	buildPos.Row, err = strconv.Atoi(line)
	if err != nil {
		return -1, game.Position{}, game.Position{}, customerrors.NewInfoError(err.Error())
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

	case strings.HasPrefix(input, "create ") && strings.Count(input, " ") == 2 && len(input) > 7:
		lobbyId := strings.Split(input, " ")[1]
		numPlayers, err := strconv.Atoi(strings.Split(input, " ")[2])
		if err != nil {
			return "", err
		}
		err = LobbyCreate(session, lobbyId, numPlayers)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Lobby '%s' created!", lobbyId), nil
	}
}

func handleError(err error) string {
	if _, ok := err.(customerrors.InfoError); ok {
		if err.Error() == "EOF" {
			log.Printf("Error handling input, closing stream: %v", err)
			return "EXIT"
		}
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
 - 'create <lobbyId> <numPlayers [2-4]>' to create a new Lobby
`
