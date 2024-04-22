import sys
import subprocess

# Use subprocess to get the installation path of streamlit
streamlit_path = subprocess.check_output(['pip', 'show', 'streamlit']).decode('utf-8')
streamlit_path = streamlit_path.split('\n')[8].split(': ')[1]

# Add the installation path of streamlit to the Python path
sys.path.append(streamlit_path)

import streamlit as st # type: ignore
import requests # type: ignore

# Define API base URL
API_URL = "http://localhost:8080/api/tasks"

# Custom CSS styles
custom_css = """
<style>
    body {
        background: linear-gradient(to right, #ff5e57, #0072ff);
        animation: gradient 15s ease infinite;
    }

    @keyframes gradient {
        0% {
            background-position: 0% 50%;
        }
        50% {
            background-position: 100% 50%;
        }
        100% {
            background-position: 0% 50%;
        }
    }

    h1, h2 {
        color: purple;
        font-weight: bold;
        text-align: center;
        text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.5);
        animation: fadeIn 2s ease;
    }

    .stButton {
        font-size: 16px;
        font-weight: bold;
        color: purple;
        background-color: #2e4a66;
        border-radius: 5px;
        box-shadow: 2px 2px 4px rgba(0, 0, 0, 0.5);
        transition: all 0.3s ease;
        animation: fadeIn 2s ease;
    }

    .stButton:hover {
        background-color: #4a708b;
        transform: scale(1.05);
    }

    .stTextInput, .stSelectbox {
        font-size: 16px;
        border-radius: 5px;
        box-shadow: 2px 2px 4px rgba(0, 0, 0, 0.5);
        animation: fadeIn 2s ease;
    }

    .task-container {
        background-color: rgba(255, 255, 255, 0.8);
        border-radius: 10px;
        padding: 20px;
        margin-bottom: 20px;
        box-shadow: 2px 2px 4px rgba(0, 0, 0, 0.5);
        animation: fadeIn 2s ease;
    }

    @keyframes fadeIn {
        0% {
            opacity: 0;
        }
        100% {
            opacity: 1;
        }
    }
</style>

"""
st.markdown(custom_css, unsafe_allow_html=True)

def add_task(description, status):
    data = {"description": description, "status": status}
    response = requests.post(API_URL, json=data)
    if response.status_code == 200:
        return True
    else:
        return False

def get_tasks():
    response = requests.get(API_URL)
    if response.status_code == 200:
        return response.json()
    else:
        return []

def update_task(id, description, status):
    data = {"id": id, "description": description, "status": status}
    response = requests.put(API_URL + f"/{id}", json=data)
    if response.status_code == 200:
        return True
    else:
        st.error(f"Failed to update task {id}! Error: {response.text}")
        return False

def delete_task(id):
    response = requests.delete(API_URL + f"/{id}")
    if response.status_code == 200:
        return True
    else:
        return False

# Streamlit UI
st.title("Task Manager")

# Add Task Section
st.header("Add Task")
description = st.text_input("Description", placeholder="Enter task description")
status = st.selectbox("Status", ["Pending", "Completed"])
if st.button("Add Task", use_container_width=True):
    if add_task(description, status):
        st.success("Task added successfully!")
    else:
        st.error("Failed to add task!")

# View Tasks Section
st.header("Tasks")
tasks = get_tasks()
if tasks:
    for task in tasks:
        task_id = task['ID']
        with st.container():
            st.markdown(f"**Task**")
            cols = st.columns(3)
            with cols[0]:
                task_description = st.text_input(f"Description", value=task['Description'])
            with cols[1]:
                task_status = st.selectbox(f"Status {task_description}", ["Pending", "Completed"], index=["Pending", "Completed"].index(task['Status']))
            with cols[2]:
                update_button = st.button(f"Update", key=f"update_{task_id}", use_container_width=True)
                delete_button = st.button(f"Delete", key=f"delete_{task_id}", use_container_width=True)



            if update_button:
                if update_task(task_id, task_description, task_status):
                    st.success(f"Task {task_id} updated successfully!")
                else:
                    st.error(f"Failed to update task {task_id}!")
            if delete_button:
                if delete_task(task_id):
                    st.success(f"Task {task_id} deleted successfully!")
                else:
                    st.error(f"Failed to delete task {task_id}!")
else:
    st.warning("No tasks available.")
