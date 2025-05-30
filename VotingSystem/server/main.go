package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"
	pb "voting_system/proto" // IMPORTANT: Correct import path
)

const (
	TCP_PORT        = ":8080"
	VOTING_DURATION = 5 * time.Minute
	MAX_MSG_SIZE    = 4096
)

type User struct {
	ID       string
	Password string
	UserType pb.UserType
	HasVoted bool
	Conn     net.Conn
}

type Server struct {
	listener        net.Listener
	users           map[string]*User
	candidates      map[string]*pb.Candidate
	votes           map[string]int32 // candidateID -> vote count (simplified from full candidate for tally)
	votingDeadline  time.Time
	isVotingOpen    bool
	electionResults *pb.ElectionResultsPayload
	mu              sync.Mutex
}

func NewServer() *Server {
	return &Server{
		users: map[string]*User{
			"elector1": {ID: "elector1", Password: "password", UserType: pb.UserType_ELECTOR},
			"elector2": {ID: "elector2", Password: "password", UserType: pb.UserType_ELECTOR},
			"admin1":   {ID: "admin1", Password: "adminpass", UserType: pb.UserType_ADMIN},
		},
		candidates:   make(map[string]*pb.Candidate),
		votes:        make(map[string]int32),
		isVotingOpen: false,
	}
}

func (s *Server) Start() {
	var err error
	s.listener, err = net.Listen("tcp", TCP_PORT)
	if err != nil {
		log.Fatalf("Failed to start TCP server: %v", err)
	}
	defer s.listener.Close()
	log.Printf("Server listening on %s", TCP_PORT)

	s.startVotingPeriod(VOTING_DURATION)

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		log.Printf("Accepted connection from %s", conn.RemoteAddr())
		go s.handleConnection(conn)
	}
}

func (s *Server) startVotingPeriod(duration time.Duration) {
	s.mu.Lock()
	if s.isVotingOpen {
		s.mu.Unlock()
		log.Println("Voting is already open.")
		return
	}
	s.isVotingOpen = true
	s.votingDeadline = time.Now().Add(duration)
	s.votes = make(map[string]int32) // Reset votes for candidates
	for id := range s.candidates { // Reset vote counts in candidate objects too
		if cand, ok := s.candidates[id]; ok {
			cand.VoteCount = 0
			cand.Percentage = 0
		}
	}
	for _, u := range s.users { // Reset elector voted status
		if u.UserType == pb.UserType_ELECTOR {
			u.HasVoted = false
		}
	}
	s.electionResults = nil // Clear previous results
	log.Printf("Voting started. Deadline: %s", s.votingDeadline.Format(time.RFC3339))
	s.mu.Unlock()

	go func() {
		timer := time.NewTimer(time.Until(s.votingDeadline))
		<-timer.C
		s.endVotingAndCalculateResults()
	}()
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	var loggedInUser *User // To track which user is on this connection

	for {
		var msgLen uint32
		if err := binary.Read(conn, binary.BigEndian, &msgLen); err != nil {
			if err == io.EOF {
				log.Printf("Client %s disconnected", conn.RemoteAddr())
				if loggedInUser != nil {
					s.mu.Lock()
					loggedInUser.Conn = nil // Mark as disconnected
					s.mu.Unlock()
				}
				return
			}
			log.Printf("Error reading message length from %s: %v", conn.RemoteAddr(), err)
			return
		}

		if msgLen > MAX_MSG_SIZE {
			log.Printf("Message from %s too large: %d bytes. Closing connection.", conn.RemoteAddr(), msgLen)
			s.sendErrorResponse(conn, "Message too large.")
			return
		}

		msgBytes := make([]byte, msgLen)
		if _, err := io.ReadFull(conn, msgBytes); err != nil {
			log.Printf("Error reading message from %s: %v", conn.RemoteAddr(), err)
			return
		}

		req := &pb.GenericRequest{}
		if err := proto.Unmarshal(msgBytes, req); err != nil {
			log.Printf("Failed to unmarshal request from %s: %v", conn.RemoteAddr(), err)
			s.sendErrorResponse(conn, "Invalid request format")
			continue
		}

		log.Printf("Received %s request from %s", req.Type, conn.RemoteAddr())

		switch req.Type {
		case pb.GenericRequest_LOGIN:
			user := s.handleLogin(conn, req.Payload)
			if user != nil {
				loggedInUser = user // Associate user with this connection handler
				log.Printf("User %s (%s) logged in from %s", loggedInUser.ID, loggedInUser.UserType, conn.RemoteAddr())
			}
		case pb.GenericRequest_GET_CANDIDATES:
			if loggedInUser == nil {
				s.sendErrorResponse(conn, "Not logged in")
				continue
			}
			s.handleGetCandidates(conn)
		case pb.GenericRequest_SUBMIT_VOTE:
			if loggedInUser == nil || loggedInUser.UserType != pb.UserType_ELECTOR {
				s.sendErrorResponse(conn, "Only logged-in electors can vote")
				continue
			}
			s.handleSubmitVote(conn, req.Payload, loggedInUser) // Pass the loggedInUser
		case pb.GenericRequest_ADD_CANDIDATE:
			if loggedInUser == nil || loggedInUser.UserType != pb.UserType_ADMIN {
				s.sendErrorResponse(conn, "Only logged-in admins can add candidates")
				continue
			}
			s.handleAddCandidate(conn, req.Payload)
		case pb.GenericRequest_REMOVE_CANDIDATE:
			if loggedInUser == nil || loggedInUser.UserType != pb.UserType_ADMIN {
				s.sendErrorResponse(conn, "Only logged-in admins can remove candidates")
				continue
			}
			s.handleRemoveCandidate(conn, req.Payload)
		default:
			log.Printf("Unknown request type from %s: %v", conn.RemoteAddr(), req.Type)
			s.sendErrorResponse(conn, "Unknown request type")
		}
	}
}

func (s *Server) handleLogin(conn net.Conn, payload []byte) *User {
	s.mu.Lock()
	defer s.mu.Unlock()

	loginReq := &pb.LoginPayload{}
	if err := proto.Unmarshal(payload, loginReq); err != nil {
		s.sendErrorResponseLocked(conn, "Invalid login payload")
		return nil
	}

	user, exists := s.users[loginReq.UserId]
	if !exists || user.Password != loginReq.Password || user.UserType != loginReq.UserType {
		s.sendErrorResponseLocked(conn, "Invalid credentials or user type")
		return nil
	}

	if user.Conn != nil && user.Conn != conn { // Allow re-login on same conn, but not if active elsewhere
		s.sendErrorResponseLocked(conn, "User already logged in on another connection")
		return nil
	}
	user.Conn = conn

	respType := pb.GenericResponse_LOGIN_SUCCESS_ELECTOR
	var respPayloadData []byte
	var err error
	message := "Login successful."

	if user.UserType == pb.UserType_ELECTOR {
		if s.isVotingOpen {
			candidatesList := make([]*pb.Candidate, 0, len(s.candidates))
			for _, c := range s.candidates {
				candidatesList = append(candidatesList, &pb.Candidate{Id: c.Id, Name: c.Name})
			}
			clp := &pb.CandidateListPayload{
				Candidates:     candidatesList,
				VotingDeadline: s.votingDeadline.Format(time.RFC3339),
			}
			respPayloadData, err = proto.Marshal(clp)
			if err != nil {
				s.sendErrorResponseLocked(conn, "Failed to prepare candidate list")
				user.Conn = nil
				return nil
			}
		} else {
			message = "Login successful. Voting is not currently open."
			if s.electionResults != nil {
				message = "Login successful. Voting has ended."
				respType = pb.GenericResponse_ELECTION_RESULTS // Send results instead if available
				respPayloadData, _ = proto.Marshal(s.electionResults)
			}
		}
	} else if user.UserType == pb.UserType_ADMIN {
		respType = pb.GenericResponse_LOGIN_SUCCESS_ADMIN
	}

	s.sendProtoResponse(conn, respType, respPayloadData, message, true)
	return user
}

func (s *Server) handleGetCandidates(conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isVotingOpen {
		if s.electionResults != nil {
			resultsPayloadBytes, err := proto.Marshal(s.electionResults)
			if err != nil {
				s.sendErrorResponseLocked(conn, "Failed to serialize results")
				return
			}
			s.sendProtoResponse(conn, pb.GenericResponse_ELECTION_RESULTS, resultsPayloadBytes, "Voting has ended. Here are the results.", true)
		} else {
			s.sendProtoResponse(conn, pb.GenericResponse_GENERAL_STATUS, nil, "Voting is not currently open and no results available.", false)
		}
		return
	}

	candidatesList := make([]*pb.Candidate, 0, len(s.candidates))
	for _, c := range s.candidates {
		candidatesList = append(candidatesList, &pb.Candidate{Id: c.Id, Name: c.Name})
	}
	clp := &pb.CandidateListPayload{
		Candidates:     candidatesList,
		VotingDeadline: s.votingDeadline.Format(time.RFC3339),
	}
	payloadBytes, err := proto.Marshal(clp)
	if err != nil {
		s.sendErrorResponseLocked(conn, "Failed to prepare candidate list")
		return
	}
	s.sendProtoResponse(conn, pb.GenericResponse_CANDIDATE_LIST, payloadBytes, "Current candidates and deadline", true)
}

func (s *Server) handleSubmitVote(conn net.Conn, payload []byte, elector *User) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isVotingOpen {
		s.sendErrorResponseLocked(conn, "Voting is closed.")
		return
	}
	if elector.HasVoted {
		s.sendErrorResponseLocked(conn, "You have already voted.")
		return
	}

	voteReq := &pb.SubmitVotePayload{}
	if err := proto.Unmarshal(payload, voteReq); err != nil {
		s.sendErrorResponseLocked(conn, "Invalid vote payload.")
		return
	}
	if voteReq.ElectorId != elector.ID { // Sanity check
		s.sendErrorResponseLocked(conn, "Vote payload elector ID mismatch.")
		return
	}

	candidate, exists := s.candidates[voteReq.CandidateId]
	if !exists {
		s.sendErrorResponseLocked(conn, "Invalid candidate ID.")
		return
	}

	candidate.VoteCount++      // This is a pointer, updates the map's value.
	s.votes[candidate.Id]++ // Also update the specific tally map.
	elector.HasVoted = true

	log.Printf("Elector %s voted for %s (%s)", elector.ID, candidate.Name, candidate.Id)
	s.sendProtoResponse(conn, pb.GenericResponse_VOTE_ACK, nil, "Vote successfully recorded.", true)
}

func (s *Server) handleAddCandidate(conn net.Conn, payload []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isVotingOpen { // Stricter: disallow if voting has ever started for this session
		s.sendErrorResponseLocked(conn, "Cannot add candidates while voting is open or has concluded.")
		return
	}

	addReq := &pb.AddCandidatePayload{}
	if err := proto.Unmarshal(payload, addReq); err != nil {
		s.sendErrorResponseLocked(conn, "Invalid add candidate payload.")
		return
	}
	if addReq.Candidate == nil || addReq.Candidate.Id == "" || addReq.Candidate.Name == "" {
		s.sendErrorResponseLocked(conn, "Candidate ID and Name cannot be empty.")
		return
	}
	if _, exists := s.candidates[addReq.Candidate.Id]; exists {
		s.sendErrorResponseLocked(conn, "Candidate ID already exists.")
		return
	}

	newCand := &pb.Candidate{
		Id:        addReq.Candidate.Id,
		Name:      addReq.Candidate.Name,
		VoteCount: 0,
	}
	s.candidates[newCand.Id] = newCand
	s.votes[newCand.Id] = 0 // Ensure it's in the tally map

	log.Printf("Admin added candidate: %s (%s)", newCand.Name, newCand.Id)
	s.sendProtoResponse(conn, pb.GenericResponse_ADMIN_ACTION_ACK, nil, "Candidate added successfully.", true)
}

func (s *Server) handleRemoveCandidate(conn net.Conn, payload []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isVotingOpen {
		s.sendErrorResponseLocked(conn, "Cannot remove candidates while voting is open or has concluded.")
		return
	}

	removeReq := &pb.RemoveCandidatePayload{}
	if err := proto.Unmarshal(payload, removeReq); err != nil {
		s.sendErrorResponseLocked(conn, "Invalid remove candidate payload.")
		return
	}
	if _, exists := s.candidates[removeReq.CandidateId]; !exists {
		s.sendErrorResponseLocked(conn, "Candidate ID not found.")
		return
	}

	delete(s.candidates, removeReq.CandidateId)
	delete(s.votes, removeReq.CandidateId)

	log.Printf("Admin removed candidate ID: %s", removeReq.CandidateId)
	s.sendProtoResponse(conn, pb.GenericResponse_ADMIN_ACTION_ACK, nil, "Candidate removed successfully.", true)
}

func (s *Server) endVotingAndCalculateResults() {
	s.mu.Lock()
	// Check if already ended by another path or called multiple times
	if !s.isVotingOpen && s.electionResults != nil {
		s.mu.Unlock()
		return
	}
	s.isVotingOpen = false // Ensure voting is marked closed
	log.Println("Voting has officially ended. Calculating results...")

	totalVotes := int32(0)
	// Use s.votes for final tally as s.candidates[id].VoteCount might not be consistently updated if logic error
	for _, count := range s.votes {
		totalVotes += count
	}
	
	// Update candidate objects with final counts from s.votes
	for id, cand := range s.candidates {
		cand.VoteCount = s.votes[id] // Ensure candidate object has the correct final count
	}

	results := &pb.ElectionResultsPayload{
		TotalVotes:       totalVotes,
		CandidateResults: make([]*pb.Candidate, 0, len(s.candidates)),
		StatusMessage:    "Voting has ended. Final Results:",
	}

	maxVotes := int32(-1)
	var winner *pb.Candidate = nil // Initialize winner as nil

	for _, cand := range s.candidates { // Iterate over the s.candidates map which has full Candidate objects
		percentage := 0.0
		if totalVotes > 0 {
			percentage = (float64(cand.VoteCount) / float64(totalVotes)) * 100.0
		}
		cand.Percentage = percentage // Update percentage in the server's candidate map instance

		resultCand := &pb.Candidate{ // Create a copy for the results payload
			Id:         cand.Id,
			Name:       cand.Name,
			VoteCount:  cand.VoteCount,
			Percentage: percentage,
		}
		results.CandidateResults = append(results.CandidateResults, resultCand)

		if cand.VoteCount > 0 && cand.VoteCount > maxVotes {
			maxVotes = cand.VoteCount
			winner = resultCand
		} else if cand.VoteCount > 0 && cand.VoteCount == maxVotes {
			if winner != nil { // Tie
				winner.Name += " (tie)" // Simplistic tie indication
			} else { // First candidate with votes if all were 0 before
				winner = resultCand
			}
		}
	}
	
	if totalVotes == 0 {
		results.StatusMessage = "Voting has ended. No votes were cast."
		results.Winner = &pb.Candidate{Name: "No winner (no votes)"}
	} else if winner == nil && totalVotes > 0 { // Votes exist, but no single winner (e.g. all candidates got 0 votes, but totalVotes > 0 means error in logic)
		results.Winner = &pb.Candidate{Name: "No clear winner"} // Or specific tie logic
	} else {
		results.Winner = winner
	}

	s.electionResults = results
	s.mu.Unlock() // Unlock before logging or broadcasting

	log.Printf("Results Calculated: Total Votes: %d", results.TotalVotes)
	if results.Winner != nil {
		log.Printf("Winner: %s with %d votes (%.2f%%)", results.Winner.Name, results.Winner.VoteCount, results.Winner.Percentage)
	} else {
		log.Println("No winner determined or no votes cast.")
	}
	// Optionally, broadcast results to all connected clients.
}

func sendProtoMessage(conn net.Conn, msg proto.Message) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	if err := binary.Write(conn, binary.BigEndian, uint32(len(data))); err != nil {
		return fmt.Errorf("failed to write message length: %w", err)
	}
	if _, err := conn.Write(data); err != nil {
		return fmt.Errorf("failed to write message data: %w", err)
	}
	return nil
}

func (s *Server) sendProtoResponse(conn net.Conn, respType pb.GenericResponse_Type, payloadData []byte, message string, success bool) {
	resp := &pb.GenericResponse{
		Type:    respType,
		Payload: payloadData,
		Message: message,
		Success: success,
	}
	if err := sendProtoMessage(conn, resp); err != nil {
		log.Printf("Error sending %s response to %s: %v", respType, conn.RemoteAddr(), err)
	}
}

func (s *Server) sendErrorResponse(conn net.Conn, message string) {
	s.sendProtoResponse(conn, pb.GenericResponse_GENERAL_STATUS, nil, message, false)
}

func (s *Server) sendErrorResponseLocked(conn net.Conn, message string) { // For use when s.mu is already locked
	s.sendProtoResponse(conn, pb.GenericResponse_GENERAL_STATUS, nil, message, false)
}

func main() {
	server := NewServer()
	// Example: Pre-add some candidates before starting
	server.mu.Lock()
	server.candidates["c1"] = &pb.Candidate{Id: "c1", Name: "Candidate Alpha", VoteCount: 0}
	server.candidates["c2"] = &pb.Candidate{Id: "c2", Name: "Candidate Beta", VoteCount: 0}
	server.votes["c1"] = 0
	server.votes["c2"] = 0
	server.mu.Unlock()

	server.Start()
}