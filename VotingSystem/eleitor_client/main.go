package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"
	pb "voting_system/proto" // IMPORTANT: Correct import path
)

const (
	SERVER_ADDR         = "localhost:8080"
	MULTICAST_ADDR      = "224.0.0.1:9999"
	MAX_MSG_SIZE_CLIENT = 4096
)

var electorID string
var currentCandidates []*pb.Candidate // Cache candidates for voting

// Helper to send a framed proto message
func sendRequest(conn net.Conn, reqType pb.GenericRequest_Type, payload proto.Message) error {
	var payloadBytes []byte
	var err error
	if payload != nil {
		payloadBytes, err = proto.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
	}

	genericReq := &pb.GenericRequest{
		Type:    reqType,
		Payload: payloadBytes,
	}

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
	log.Printf("Sent %s request", reqType)
	return nil
}

// Helper to read a framed proto response
func readResponse(conn net.Conn) (*pb.GenericResponse, error) {
	var msgLen uint32
	if err := binary.Read(conn, binary.BigEndian, &msgLen); err != nil {
		if err == io.EOF {
			return nil, io.EOF
		}
		return nil, fmt.Errorf("error reading message length: %w", err)
	}

	if msgLen == 0 { // Handle cases where server might send an empty ack or error without payload.
		// This depends on server protocol, for now assume it always sends GenericResponse.
		// If it's a keep-alive or unexpected empty message, might need specific handling.
		log.Println("Received empty message (length 0) from server.")
		// We expect a GenericResponse, so if length is 0, it implies an issue or specific protocol design.
		// For this setup, assume GenericResponse is always sent.
	}


	if msgLen > MAX_MSG_SIZE_CLIENT {
		return nil, fmt.Errorf("message from server too large: %d bytes", msgLen)
	}

	msgBytes := make([]byte, msgLen)
	if _, err := io.ReadFull(conn, msgBytes); err != nil {
		return nil, fmt.Errorf("error reading message data: %w", err)
	}

	resp := &pb.GenericResponse{}
	if err := proto.Unmarshal(msgBytes, resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal generic response: %w", err)
	}
	log.Printf("Received response: Type=%s, Success=%t, Message='%s'", resp.Type, resp.Success, resp.Message)
	return resp, nil
}

func listenForMulticastNotes() {
	addr, err := net.ResolveUDPAddr("udp", MULTICAST_ADDR)
	if err != nil {
		log.Fatalf("Elector: Error resolving multicast UDP address: %v", err)
	}

	conn, err := net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		log.Printf("Elector: Error listening to multicast UDP: %v. Multicast may not be available.", err)
		return // Don't make this fatal, client can still function for TCP
	}
	defer conn.Close()

	err = conn.SetReadBuffer(8192)
	if err != nil {
		log.Printf("Elector: Error setting read buffer for multicast: %v", err)
	}

	log.Printf("Elector: Listening for admin notes on multicast group %s", MULTICAST_ADDR)
	buffer := make([]byte, 1500) // Typical MTU size for UDP
	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue // Read timeout, just continue listening
			}
			log.Printf("Elector: Error reading from multicast UDP: %v", err)
			time.Sleep(1 * time.Second) // Avoid busy-loop on persistent error
			continue
		}

		note := &pb.InformativeNote{}
		if err := proto.Unmarshal(buffer[:n], note); err != nil {
			log.Printf("Elector: Error unmarshalling informative note: %v", err)
			continue
		}
		fmt.Printf("\n📢 [ADMIN NOTE from %s @ %s]: %s\n> ", note.AdminId, note.Timestamp, note.Content)
	}
}

func main() {
	conn, err := net.Dial("tcp", SERVER_ADDR)
	if err != nil {
		log.Fatalf("Elector: Failed to connect to server: %v", err)
	}
	defer conn.Close()

	go listenForMulticastNotes()

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Elector User ID: ")
	userIDInput, _ := reader.ReadString('\n')
	electorID = strings.TrimSpace(userIDInput)

	fmt.Print("Enter Password: ")
	password, _ := reader.ReadString('\n')

	loginPayload := &pb.LoginPayload{
		UserId:   electorID,
		Password: strings.TrimSpace(password),
		UserType: pb.UserType_ELECTOR,
	}
	if err := sendRequest(conn, pb.GenericRequest_LOGIN, loginPayload); err != nil {
		log.Fatalf("Elector: Failed to send login request: %v", err)
	}

	resp, err := readResponse(conn)
	if err != nil {
		if err == io.EOF {
			log.Fatalf("Elector: Connection closed by server during login.")
		}
		log.Fatalf("Elector: Failed to read login response: %v", err)
	}

	if !resp.Success {
		log.Fatalf("Elector login failed: %s", resp.Message)
	}
	fmt.Println("Elector login successful!")
	fmt.Println(resp.Message) // Display message from server (e.g. voting open/closed)

	if resp.Type == pb.GenericResponse_LOGIN_SUCCESS_ELECTOR && len(resp.Payload) > 0 {
		clp := &pb.CandidateListPayload{}
		if err := proto.Unmarshal(resp.Payload, clp); err == nil {
			currentCandidates = clp.Candidates
			fmt.Printf("Voting is open until: %s\n", clp.VotingDeadline)
		}
	} else if resp.Type == pb.GenericResponse_ELECTION_RESULTS && len(resp.Payload) > 0 {
		erp := &pb.ElectionResultsPayload{}
         if err := proto.Unmarshal(resp.Payload, erp); err == nil {
            displayResults(erp)
        }
	}


	for {
		fmt.Println("\nElector Menu:")
		fmt.Println("1. View Candidates / Check Election Status")
		fmt.Println("2. Vote")
		fmt.Println("3. Exit")
		fmt.Print("> ")

		choiceInput, _ := reader.ReadString('\n')
		choice := strings.TrimSpace(choiceInput)

		switch choice {
		case "1":
			if err := sendRequest(conn, pb.GenericRequest_GET_CANDIDATES, nil); err != nil {
				log.Printf("Elector: Failed to send get candidates request: %v", err)
				continue
			}
			resp, err := readResponse(conn)
			if err != nil {
				if err == io.EOF { log.Println("Server closed connection."); return }
				log.Printf("Elector: Failed to read get candidates response: %v", err)
				continue
			}
			if !resp.Success {
				fmt.Printf("Error: %s\n", resp.Message)
				continue
			}

			if resp.Type == pb.GenericResponse_CANDIDATE_LIST {
				clp := &pb.CandidateListPayload{}
				if err := proto.Unmarshal(resp.Payload, clp); err != nil {
					log.Printf("Elector: Failed to unmarshal candidate list: %v", err)
					continue
				}
				currentCandidates = clp.Candidates
				fmt.Println("\n--- Candidates ---")
				if len(currentCandidates) == 0 {
					fmt.Println("No candidates available for voting yet.")
				}
				for i, c := range currentCandidates {
					fmt.Printf("%d. %s (%s)\n", i+1, c.Name, c.Id)
				}
				fmt.Printf("Voting Deadline: %s\n", clp.VotingDeadline)
			} else if resp.Type == pb.GenericResponse_ELECTION_RESULTS {
				erp := &pb.ElectionResultsPayload{}
				if err := proto.Unmarshal(resp.Payload, erp); err != nil {
					log.Printf("Elector: Failed to unmarshal election results: %v", err)
					continue
				}
                displayResults(erp)
			} else {
				fmt.Printf("Server Response: %s\n", resp.Message)
			}

		case "2":
			if len(currentCandidates) == 0 {
				fmt.Println("No candidates loaded. Please view candidates first (option 1).")
				continue
			}
			fmt.Println("\n--- Select Candidate to Vote ---")
			for i, c := range currentCandidates {
				fmt.Printf("%d. %s (%s)\n", i+1, c.Name, c.Id)
			}
			fmt.Print("Enter candidate number to vote for: ")
			voteChoiceInput, _ := reader.ReadString('\n')
			
			voteChoiceIdx, err := strconv.Atoi(strings.TrimSpace(voteChoiceInput))
			if err != nil || voteChoiceIdx < 1 || voteChoiceIdx > len(currentCandidates) {
				fmt.Println("Invalid choice. Please enter a number from the list.")
				continue
			}
			selectedCandidate := currentCandidates[voteChoiceIdx-1]

			votePayload := &pb.SubmitVotePayload{
				ElectorId:   electorID,
				CandidateId: selectedCandidate.Id,
			}
			if err := sendRequest(conn, pb.GenericRequest_SUBMIT_VOTE, votePayload); err != nil {
				log.Printf("Elector: Failed to send vote: %v", err)
				continue
			}
			resp, err := readResponse(conn)
			if err != nil {
				if err == io.EOF { log.Println("Server closed connection."); return }
				log.Printf("Elector: Failed to read vote response: %v", err)
				continue
			}
			fmt.Println(resp.Message) // Display server's ack/error for the vote

		case "3":
			fmt.Println("Exiting.")
			return
		default:
			fmt.Println("Invalid option. Please try again.")
		}
	}
}

func displayResults(erp *pb.ElectionResultsPayload) {
    fmt.Println("\n--- Election Results ---")
    fmt.Printf("%s\n", erp.StatusMessage)
    fmt.Printf("Total Votes: %d\n", erp.TotalVotes)
    if len(erp.CandidateResults) == 0 && erp.TotalVotes == 0 {
        fmt.Println("No candidates or votes were recorded in this election.")
    }
    for _, c := range erp.CandidateResults {
        fmt.Printf("- %s (%s): %d votes (%.2f%%)\n", c.Name, c.Id, c.VoteCount, c.Percentage)
    }
    if erp.Winner != nil && erp.Winner.Id != "" {
        fmt.Printf("Winner: %s (%s) with %d votes (%.2f%%)\n", erp.Winner.Name, erp.Winner.Id, erp.Winner.VoteCount, erp.Winner.Percentage)
    } else if erp.Winner != nil && erp.Winner.Name != "" {
        fmt.Printf("Winner: %s\n", erp.Winner.Name) // Handles "No winner (no votes)" or tie messages
    } else if erp.TotalVotes > 0 {
         fmt.Println("No single winner determined (possibly a tie not explicitly marked, or all candidates had 0 votes).")
    }
}