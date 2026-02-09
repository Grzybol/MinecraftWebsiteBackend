# Eco Balance Endpoint

## Overview

The eco balance endpoint allows you to retrieve a player's economic balance from the Minecraft server through a Unix socket connection.

## Endpoint

```
GET /api/getEcoBalance
```

## Authentication

**Required**: JWT token in Authorization header

```
Authorization: Bearer <your-jwt-token>
```

## Rate Limiting

**Rate Limit**: 5 requests per millisecond

## Query Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `server` | string | Yes | The server name (e.g., "boxpvp", "survival", "skygen") |
| `player` | string | Yes | The player's username |

## Example Request

```bash
curl -X GET "https://bestservers.fun:8443/api/getEcoBalance?server=boxpvp&player=PlayerName" \
  -H "Authorization: Bearer your-jwt-token"
```

## Response

### Success Response (200 OK)

The response is a JSON object returned directly from the Minecraft server socket:

```json
{
  "balance": 1000.50,
  "currency": "coins",
  "player": "PlayerName"
}
```

**Note**: The exact response format depends on the Minecraft server's eco plugin implementation.

### Error Responses

#### 400 Bad Request
```json
{
  "error": "missing server name"
}
```
or
```json
{
  "error": "missing player name"
}
```

#### 401 Unauthorized
```json
{
  "error": "Missing token"
}
```
or
```json
{
  "error": "Invalid token format"
}
```
or
```json
{
  "error": "Invalid token"
}
```

#### 403 Forbidden
```json
{
  "error": "unauthorized or unknown server"
}
```

#### 429 Too Many Requests
```json
{
  "error": "Zbyt wiele zapytań do /api/getEcoBalance – spróbuj ponownie za 0.0 sekund."
}
```

#### 500 Internal Server Error
```json
{
  "error": "failed to connect to socket: connection refused"
}
```

## Technical Details

### Socket Communication

The endpoint communicates with the Minecraft server through a Unix socket:

- **Socket Path**: `/tmp/plugin_minecraft_ecobalance.sock`
- **Protocol**: Unix domain socket
- **Message Format**: `token:playerName\n`
- **Response**: JSON string terminated with newline

### Server Tokens

The endpoint uses server-specific authentication tokens configured in `config.go`. See setup instructions for configuration details.

### Logging

All requests are logged with the following information:
- Client IP address
- Request details (server, player, socket path, token)
- Success/failure status
- Response data (truncated for security)

## Usage Examples

### JavaScript/Fetch

```javascript
const response = await fetch('/api/getEcoBalance?server=boxpvp&player=PlayerName', {
  headers: {
    'Authorization': 'Bearer ' + token
  }
});

if (response.ok) {
  const balance = await response.json();
  console.log('Player balance:', balance);
} else {
  const error = await response.json();
  console.error('Error:', error.error);
}
```

### Python/Requests

```python
import requests

response = requests.get(
    'https://bestservers.fun:8443/api/getEcoBalance',
    params={'server': 'boxpvp', 'player': 'PlayerName'},
    headers={'Authorization': f'Bearer {token}'}
)

if response.status_code == 200:
    balance = response.json()
    print(f'Player balance: {balance}')
else:
    error = response.json()
    print(f'Error: {error["error"]}')
```

## Notes

- The endpoint requires the Minecraft server to have the eco balance plugin running
- The socket must be accessible at `/tmp/plugin_minecraft_ecobalance.sock`
- Server names are case-insensitive and will be converted to lowercase
- All requests are logged for monitoring and debugging purposes
- Server tokens must be configured in `config.go` (see setup instructions) 