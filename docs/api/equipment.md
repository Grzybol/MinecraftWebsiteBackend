# Equipment Endpoints

## Overview

The equipment endpoints allow you to retrieve player equipment data from Minecraft servers through Unix socket connections.

## Available Endpoints

### Get Backpack
```
GET /api/getBackpack
```

### Get Armor
```
GET /api/getArmor
```

### Get Hotbar
```
GET /api/getHotbar
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
curl -X GET "https://bestservers.fun:8443/api/getBackpack?server=boxpvp&player=PlayerName" \
  -H "Authorization: Bearer your-jwt-token"
```

## Response

### Success Response (200 OK)

The response is a JSON object returned directly from the Minecraft server socket. The exact format depends on the equipment type and server plugin implementation.

**Backpack Example:**
```json
{
  "items": [
    {
      "id": "minecraft:diamond_sword",
      "count": 1,
      "enchantments": ["sharpness:5"]
    }
  ],
  "player": "PlayerName"
}
```

**Armor Example:**
```json
{
  "helmet": {"id": "minecraft:diamond_helmet", "enchantments": ["protection:4"]},
  "chestplate": {"id": "minecraft:diamond_chestplate", "enchantments": ["protection:4"]},
  "leggings": {"id": "minecraft:diamond_leggings", "enchantments": ["protection:4"]},
  "boots": {"id": "minecraft:diamond_boots", "enchantments": ["protection:4"]}
}
```

**Hotbar Example:**
```json
{
  "slots": [
    {"id": "minecraft:diamond_sword", "count": 1},
    {"id": "minecraft:bow", "count": 1},
    null,
    {"id": "minecraft:arrow", "count": 64}
  ]
}
```

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

#### 403 Forbidden
```json
{
  "error": "unauthorized or unknown server"
}
```

#### 429 Too Many Requests
```json
{
  "error": "Zbyt wiele zapytań do /api/getBackpack – spróbuj ponownie za 0.0 sekund."
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

Each endpoint communicates with the Minecraft server through a Unix socket:

- **Backpack Socket**: `/tmp/plugin_{server}_backpack.sock`
- **Armor Socket**: `/tmp/plugin_{server}_armor.sock`
- **Hotbar Socket**: `/tmp/plugin_{server}_hotbar.sock`

**Protocol**: Unix domain socket
**Message Format**: `token:playerName\n`
**Response**: JSON string terminated with newline

### Server Tokens

The endpoints use server-specific authentication tokens configured in `config.go`. See setup instructions for configuration details.

### Logging

All requests are logged with the following information:
- Client IP address
- Request details (equipment type, server, player, socket path, token)
- Success/failure status
- Response data (truncated for security)

## Usage Examples

### JavaScript/Fetch

```javascript
async function getPlayerEquipment(server, player, equipmentType) {
  const response = await fetch(`/api/get${equipmentType}?server=${server}&player=${player}`, {
    headers: {
      'Authorization': 'Bearer ' + token
    }
  });

  if (response.ok) {
    const equipment = await response.json();
    console.log(`Player ${equipmentType}:`, equipment);
    return equipment;
  } else {
    const error = await response.json();
    console.error('Error:', error.error);
    throw new Error(error.error);
  }
}

// Usage
getPlayerEquipment('boxpvp', 'PlayerName', 'Backpack');
getPlayerEquipment('boxpvp', 'PlayerName', 'Armor');
getPlayerEquipment('boxpvp', 'PlayerName', 'Hotbar');
```

### Python/Requests

```python
import requests

def get_player_equipment(server, player, equipment_type):
    response = requests.get(
        f'https://bestservers.fun:8443/api/get{equipment_type}',
        params={'server': server, 'player': player},
        headers={'Authorization': f'Bearer {token}'}
    )
    
    if response.status_code == 200:
        equipment = response.json()
        print(f'Player {equipment_type}: {equipment}')
        return equipment
    else:
        error = response.json()
        print(f'Error: {error["error"]}')
        raise Exception(error["error"])

# Usage
get_player_equipment('boxpvp', 'PlayerName', 'Backpack')
get_player_equipment('boxpvp', 'PlayerName', 'Armor')
get_player_equipment('boxpvp', 'PlayerName', 'Hotbar')
```

## Notes

- Each endpoint requires the Minecraft server to have the corresponding equipment plugin running
- Sockets must be accessible at the specified paths
- Server names are case-insensitive and will be converted to lowercase
- All requests are logged for monitoring and debugging purposes
- The response format may vary depending on the server's plugin implementation
- Server tokens must be configured in `config.go` (see setup instructions) 