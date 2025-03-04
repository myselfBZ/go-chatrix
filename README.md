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
- **PostgreSQL** 
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
DB_NAME=chatrix
DB_HOST=localhost
DB_PORT=32768
DB_USER=postgres
DB_PASSWORD=new_password
```

### Running the Server
```sh
make run
```

## WebSocket API

### Connecting
Clients should connect via WebSockets, (and send token after the handshake):
```js
const ws = new WebSocket("ws://localhost:8080/ws");
ws.onopen = () => {
    ws.send(JSON.stringify({ token: "your_jwt_token" }));
};
```

### Events

TEXT EventType = iota
DELIVERED 
MARK_READ
ERR
PROFILE_INFO
CHATPREVIEWS

SearchUserRequest
SearchUserResponse

| Event Type         | Description                                      | Code |
|--------------------|--------------------------------------------------|------|
| `TEXT`            | Send a message to another user                   | 0    |
| `DELIVERED`       | Message delivered confirmation                    | 1    |
| `MARK_READ`       | Mark a message as read                           | 2    |
| `ERR`             | Error message                                    | 3    |
| `PROFILE_INFO`    | Sends user profile info after authentication      | 4    |
| `CHATPREVIEWS`    | Sends list of recent chats on connection          | 5    |
| `SearchUserRequest`  | Request to search for a user                     | 6    |
| `SearchUserResponse` | Response with matching users                    | 7    |

### Sending a Message
```js
ws.send(JSON.stringify({
    type: 0, // TEXT
    body: JSON.stringify({
        to: "username",
        content: "Hello!",
        mark: `some unique mark for clients to recognize the DELIVERD event for this message`
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
