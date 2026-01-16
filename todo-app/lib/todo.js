/**
 * Todo business logic
 * Provides CRUD operations for managing todos
 */

/**
 * Add a new todo
 * @param {Array} todos - Array of existing todos
 * @param {number} nextId - Next available ID
 * @param {string} text - Todo description
 * @returns {{success: boolean, message: string, todo?: object, nextId?: number}}
 */
function addTodo(todos, nextId, text) {
  if (!text || text.trim() === '') {
    return { success: false, message: 'Error: Todo text cannot be empty' };
  }

  const todo = {
    id: nextId,
    text: text.trim(),
    completed: false,
    createdAt: Date.now()
  };

  todos.push(todo);

  return {
    success: true,
    message: `Added todo #${nextId}: "${todo.text}"`,
    todo,
    nextId: nextId + 1
  };
}

/**
 * List all todos
 * @param {Array} todos - Array of todos
 * @returns {string} Formatted list of todos
 */
function listTodos(todos) {
  if (todos.length === 0) {
    return 'No todos yet! Add one with: node index.js add "Your todo"';
  }

  let output = '\nYour Todos:\n';
  output += '─'.repeat(50) + '\n';

  todos.forEach(todo => {
    const status = todo.completed ? '✓' : ' ';
    const textStyle = todo.completed ? `(done) ${todo.text}` : todo.text;
    output += `[${status}] #${todo.id} - ${textStyle}\n`;
  });

  output += '─'.repeat(50);
  return output;
}

/**
 * Mark a todo as completed
 * @param {Array} todos - Array of todos
 * @param {number} id - Todo ID to mark as done
 * @returns {{success: boolean, message: string}}
 */
function markDone(todos, id) {
  const todo = todos.find(t => t.id === id);

  if (!todo) {
    return { success: false, message: `Error: Todo #${id} not found` };
  }

  if (todo.completed) {
    return { success: true, message: `Todo #${id} is already completed` };
  }

  todo.completed = true;
  return { success: true, message: `Marked todo #${id} as completed: "${todo.text}"` };
}

/**
 * Remove a todo
 * @param {Array} todos - Array of todos
 * @param {number} id - Todo ID to remove
 * @returns {{success: boolean, message: string, index?: number}}
 */
function removeTodo(todos, id) {
  const index = todos.findIndex(t => t.id === id);

  if (index === -1) {
    return { success: false, message: `Error: Todo #${id} not found` };
  }

  const todo = todos[index];
  todos.splice(index, 1);

  return {
    success: true,
    message: `Removed todo #${id}: "${todo.text}"`,
    index
  };
}

module.exports = {
  addTodo,
  listTodos,
  markDone,
  removeTodo
};
