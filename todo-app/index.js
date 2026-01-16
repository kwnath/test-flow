#!/usr/bin/env node

/**
 * CLI Todo App
 * A simple command-line todo application with in-memory storage
 */

const { addTodo, listTodos, markDone, removeTodo } = require('./lib/todo');

// In-memory storage
let todos = [];
let nextId = 1;

/**
 * Display help information
 */
function showHelp() {
  console.log(`
Todo App - Simple CLI Todo Manager

Usage:
  node index.js <command> [arguments]

Commands:
  add <text>     Add a new todo
  list           List all todos
  done <id>      Mark a todo as completed
  rm <id>        Remove a todo
  help           Show this help message

Examples:
  node index.js add "Buy groceries"
  node index.js list
  node index.js done 1
  node index.js rm 2
`);
}

/**
 * Main CLI handler
 */
function main() {
  const args = process.argv.slice(2);
  const command = args[0];

  if (!command || command === 'help') {
    showHelp();
    return;
  }

  switch (command) {
    case 'add': {
      const text = args.slice(1).join(' ');
      const result = addTodo(todos, nextId, text);

      if (result.success) {
        nextId = result.nextId;
      }

      console.log(result.message);
      break;
    }

    case 'list': {
      const output = listTodos(todos);
      console.log(output);
      break;
    }

    case 'done': {
      const id = parseInt(args[1], 10);

      if (isNaN(id)) {
        console.log('Error: Please provide a valid todo ID');
        console.log('Usage: node index.js done <id>');
        break;
      }

      const result = markDone(todos, id);
      console.log(result.message);
      break;
    }

    case 'rm': {
      const id = parseInt(args[1], 10);

      if (isNaN(id)) {
        console.log('Error: Please provide a valid todo ID');
        console.log('Usage: node index.js rm <id>');
        break;
      }

      const result = removeTodo(todos, id);
      console.log(result.message);
      break;
    }

    default:
      console.log(`Error: Unknown command "${command}"`);
      console.log('Run "node index.js help" for usage information');
  }
}

// Run the CLI
main();
