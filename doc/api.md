## User Management Endpoints

These endopoints are implemented only for the testing purposes and to make the server into a finished product. In the real-world design, user creation and issuing JWT should be done by a separate entity, and the chat server should only validate JWT.

### POST /createUser

Creates a new user.

Request:

```json
{
  "email": "user_email@example.com",
  "username": "User Name",
  "password": "<password>"
}
```

Response:

```json
{
  "id": "4",
  "username": "User Name",
  "email": "user_email@example.com"
}
```

The endpoint does not require authentication.

### POST /login

Log-in the user and on success issues JWT authentication tocken.

Request:

```json
{
  "email": "user_name@example.com",
  "password": "<password>"
}
```

Response:
```json
{
  "id": "4",
  "username": "User Name",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjEiLCJ1c2VybmFtZSI6Iklnb3IiLCJpc3MiOiIxIiwiZXhwIjoxNzA5NjI3MzkxfQ.Ln6B22AzEqkwIxL8MKIHejrycHkb1iw03a2H5LqQx6g"
}
```

The endpoint does not require authentication.

## Chat API

The server communicate with the client using a stream of JSON messages over a websockets connection. Authentication is done by JWT passed in the `Authorization: Bearer <TOKEN>` header.

### POST /join

Opens a websocket connection to the chat server. The server will send a stream of JSON messages and read JSON sent from the client. The possible message types are listed below. To leave the chat close the websocket connection.

### users_online

From the server to client. This message is sent as the first message when a new client joins. It contains the list of the users currently in the chat.

Example:

```json
{
  "type": "users_online",
  "sent_at": "2024-03-04T09:47:45.137360195+02:00",
  "msg": {
    "list": [
      "Patrick",
      "Bob"
    ]
  }
}
```

### user_joined

From the server to client. Sent when a new user joins the chat.

Example:
```json
{
  "type": "user_joined",
  "sent_at": "2024-03-04T09:47:45.139425846+02:00",
  "msg": {
    "username": "Igor"
  }
}
```

### user_left

From the server to client. Sent when a user leaves the chat.

Example:
```json
{
  "type": "user_left",
  "sent_at": "2024-03-04T09:48:35.973528787+02:00",
  "msg": {
    "username": "Bob"
  }
}
```

### chat_message

From the server to client and from the client to server. Contains the message sent to the chat.

 When this message is sent from client to server, `from_user` and `sent_at` are ignored to prevent spoofing. `from_user` is automatically set by the server to the username of the current user (taken from JWT) and `sent_at` is set to the current time.

Example:
```json
{
  "type": "chat_message",
  "sent_at": "2024-03-04T09:48:30.59855695+02:00",
  "msg": {
    "from_user": "Bob",
    "text": "Hello!"
  }
}
```