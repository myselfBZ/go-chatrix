# Chatrix Backend

Chatrix is a **real-time, event-driven messaging backend** built using **Go** and **WebSockets**. It enables instant messaging with automatic conversation creation, user search, and message delivery tracking.

## Features

✅ **WebSocket-based real-time messaging**
✅ **JWT authentication** for secure communication
✅ **Automatic conversation creation** when users message each other for the first time
✅ **Message delivery tracking** (✅ seen status, pending messages, etc.)
✅ **User search functionality**
✅ **Event-driven architecture** for seamless communication

## Tech Stack

- **Golang** (backend)
- **WebSockets** for real-time communication
- **PostgreSQL** (or any SQL DB) for data storage
- **JWT** for authentication

## Getting Started

### Prerequisites
Make sure you have the following installed:
- Go 1.21+
- PostgreSQL
- Redis (optional, for scalability)

### Installation

Clone the repository:
```sh
 git clone https://github.com/yourusername/chatrix.git
 cd chatrix
```

Install dependencies:
```sh
go mod tidy
```

### Configuration

Set up your **environment variables** in a `.env` file:
```
DATABASE_URL=postgres://user:password@localhost:5432/chatrix?sslmode=disable
JWT_SECRET=your_secret_key
PORT=8080
```

### Running the Server
```sh
go run main.go
```

## WebSocket API

### Connecting
Clients should connect via WebSockets:
```js
const ws = new WebSocket("ws://localhost:8080/ws");
ws.onopen = () => {
    ws.send(JSON.stringify({ token: "your_jwt_token" }));
};
```

### Events

| Event Type       | Description                                      |
|------------------|--------------------------------------------------|
| `TEXT`          | Send a message to another user                  |
| `DELIVERED`     | Message delivered confirmation                   |
| `PROFILE_INFO`  | Sends user profile info after authentication    |
| `CHATPREVIEWS`  | Sends list of recent chats on connection        |
| `SearchUserRequest` | Search for a user                            |
| `SearchUserResponse` | Response with matching users                |

### Sending a Message
```js
ws.send(JSON.stringify({
    type: 0, // TEXT
    body: JSON.stringify({
        to: "username",
        content: "Hello!",
        mark: 12345
    })
}));
```

### Receiving a Message
```js
ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    console.log("New message: ", data);
};
```

## Contributing
Feel free to fork and improve Chatrix! Open a pull request if you have something awesome to add.
