package main

import (
	"fmt"
	"io"
	"log"
	"os/exec"

	"github.com/creack/pty"
	"github.com/gliderlabs/ssh"
)

func GilderLabs() {
	ssh.Handle(func(s ssh.Session) {
		cmd := exec.Command("top")
		ptyReq, _, isPty := s.Pty()
		if isPty {
			cmd.Env = append(cmd.Env, fmt.Sprintf("TERM=%s", ptyReq.Term))
			log.Println(ptyReq.Term)
			f, err := pty.Start(cmd)
			if err != nil {
				panic(err)
			}
			go func() {
				/*for win := range winCh {
					setWinsize(f, win.Width, win.Height)
				}*/
			}()
			go func() {
				io.Copy(f, s) // stdin
			}()
			io.Copy(s, f) // stdout
			cmd.Wait()
		} else {
			io.WriteString(s, "No PTY requested.\n")
			s.Exit(1)
		}
	})

	log.Println("starting ssh server on port 2222...")
	log.Fatal(ssh.ListenAndServe(":2222", nil, ssh.HostKeyFile("server/keys/id_rsa"),
		ssh.PasswordAuth(func(ctx ssh.Context, pass string) bool {
			return pass == "secret"
		})))
}

/*
func Test() {
	fmt.Println("Starting server!!!")

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

	fmt.Println(authorizedKeysMap)

	// An SSH server is represented by a ServerConfig, which holds
	// certificate details and handles authentication of ServerConns.
	sshConfig := &ssh.ServerConfig{
		// Remove to disable password auth.
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			// Should use constant-time compare (or better, salt+hash) in
			// a production setting.
			if c.User() == "qwe" && string(pass) == "qwe" {
				return nil, nil
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
	sshConfig.AddHostKey(private)

	// Once a ServerConfig has been configured, connections can be
	// accepted.
	listener, err := net.Listen("tcp", "0.0.0.0:2222")
	if err != nil {
		log.Fatal("failed to listen for connection: ", err)
	}

	log.Println("Listening...")
	for {
		tcpConn, err := listener.Accept()
		if err != nil {
			log.Fatal("failed to accept incoming connection: ", err)
		}
		log.Println("Got something, opening connection...")

		sshConn, channels, reqs, err := ssh.NewServerConn(tcpConn, sshConfig)

		log.Println("Connection open...")

		var wg sync.WaitGroup
		defer wg.Wait()

		// The incoming Request channel must be serviced.
		// The reqs variable is also a Go channel, but it contains global requests. We won't deal with these now, so let's disregard these completely
		wg.Add(1)
		go func() {
			ssh.DiscardRequests(reqs)
			wg.Done()
		}()

		log.Println("Handling channels...")
		handleChannels(sshConn, channels)

	}

}

func handleChannels(conn *ssh.ServerConn, chans <-chan ssh.NewChannel) {
	for newChannel := range chans {
		go handleChannel(conn, newChannel)
	}
}

func handleChannel(conn *ssh.ServerConn, newChannel ssh.NewChannel) {

	log.Printf("Hanling channel %s", newChannel.ChannelType())

	if t := newChannel.ChannelType(); t != "session" {
		log.Printf("Rejecting channel :(")
		_ = newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
		return
	}
	ch, requests, err := newChannel.Accept()
	if err != nil {
		log.Printf("error accepting channel: %v", err)
	}

	log.Println("Channel accepted!")

	for req := range requests {
		log.Printf("Checking req.Type: %s", req.Type)
		switch req.Type {
		case "env":
			// Save environment variables for later use
		case "pty-req":
			//reply(true, []byte("Hello comrade!!"))
			//req.Reply(true, []byte("Hello"))
		case "window-change":
			// Use the ContainerResize method on the Docker client later
		case "shell":
			log.Println("Replying...")
			ch.Write([]byte("HEllo!!!! COmrades!!"))
			// Create a container and run it
		case "exec":
			// Create a container and run it
		}
	}
}

func OfficialExample() {
	fmt.Println("Starting server!!!")

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

	fmt.Println(authorizedKeysMap)

	// An SSH server is represented by a ServerConfig, which holds
	// certificate details and handles authentication of ServerConns.
	config := &ssh.ServerConfig{
		// Remove to disable password auth.
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			// Should use constant-time compare (or better, salt+hash) in
			// a production setting.
			if c.User() == "qwe" && string(pass) == "qwe" {
				return nil, nil
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
	listener, err := net.Listen("tcp", "0.0.0.0:2222")
	if err != nil {
		log.Fatal("failed to listen for connection: ", err)
	}
	nConn, err := listener.Accept()
	if err != nil {
		log.Fatal("failed to accept incoming connection: ", err)
	}

	// Before use, a handshake must be performed on the incoming
	// net.Conn.
	conn, chans, reqs, err := ssh.NewServerConn(nConn, config)
	if err != nil {
		log.Fatal("failed to handshake: ", err)
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
				req.Reply(req.Type == "shell", nil)
				log.Printf("Request type: %s", req.Type)
			}
			wg.Done()
		}(requests)

		log.Println("Opening terminal...")
		term := terminal.NewTerminal(channel, "> ")

		log.Println("wg.Add(1)...")
		wg.Add(1)
		go func() {
			log.Println("Outer func...")
			defer func() {
				channel.Close()
				wg.Done()
			}()
			for {
				log.Println("GO for...")
				line, err := term.ReadLine()
				if err != nil {
					log.Println("ERROR!")
					log.Println(err)
					break
				}
				fmt.Println("QWEQWE")
				fmt.Println(line)
				log.Println("LOGGGG")
			}
		}()

	}

	log.Println("Ending...")
}
*/
