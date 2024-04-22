-- Create database
CREATE DATABASE task_manager;

-- Use the created database
USE task_manager;

-- Create tasks table
CREATE TABLE tasks (
    id INT AUTO_INCREMENT PRIMARY KEY,
    description VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL
);