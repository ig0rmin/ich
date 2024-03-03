ich means Igor's Chat.

## Testing from CLI

Create one ore more test users:
```bash
curl -X POST -d '{"email":"igor@example.com", "username":"Igor", "password":"qwerty123456"}' localhost:8080/createUser
```

```bash
TOKEN=$(curl -s -X POST -d '{"email":"igor@example.com", "password":"qwerty123456"}' localhost:8080/login | jq -r '.token')
```

```bash
curl -H "Authorization: Bearer $TOKEN" localhost:8080/auth-test
```

```bash
websocat -H="Authorization: Bearer $TOKEN" ws://localhost:8080/join
```

Output:

```
{"message":"hello to the lobby"}
{"message":"{\"message\": \"Igor joined the chat\"}"}
```