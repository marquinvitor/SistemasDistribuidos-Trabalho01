package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"
	pb "voting_system/proto" // IMPORTANT: Correct import path
)

const (
	ADMIN_SERVER_ADDR    = "localhost:8080"
	ADMIN_MULTICAST_ADDR = "224.0.0.1:9999"
	MAX_MSG_SIZE_ADMIN   = 4096
)

var adminID string

// Re-use sendRequest and readResponse (can be refactored into a shared client_util package)
func sendAdminRequest(conn net.Conn, reqType pb.GenericRequest_Type, payload proto.Message) error {
	var payloadBytes []byte
	var err error
	if payload != nil {
		payloadBytes, err = proto.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
	}
	genericReq := &pb.GenericRequest{Type: reqType, Payload: payloadBytes}
	data, err := proto.Marshal(genericReq)
	if err != nil {
		return fmt.Errorf("failed to marshal generic request: %w", err)
	}
	if err := binary.Write(conn, binary.BigEndian, uint32(len(data))); err != nil {
		return fmt.Errorf("failed to write message length: %w", err)
	}
	if _, err := conn.Write(data); err != nil {
		return fmt.Errorf("failed to write message data: %w", err)
	}
	log.Printf("Admin: Sent %s request", reqType)
	return nil
}

func readAdminResponse(conn net.Conn) (*pb.GenericResponse, error) {
	var msgLen uint32
	if err := binary.Read(conn, binary.BigEndian, &msgLen); err != nil {
		if err == io.EOF { return nil, io.EOF }
		return nil, fmt.Errorf("error reading admin message length: %w", err)
	}
	if msgLen > MAX_MSG_SIZE_ADMIN {
		return nil, fmt.Errorf("message from server too large for admin: %d bytes", msgLen)
	}
	msgBytes := make([]byte, msgLen)
	if _, err := io.ReadFull(conn, msgBytes); err != nil {
		return nil, fmt.Errorf("error reading admin message data: %w", err)
	}
	resp := &pb.GenericResponse{}
	if err := proto.Unmarshal(msgBytes, resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal admin generic response: %w", err)
	}
	log.Printf("Admin: Received response: Type=%s, Success=%t, Message='%s'", resp.Type, resp.Success, resp.Message)
	return resp, nil
}

func sendMulticastNote(loggedInAdminID string, content string) {
	mAddr, err := net.ResolveUDPAddr("udp", ADMIN_MULTICAST_ADDR)
	if err != nil {
		log.Printf("Admin: Error resolving multicast UDP address for sending: %v", err)
		return
	}
	// For sending, we "dial" to the multicast address.
	// The OS handles sending it to the multicast group.
	// We don't need to bind to the multicast address for sending.
	conn, err := net.DialUDP("udp", nil, mAddr)
	if err != nil {
		log.Printf("Admin: Error setting up UDP for multicast sending: %v", err)
		return
	}
	defer conn.Close()

	note := &pb.InformativeNote{
		AdminId:   loggedInAdminID,
		Content:   content,
		Timestamp: time.Now().Format(time.RFC3339Nano), // More precision
	}
	data, err := proto.Marshal(note)
	if err != nil {
		log.Printf("Admin: Error marshalling informative note: %v", err)
		return
	}

	_, err = conn.Write(data)
	if err != nil {
		log.Printf("Admin: Error sending multicast note: %v", err)
	} else {
		fmt.Println("Informative note sent via multicast.")
	}
}

func main() {
	conn, err := net.Dial("tcp", ADMIN_SERVER_ADDR)
	if err != nil {
		log.Fatalf("Admin: Failed to connect to server: %v", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Admin User ID: ")
	adminIDInput, _ := reader.ReadString('\n')
	adminID = strings.TrimSpace(adminIDInput)

	fmt.Print("Enter Admin Password: ")
	password, _ := reader.ReadString('\n')

	loginPayload := &pb.LoginPayload{
		UserId:   adminID,
		Password: strings.TrimSpace(password),
		UserType: pb.UserType_ADMIN,
	}
	if err := sendAdminRequest(conn, pb.GenericRequest_LOGIN, loginPayload); err != nil {
		log.Fatalf("Admin: Failed to send login request: %v", err)
	}

	resp, err := readAdminResponse(conn)
	if err != nil {
		if err == io.EOF { log.Fatalf("Admin: Connection closed by server during login.")}
		log.Fatalf("Admin: Failed to read login response: %v", err)
	}
	if !resp.Success || resp.Type != pb.GenericResponse_LOGIN_SUCCESS_ADMIN {
		log.Fatalf("Admin login failed: %s", resp.Message)
	}
	fmt.Println("Admin login successful!")
	fmt.Println(resp.Message)


	for {
		fmt.Println("\nAdmin Menu:")
		fmt.Println("1. Add Candidate")
		fmt.Println("2. Remove Candidate")
		fmt.Println("3. Send Informative Note (Multicast)")
		// Could add: 4. Start New Election (would reset server state, set new deadline)
		fmt.Println("4. Exit")
		fmt.Print("> ")

		choiceInput, _ := reader.ReadString('\n')
		choice := strings.TrimSpace(choiceInput)

		switch choice {
		case "1":
			fmt.Print("Enter new candidate ID: ")
			candID, _ := reader.ReadString('\n')
			fmt.Print("Enter new candidate Name: ")
			candName, _ := reader.ReadString('\n')

			addPayload := &pb.AddCandidatePayload{
				Candidate: &pb.Candidate{
					Id:   strings.TrimSpace(candID),
					Name: strings.TrimSpace(candName),
				},
			}
			if err := sendAdminRequest(conn, pb.GenericRequest_ADD_CANDIDATE, addPayload); err != nil {
				log.Printf("Admin: Failed to send add candidate request: %v", err)
				continue
			}
			resp, err := readAdminResponse(conn)
			if err != nil {
				if err == io.EOF { log.Println("Server closed connection."); return }
				log.Printf("Admin: Failed to read add candidate response: %v", err)
				continue
			}
			fmt.Println(resp.Message)

		case "2":
			fmt.Print("Enter candidate ID to remove: ")
			candIDToRemove, _ := reader.ReadString('\n')
			removePayload := &pb.RemoveCandidatePayload{
				CandidateId: strings.TrimSpace(candIDToRemove),
			}
			if err := sendAdminRequest(conn, pb.GenericRequest_REMOVE_CANDIDATE, removePayload); err != nil {
				log.Printf("Admin: Failed to send remove candidate request: %v", err)
				continue
			}
			resp, err := readAdminResponse(conn)
			if err != nil {
				if err == io.EOF { log.Println("Server closed connection."); return }
				log.Printf("Admin: Failed to read remove candidate response: %v", err)
				continue
			}
			fmt.Println(resp.Message)

		case "3":
			fmt.Print("Enter note content to multicast: ")
			noteContent, _ := reader.ReadString('\n')
			sendMulticastNote(adminID, strings.TrimSpace(noteContent))

		case "4":
			fmt.Println("Admin exiting.")
			return
		default:
			fmt.Println("Invalid option.")
		}
	}
}