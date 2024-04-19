package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/venom1270/santorini/customerrors"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

type SessionState int

const (
	STATE_NAVIGATION = iota
	STATE_LOBBY_WAITING
	STATE_LOBBY_READY_WAIT
	STATE_LOBBY_READY
	STATE_LOBBY_WAITING_INPUT
	STATE_LOBBY_WAITING_PLAYERS
)

type ConnSession struct {
	ClientId string
	conn     net.Conn
	State    SessionState
	Lobby    *Lobby
}

func OfficialExample() {
	log.Println("Starting server!!!")

	authorizedKeysBytes, err := os.ReadFile("server/keys/id_rsa.pub")
	if err != nil {
		log.Fatalf("Failed to load authorized_keys, err: %v", err)
	}

	authorizedKeysMap := map[string]bool{}
	for len(authorizedKeysBytes) > 0 {
		pubKey, _, _, rest, err := ssh.ParseAuthorizedKey(authorizedKeysBytes)
		if err != nil {
			log.Fatal(err)
		}

		authorizedKeysMap[string(pubKey.Marshal())] = true
		authorizedKeysBytes = rest
	}

	// fmt.Println(authorizedKeysMap)

	// An SSH server is represented by a ServerConfig, which holds
	// certificate details and handles authentication of ServerConns.
	config := &ssh.ServerConfig{
		// Remove to disable password auth.
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			// Should use constant-time compare (or better, salt+hash) in
			// a production setting.
			if c.User() == "qwe" && string(pass) == "qwe" {
				return &ssh.Permissions{}, nil
			}
			return nil, fmt.Errorf("password rejected for %q", c.User())
		},

		// Remove to disable public key auth.
		PublicKeyCallback: func(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
			if authorizedKeysMap[string(pubKey.Marshal())] {
				return &ssh.Permissions{
					// Record the public key used for authentication.
					Extensions: map[string]string{
						"pubkey-fp": ssh.FingerprintSHA256(pubKey),
					},
				}, nil
			}
			return nil, fmt.Errorf("unknown public key for %q", c.User())
		},
	}

	privateBytes, err := os.ReadFile("server/keys/id_rsa")
	if err != nil {
		log.Fatal("Failed to load private key: ", err)
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		log.Fatal("Failed to parse private key: ", err)
	}
	config.AddHostKey(private)

	// Once a ServerConfig has been configured, connections can be
	// accepted.
	address := "0.0.0.0:2222"
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("failed to listen for connection: ", err)
	}

	// Accept connections forever...
	for {
		log.Printf("Listening on %s", address)
		nConn, err := listener.Accept()
		if err != nil {
			log.Fatal("failed to accept incoming connection: ", err)
		}

		log.Printf("Connection accepted: %s", nConn.RemoteAddr().String())

		go handleConnection(&nConn, config)
	}

	log.Println("Ending...")

}

func handleConnection(nConn *net.Conn, config *ssh.ServerConfig) {

	// Before use, a handshake must be performed on the incoming
	// net.Conn.
	conn, chans, reqs, err := ssh.NewServerConn(*nConn, config)
	if err != nil {
		log.Print("failed to handshake: ", err)
		return
	}
	log.Printf("logged in with key %s", conn.Permissions.Extensions["pubkey-fp"])

	var wg sync.WaitGroup
	defer wg.Wait()

	// The incoming Request channel must be serviced.
	// TODO: I DONT UNDERSTAND THIS!!!?!?!?!?
	wg.Add(1)
	go func() {
		ssh.DiscardRequests(reqs)
		wg.Done()
	}()

	// Service the incoming Channel channel.
	for newChannel := range chans {
		// Channels have a type, depending on the application level
		// protocol intended. In the case of a shell, the type is
		// "session" and ServerShell may be used to present a simple
		// terminal interface.
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}
		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Fatalf("Could not accept channel: %v", err)
		}

		log.Println("Accepting channel...")

		// Sessions have out-of-band requests such as "shell",
		// "pty-req" and "env".  Here we handle only the
		// "shell" request.
		wg.Add(1)
		go func(in <-chan *ssh.Request) {
			for req := range in {
				req.Reply(req.Type == "shell" || req.Type == "pty-req", nil)
				log.Printf("Request type: %s", req.Type)
			}
			wg.Done()
		}(requests)

		clientId := conn.User() + "#" + conn.RemoteAddr().String()
		connSession := ConnSession{clientId, *nConn, STATE_NAVIGATION, nil}

		handleTerminal(&channel, connSession, &wg)

	}

}

func handleTerminal(channel *ssh.Channel, connSession ConnSession, wg *sync.WaitGroup) {

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
					if _, ok := err.(customerrors.InfoError); ok {
						log.Printf("Error handling input, InfoError: %v", err)
						out = err.Error()
					} else {
						log.Printf("Error handling input, closing stream: %v", err)
						break
					}
				}
				term.Write([]byte(out))
				term.Write([]byte("\n"))
			} else {
				// In lobby (game)

				switch connSession.State {
				default:
					term.Write([]byte("Not implemented....... fin"))
				case STATE_LOBBY_WAITING:
					term.Write([]byte("Waiting in lobby...\n"))
					connSession.Lobby.wg.Done()
					connSession.Lobby.wg.Wait()
					connSession.State = STATE_LOBBY_READY_WAIT
				case STATE_LOBBY_READY_WAIT:
					term.Write([]byte("All players joined, press any button to ready up!"))
					line, _ := term.ReadLine()
					fmt.Printf(line) // Just so I don't get an error
					term.Write([]byte("READY! Waiting other players TODO..."))
					connSession.State = STATE_NAVIGATION
					connSession.Lobby = nil
				}

			}

		}
	}()
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
