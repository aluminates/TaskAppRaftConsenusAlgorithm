package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Define task structure
type Task struct {
	ID          int
	Description string
	Status      string
}

// Define RaftNode structure
type RaftNode struct {
	id          int
	term        int
	isLeader    bool
	voteGranted bool
	votes       int
	mu          sync.Mutex
	inbox       chan Message
	status      string  // "alive" or "failed"
	db          *sql.DB // MySQL database connection
}

// Message structure for communication between nodes
type Message struct {
	MessageType string
	Args        interface{}
	SenderID    int
}

// RequestVoteArgs structure for RequestVote RPC method
type RequestVoteArgs struct {
	Term        int
	CandidateID int
}

// RequestVoteReply structure for RequestVote RPC method reply
type RequestVoteReply struct {
	Term        int
	VoteGranted bool
}

// Initialize MySQL database connection
func InitDB() *sql.DB {
	db, err := sql.Open("mysql", "root:Riya2002@tcp(127.0.0.1:3306)/task_manager")
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}

	// Check if the database connection is successful
	if err := db.Ping(); err != nil {
		log.Fatal("Error pinging database:", err)
	}

	// Print a message indicating successful database connection
	fmt.Println("Connected to MySQL database!")
	return db
}

// CRUD operations for tasks
func CreateTask(db *sql.DB, task Task) error {
	_, err := db.Exec("INSERT INTO tasks (description, status) VALUES (?, ?)", task.Description, task.Status)
	return err
}

func ReadTask(db *sql.DB, id int) (Task, error) {
	var task Task
	row := db.QueryRow("SELECT id, description, status FROM tasks WHERE id = ?", id)
	err := row.Scan(&task.ID, &task.Description, &task.Status)
	if err != nil {
		return Task{}, err
	}
	return task, nil
}

//func UpdateTask(db *sql.DB, task Task) error {
//	_, err := db.Exec("UPDATE tasks SET description = ?, status = ? WHERE id = ?", task.Description, task.Status, task.ID)
//	return err
//}

func UpdateTask(db *sql.DB, task Task) error {
	// Execute UpdateTask operation on the MySQL database
	_, err := db.Exec("UPDATE tasks SET description = ?, status = ? WHERE id = ?", task.Description, task.Status, task.ID)
	if err != nil {
		log.Printf("Error updating task with ID %d: %v", task.ID, err)
	} else {
		log.Printf("Task with ID %d updated successfully", task.ID)
	}
	return err
}

func DeleteTask(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM tasks WHERE id = ?", id)
	return err
}

// API handlers
func handleTasks(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Fetch all tasks
		tasks, err := getAllTasks(db)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(tasks)
	case http.MethodPost:
		// Create a new task
		var task Task
		err := json.NewDecoder(r.Body).Decode(&task)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		err = CreateTask(db, task)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleTaskByID(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/api/tasks/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// Fetch a task by ID
		task, err := ReadTask(db, id)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(task)
	case http.MethodDelete:
		// Delete a task by ID
		err := DeleteTask(db, id)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getAllTasks(db *sql.DB) ([]Task, error) {
	rows, err := db.Query("SELECT id, description, status FROM tasks")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Description, &task.Status); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// RequestVote RPC method
func (n *RaftNode) RequestVote(args RequestVoteArgs, reply *RequestVoteReply) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if args.Term > n.term {
		n.term = args.Term
		n.voteGranted = true
		n.isLeader = false
	}

	reply.Term = n.term
	reply.VoteGranted = n.voteGranted
}

// ElectLeader method to initiate leader election
func (n *RaftNode) ElectLeader() {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.status == "failed" {
		return // If node is failed, do not initiate leader election
	}

	n.term++
	n.voteGranted = true
	n.votes = 1

	args := RequestVoteArgs{
		Term:        n.term,
		CandidateID: n.id,
	}

	for i := 0; i < 3; i++ {
		if i != n.id && nodes[i].status != "failed" {
			msg := Message{
				MessageType: "RequestVote",
				Args:        args,
				SenderID:    n.id,
			}
			fmt.Printf("Node %d sending vote request to node %d\n", n.id, i)
			nodes[i].inbox <- msg
		}
	}

	// Simulate the time it takes to receive votes
	time.Sleep(200 * time.Millisecond)

	fmt.Printf("Node %d received %d votes, status: %s\n", n.id, n.votes, n.status)

	if n.votes <= 2 && n.isLeader && n.status == "alive" {
		fmt.Printf("Leader (%d) failed.\n", n.id)
		n.status = "failed"
		n.mu.Unlock() // Unlock before triggering re-election
		var wg sync.WaitGroup
		wg.Add(1)
		go n.reElection(&wg) // Trigger re-election
		wg.Wait()            // Wait for re-election to complete
		return
	}

	if n.votes > 2 {
		n.isLeader = true
		fmt.Printf("Node %d becomes the leader.\n", n.id)
	}
}

// handleMessage method to handle incoming messages
func (n *RaftNode) handleMessage() {
	for msg := range n.inbox {
		switch msg.MessageType {
		case "RequestVote":
			args := msg.Args.(RequestVoteArgs)
			var reply RequestVoteReply
			n.RequestVote(args, &reply)
			nodes[msg.SenderID].inbox <- Message{MessageType: "RequestVoteReply", Args: reply, SenderID: n.id}
		case "RequestVoteReply":
			// Handle RequestVoteReply message if needed
		case "CreateTask":
			task := msg.Args.(Task)
			// Execute CreateTask operation on the MySQL database
			err := CreateTask(n.db, task)
			// Replicate the operation to other nodes if necessary
			if err == nil && n.isLeader {
				for _, node := range nodes {
					if node != n {
						node.inbox <- msg
					}
				}
			}
		case "ReadTask":
			id := msg.Args.(int)
			// Execute ReadTask operation on the MySQL database
			task, err := ReadTask(n.db, id)
			// Send the task back to the requester
			if err == nil {
				msg := Message{
					MessageType: "ReadTaskReply",
					Args:        task,
					SenderID:    n.id,
				}
				nodes[msg.SenderID].inbox <- msg
			}
		case "UpdateTask":
			task := msg.Args.(Task)
			// Execute UpdateTask operation on the MySQL database
			err := UpdateTask(n.db, task)
			// Replicate the operation to other nodes if necessary
			if err == nil && n.isLeader {
				for _, node := range nodes {
					if node != n {
						node.inbox <- msg
					}
				}
			}
		case "DeleteTask":
			id := msg.Args.(int)
			// Execute DeleteTask operation on the MySQL database
			err := DeleteTask(n.db, id)
			// Replicate the operation to other nodes if necessary
			if err == nil && n.isLeader {
				for _, node := range nodes {
					if node != n {
						node.inbox <- msg
					}
				}
			}
		}
	}
}

// reElection method to trigger re-election
func (n *RaftNode) reElection(wg *sync.WaitGroup) {
	defer wg.Done()
	time.Sleep(500 * time.Millisecond) // Simulate time before re-election

	// Always initiate leader election
	fmt.Printf("Re-electing leader from node %d...\n", n.id)
	for _, node := range nodes {
		if node.id != n.id {
			var wg sync.WaitGroup
			wg.Add(1)
			go node.ElectLeader()
			wg.Wait()
			break // only one node should initiate the leader election
		}
	}
}

// Function to get the current leader
func getRaftLeader() *RaftNode {
	for _, node := range nodes {
		if node.isLeader {
			return node
		}
	}
	return nil // No leader found
}

var nodes []*RaftNode

func addTasksToDatabase(db *sql.DB) {
	// Create task instances
	tasks := []Task{
		{Description: "Task 11", Status: "Pending"},
		{Description: "Task 21", Status: "Completed"},
		// Add more tasks as needed
	}

	// Insert each task into the database
	for _, task := range tasks {
		err := CreateTask(db, task)
		if err != nil {
			log.Printf("Error adding task to database: %v", err)
		} else {
			log.Printf("Task added to database successfully: %+v", task)
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	db := InitDB()
	defer db.Close()

	// Initialize Raft nodes
	nodes = make([]*RaftNode, 4)
	for i := range nodes {
		nodes[i] = &RaftNode{
			id:     i,
			inbox:  make(chan Message),
			status: "alive",
			db:     db,
		}
		go nodes[i].handleMessage()
	}

	// Simulate node failure
	time.Sleep(300 * time.Millisecond)
	nodes[0].status = "failed"

	// Randomly select a node from the remaining nodes for leader election
	for _, node := range nodes[1:] {
		if node.status != "failed" {
			fmt.Printf("Initiating leader election from node %d\n", node.id)
			node.ElectLeader()
			break
		}
	}
	time.Sleep(3000 * time.Millisecond)
	nodes[1].status = "failed"

	// reElection method to trigger re-election
	time.Sleep(500 * time.Millisecond) // Simulate time before re-election
	fmt.Printf("Re-electing leader...")
	for _, node := range nodes[2:] {
		if node.status != "failed" {
			fmt.Printf("Initiating leader election from node %d\n", node.id)
			node.ElectLeader()
			break
		}
	}

	// Serve frontend files
	http.Handle("/", http.FileServer(http.Dir("../frontend")))

	// API endpoints
	http.HandleFunc("/api/tasks", func(w http.ResponseWriter, r *http.Request) {
		handleTasks(db, w, r)
	})
	http.HandleFunc("/api/tasks/", func(w http.ResponseWriter, r *http.Request) {
		handleTaskByID(db, w, r)
	})

	// Start HTTP server
	fmt.Println("Server is running on port 8080")
	http.ListenAndServe(":8080", nil)
}
