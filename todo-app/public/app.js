const API_BASE = '/api/todos';

const todoForm = document.getElementById('todo-form');
const todoInput = document.getElementById('todo-input');
const todosContainer = document.getElementById('todos');
const emptyState = document.getElementById('empty-state');

// Fetch and display all todos
async function loadTodos() {
  try {
    const response = await fetch(API_BASE);
    if (!response.ok) throw new Error('Failed to load todos');

    const todos = await response.json();
    renderTodos(todos);
  } catch (error) {
    console.error('Error loading todos:', error);
    alert('Failed to load todos. Please refresh the page.');
  }
}

// Render todos to the DOM
function renderTodos(todos) {
  todosContainer.innerHTML = '';

  if (todos.length === 0) {
    emptyState.classList.remove('hidden');
    return;
  }

  emptyState.classList.add('hidden');

  todos.forEach(todo => {
    const li = createTodoElement(todo);
    todosContainer.appendChild(li);
  });
}

// Create a todo DOM element
function createTodoElement(todo) {
  const li = document.createElement('li');
  li.className = `todo-item ${todo.completed ? 'completed' : ''}`;
  li.dataset.id = todo.id;

  const checkbox = document.createElement('input');
  checkbox.type = 'checkbox';
  checkbox.className = 'todo-checkbox';
  checkbox.checked = todo.completed;
  checkbox.addEventListener('change', () => toggleTodo(todo.id));

  const text = document.createElement('span');
  text.className = 'todo-text';
  text.textContent = escapeHtml(todo.text);

  const deleteBtn = document.createElement('button');
  deleteBtn.className = 'delete-btn';
  deleteBtn.textContent = 'Delete';
  deleteBtn.addEventListener('click', () => deleteTodo(todo.id));

  li.appendChild(checkbox);
  li.appendChild(text);
  li.appendChild(deleteBtn);

  return li;
}

// Add a new todo
async function addTodo(text) {
  try {
    const response = await fetch(API_BASE, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ text })
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Failed to add todo');
    }

    await loadTodos();
    todoInput.value = '';
  } catch (error) {
    console.error('Error adding todo:', error);
    alert(error.message);
  }
}

// Toggle todo completion
async function toggleTodo(id) {
  try {
    const response = await fetch(`${API_BASE}/${id}/complete`, {
      method: 'PATCH'
    });

    if (!response.ok) throw new Error('Failed to update todo');

    await loadTodos();
  } catch (error) {
    console.error('Error toggling todo:', error);
    alert('Failed to update todo. Please try again.');
  }
}

// Delete a todo
async function deleteTodo(id) {
  try {
    const response = await fetch(`${API_BASE}/${id}`, {
      method: 'DELETE'
    });

    if (!response.ok) throw new Error('Failed to delete todo');

    await loadTodos();
  } catch (error) {
    console.error('Error deleting todo:', error);
    alert('Failed to delete todo. Please try again.');
  }
}

// Escape HTML to prevent XSS
function escapeHtml(text) {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

// Form submit handler
todoForm.addEventListener('submit', (e) => {
  e.preventDefault();
  const text = todoInput.value.trim();

  if (!text) {
    alert('Please enter a todo');
    return;
  }

  addTodo(text);
});

// Load todos on page load
loadTodos();
