# Terminal Pet

A virtual pet (Tamagotchi-like) terminal game that runs via SSH.

## Features

- SSH-based virtual pet game
- Cute ASCII animations that change based on your pet's mood
- Pet stats: hunger, happiness, health
- Interact with your pet: feed, play, and more
- Persistent pet state (saved to SQLite database)

## Installation

1. Clone the repository
2. Build the application
   ```
   go build -o bin/pet-game ./cmd
   ```
3. Run the server
   ```
   ./bin/pet-game
   ```

## Configuration

The application can be configured using environment variables:

| Environment Variable | Default | Description |
|----------------------|---------|-------------|
| `SSH_LISTEN_ADDR` | `0.0.0.0:23234` | The address and port to listen for SSH connections |
| `SSH_PUBLIC_URL` | `ssh://localhost:23234` | Public URL for SSH connections |
| `DB_DRIVER` | `sqlite3` | Database driver to use |
| `DB_DATA_SOURCE` | `./tmp/terminal-pet.db` | Database connection string |

Example:
```bash
SSH_LISTEN_ADDR=0.0.0.0:2222 DB_DATA_SOURCE=./data/pets.db ./bin/pet-game
```

## How to Play

1. Connect to the game via SSH:
   ```
   ssh localhost -p 23235
   ```
   
   > **Important**: Use SSH keys for authentication to ensure your pet is saved and restored properly between sessions. The public key is used to identify you and associate you with your pet.

2. Use arrow keys (or j/k) to navigate menu options
3. Press Enter or Space to select an option
4. Press q or Ctrl+C to quit
5. Press ? to toggle help

## Generate SSH key

If you don't have an SSH key, you can generate one using the following command:

```bash
ssh-keygen -o
```

## Persistence

Your pet's state is automatically saved when you disconnect and restored when you reconnect. This includes:

- Pet's name and age
- Hunger, happiness, and health levels
- All other stats

The system uses your SSH public key as a unique identifier to associate you with your pet, so make sure to use the same key when reconnecting.

## Game Actions

- **Feed**: Feed your pet to reduce hunger
- **Clean**: Clean your pet's living area
- **Play**: Play with your pet to increase happiness
- **Medicine**: Use when your pet is sick
- **Sleep**: Put your pet to sleep to restore health

## Pet Care Instructions

- Feed your pet regularly to prevent hunger
- Play with your pet to keep it happy
- If you neglect your pet, it will become sad and eventually die
- Each pet has its own personality and needs

## Build

To build:

```
go build -o ./bin/pet-game ./cmd/main.go
```


## License

MIT
