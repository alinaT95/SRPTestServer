package main

import (
	srp "SRPTestServer/srp"
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
)

const group = srp.RFC5054Group2048


var users = map[string]*User{}

func main() {
	fmt.Println("Start server...")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "%s is a simple example server of the opaque package. It can be used together with cmd/client.\nUsage:\n", os.Args[0])
		flag.PrintDefaults()
	}
	addr := flag.String("l", ":9999", "Address to listen on.")
	flag.Parse()

	var err error
	//var sk []byte

	if err != nil {
		panic(err)
	}

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			continue
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	fmt.Println("Got connection from %s", conn.RemoteAddr())
	if err := doHandleConn(conn); err != nil {
		fmt.Println("Error happened in handleConn: %s\n", err)
	}
}

func doHandleConn(conn net.Conn) error {
	r := bufio.NewReader(conn)
	fmt.Println("Start connection handling...")
	cmd, err := srp.Read(r)
	if err != nil {
		return err
	}
	fmt.Println("Command from client = " + string(cmd))
	w := bufio.NewWriter(conn)
	switch string(cmd) {
	case "pwreg":
		if err := handlePwReg(r, w); err != nil {
			return fmt.Errorf("pwreg: %s", err)
		}
	case "auth":
		if err := handleAuth(r, w); err != nil {
			return fmt.Errorf("auth: %s", err)
		}
	default:
		return fmt.Errorf("Unknown command '%s'\n", string(cmd))
	}
	return nil
}

type User struct {
	userName string
	verifier string
	salt string
}

type PwRegMsg1 struct {
	UserName string
	Verifier string
	Salt string
}

func handlePwReg(r *bufio.Reader, w *bufio.Writer) error {
	fmt.Println("Start client registration...")
	data1, err := srp.Read(r)
	if err != nil {
		return err
	}

	fmt.Println(string(data1))

	var msg1 PwRegMsg1
	if err := json.Unmarshal(data1, &msg1); err != nil {
		return err
	}

	fmt.Println(msg1)

	fmt.Println("Got data from client #1:")
	fmt.Println("====================================")
	fmt.Println("Username:")
	fmt.Println(msg1.UserName)
	fmt.Println("Verifier:")
	fmt.Println(msg1.Verifier)
	fmt.Println("Salt:")
	fmt.Println(msg1.Salt)
	fmt.Println("====================================")

	var user = &User{
		userName: msg1.UserName,
		salt: msg1.Salt,
		verifier: msg1.Verifier,
	}

	fmt.Println("Added user: " + user.userName)
	users[user.userName] = user

	fmt.Println(users)
	fmt.Println(len(users))

	return nil
}

type AuthMsgFromServer struct {
	Salt string
	B string
}

type AuthMsgFromClient struct {
	UserName string
	A string
}

func handleAuth(r *bufio.Reader, w *bufio.Writer) error {
	fmt.Println("Start client authentication...")
	data1, err := srp.Read(r)
	if err != nil {
		return err
	}

	var msg1 AuthMsgFromClient
	if err := json.Unmarshal(data1, &msg1); err != nil {
		return err
	}

	fmt.Println(msg1)

	var user, ok = users[msg1.UserName]
	if !ok {
		if err := srp.Write(w, []byte("No such user")); err != nil {
			return err
		}
		return fmt.Errorf("No such user")
	}

	fmt.Println("User with username " + user.userName + " is found." )

	v := new(big.Int)
	v.SetString(user.verifier, 16)

	server := srp.NewSRPServer(srp.KnownGroups[group], v, nil)
	if server == nil {
		fmt.Println("Couldn't set up server")
	}

	A := new(big.Int)
	A.SetString(msg1.A, 16)

	if err = server.SetOthersPublic(A); err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}

	// server sends its ephemeral public key, B, to client
	// client sets it as others public key.
	B := new(big.Int)
	if B = server.EphemeralPublic(); B == nil {
		fmt.Println("server couldn't make B")
	}

	// server can now make the key.
	serverKey, err := server.Key()
	if err != nil || serverKey == nil {
		fmt.Printf("something went wrong making server key: %s\n", err)
	}

	fmt.Println("CS = " + toHexInt(server.PremasterKey))

	var authMsgFromServer2 = AuthMsgFromServer{
		Salt: user.salt,
		B: toHexInt(B),
	}

	data3, err := json.Marshal(authMsgFromServer2)
	if err != nil {
		return err
	}

	if err := srp.Write(w, data3); err != nil {
		return err
	}

	return nil
}

func toHexInt(n *big.Int) string {
	return fmt.Sprintf("%x", n) // or %x or upper case
}