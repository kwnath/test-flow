const fs = require('fs');
const path = require('path');

const DATA_FILE = path.join(__dirname, 'data', 'todos.json');

// Initialize data file if it doesn't exist
function initializeDataFile() {
  const dir = path.dirname(DATA_FILE);
  if (!fs.existsSync(dir)) {
    fs.mkdirSync(dir, { recursive: true });
  }
  if (!fs.existsSync(DATA_FILE)) {
    fs.writeFileSync(DATA_FILE, JSON.stringify([], null, 2));
  }
}

// Read todos from file
function readTodos() {
  initializeDataFile();
  const data = fs.readFileSync(DATA_FILE, 'utf8');
  return JSON.parse(data);
}

// Write todos to file atomically
function writeTodos(todos) {
  const tempFile = DATA_FILE + '.tmp';
  fs.writeFileSync(tempFile, JSON.stringify(todos, null, 2));
  fs.renameSync(tempFile, DATA_FILE);
}

// Get next available ID
function getNextId(todos) {
  if (todos.length === 0) return 1;
  return Math.max(...todos.map(t => t.id)) + 1;
}

// Get all todos
function getTodos() {
  return readTodos();
}

// Add a new todo
function addTodo(text) {
  if (!text || typeof text !== 'string' || text.trim().length === 0) {
    throw new Error('Todo text is required');
  }

  if (text.length > 500) {
    throw new Error('Todo text must be 500 characters or less');
  }

  const todos = readTodos();
  const newTodo = {
    id: getNextId(todos),
    text: text.trim(),
    completed: false,
    createdAt: new Date().toISOString()
  };

  todos.push(newTodo);
  writeTodos(todos);
  return newTodo;
}

// Toggle todo completion status
function toggleComplete(id) {
  const todos = readTodos();
  const todo = todos.find(t => t.id === parseInt(id));

  if (!todo) {
    throw new Error('Todo not found');
  }

  todo.completed = !todo.completed;
  writeTodos(todos);
  return todo;
}

// Delete a todo
function deleteTodo(id) {
  const todos = readTodos();
  const index = todos.findIndex(t => t.id === parseInt(id));

  if (index === -1) {
    throw new Error('Todo not found');
  }

  todos.splice(index, 1);
  writeTodos(todos);
  return true;
}

module.exports = {
  getTodos,
  addTodo,
  toggleComplete,
  deleteTodo
};
