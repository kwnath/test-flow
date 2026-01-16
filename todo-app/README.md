# Todo App

A simple command-line todo application with in-memory storage.

## Features

- Add todos with descriptions
- List all todos with completion status
- Mark todos as completed
- Remove todos
- Clean, user-friendly CLI interface

## Requirements

- Node.js >= 14.0.0

## Installation

No installation required! Just navigate to the `todo-app` directory.

## Usage

### Add a todo

```bash
node index.js add "Buy groceries"
node index.js add "Walk the dog"
```

### List all todos

```bash
node index.js list
```

Output:
```
Your Todos:
──────────────────────────────────────────────────
[ ] #1 - Buy groceries
[ ] #2 - Walk the dog
──────────────────────────────────────────────────
```

### Mark a todo as completed

```bash
node index.js done 1
```

### Remove a todo

```bash
node index.js rm 2
```

### Show help

```bash
node index.js help
```

## Storage

This app uses **in-memory storage**, which means:
- Todos are stored in RAM during the process lifetime
- Data is cleared when the Node.js process exits
- Perfect for quick, session-based task management

## Architecture

```
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│   CLI Layer  │───▶│  Todo Logic  │───▶│  In-Memory   │
│  (commands)  │◀───│  (business)  │◀───│   Storage    │
└──────────────┘    └──────────────┘    └──────────────┘
```

## Files

- `index.js` - CLI entry point and in-memory storage
- `lib/todo.js` - Business logic for todo operations
- `package.json` - Project metadata

## License

MIT
