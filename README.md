# A Task Management Application using Raft Consensus Algorithm and MySQL

The objective of this project is to develop a task management application on distributed systems that utilises the Raft Consensus Algorithm to ensure consistency and fault tolerance across multiple nodes. MySQL is employed as the backend database to store task data. Implemented using a Streamlit UI, the application enables users to create, update, delete, and manage tasks across the distributed system.


## Working: 

1) API_URL: This variable holds the base URL of the API endpoint where tasks will be stored. It's set to "http://localhost:8080/api/tasks" by default.

2) Function Definitions:
   - `add_task`: Sends a POST request to the API to add a new task.
   - `get_tasks`: Sends a GET request to the API to retrieve all tasks.
   - `update_task`: Sends a PUT request to the API to update an existing task.
   - `delete_task`: Sends a DELETE request to the API to delete a task.

3) HTTP handlers are defined to handle API requests related to tasks, i.e., `handleTasks` handles requests to retrieve all tasks or create a new task, while `handleTaskByID` handles requests to retrieve or delete a task by its ID.

4) A `RaftNode` structure is defined to represent each node in the Raft-based distributed system. Each node has an inbox channel for receiving messages from other nodes. `RequestVote`, `ElectLeader`, `handleMessage`, and `reElection` methods are defined to handle Raft-specific functionalities like leader election, message handling, and re-election. In the `main` function, Raft nodes are initialized, and each node starts a goroutine to handle incoming messages (`handleMessage` method). After initializing nodes, a leader election is initiated by randomly selecting a node and calling its `ElectLeader` method. Node failure is simulated by changing the status of a node to "failed". Upon failure detection, the system triggers a new leader election process (`reElection` method).



