ich means Igor's Chat.

## Testing from CLI

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