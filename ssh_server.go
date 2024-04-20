package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"golang.org/x/crypto/ssh"
)

func StartServer() {
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

		clientId := conn.User() //+ "#" + conn.RemoteAddr().String()
		connSession := ConnSession{clientId, *nConn, STATE_NAVIGATION, nil, -1}

		handleTerminal(&channel, connSession, &wg)

	}

}
