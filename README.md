ich means Igor's Chat.

## Testing from CLI

Create one ore more test users:
```bash
curl -X POST -d '{"email":"igor@example.com", "username":"Igor", "password":"qwerty123456"}' localhost:8080/createUser
```

Authenticate and get the access token:
```bash
TOKEN=$(curl -s -X POST -d '{"email":"igor@example.com", "password":"qwerty123456"}' localhost:8080/login | jq -r '.token')
```

Join the chat:
```bash
websocat -H="Authorization: Bearer $TOKEN" ws://localhost:8080/join
```

You should see somethign like this:

```
{"type":"users_online","sent_at":"2024-03-04T09:16:14.60911171+02:00","msg":{"list":[]}}
{"type":"user_joined","sent_at":"2024-03-04T09:16:14.610345784+02:00","msg":{"username":"Igor"}}
```

Post a message to the chat (by entering the JSON in the console). You should get somethign like this:

```
{"type":"chat_message", "msg": {"text": "Hello!"}}
{"type":"chat_message","sent_at":"2024-03-04T09:19:55.903219951+02:00","msg":{"from_user":"Igor","text":"Hello!"}}
```