const express = require('express');
const path = require('path');
const todoService = require('./todoService');

const app = express();
const PORT = process.env.PORT || 3000;

// Middleware
app.use(express.json());
app.use(express.static(path.join(__dirname, '../public')));

// API Routes

// GET /api/todos - Get all todos
app.get('/api/todos', (req, res) => {
  try {
    const todos = todoService.getTodos();
    res.json(todos);
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

// POST /api/todos - Create a new todo
app.post('/api/todos', (req, res) => {
  try {
    const { text } = req.body;
    const newTodo = todoService.addTodo(text);
    res.status(201).json(newTodo);
  } catch (error) {
    res.status(400).json({ error: error.message });
  }
});

// PATCH /api/todos/:id/complete - Toggle todo completion
app.patch('/api/todos/:id/complete', (req, res) => {
  try {
    const updatedTodo = todoService.toggleComplete(req.params.id);
    res.json(updatedTodo);
  } catch (error) {
    if (error.message === 'Todo not found') {
      res.status(404).json({ error: error.message });
    } else {
      res.status(500).json({ error: error.message });
    }
  }
});

// DELETE /api/todos/:id - Delete a todo
app.delete('/api/todos/:id', (req, res) => {
  try {
    todoService.deleteTodo(req.params.id);
    res.status(204).send();
  } catch (error) {
    if (error.message === 'Todo not found') {
      res.status(404).json({ error: error.message });
    } else {
      res.status(500).json({ error: error.message });
    }
  }
});

// Serve frontend
app.get('/', (req, res) => {
  res.sendFile(path.join(__dirname, '../public/index.html'));
});

// Start server
app.listen(PORT, () => {
  console.log(`Todo app server running on http://localhost:${PORT}`);
});
