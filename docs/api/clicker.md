# Clicker Game API

## Overview

The Clicker Game API provides endpoints for managing player progress in the clicker game, including synchronization, progress retrieval, and leaderboards.

## Available Endpoints

### Sync Progress
```
POST /api/progress/sync
```

### Get Progress
```
GET /api/progress
```

### Get Leaderboard
```
GET /api/leaderboard
```

## Authentication

**Required**: JWT token in Authorization header

```
Authorization: Bearer <your-jwt-token>
```

## Rate Limiting

- Sync Progress: 10 requests per millisecond
- Get Progress: 5 requests per millisecond
- Leaderboard: 10 requests per millisecond

## Data Models

### Player Progress Structure

```json
{
  "playerName": "Steve",
  "level": 12,
  "experience": 12345.0,
  "experienceToNextLevel": 15000.0,
  "coins": 123456.0,
  "diamonds": 50.0,
  "premiumBalance": 10.0,
  "betterCoinBalance": 5.0,
  "totalCoinsEarned": 999999.0,
  "totalDiamondsEarned": 100.0,
  "totalClicks": 12345,
  "totalUpgrades": 20,
  "upgrades": [
    { "id": "wooden_pickaxe", "level": 5 },
    { "id": "stone_pickaxe", "level": 2 }
  ],
  "lastSync": "2024-05-10T12:34:56Z",
  "lastDevice": "android",
  "lastIp": "1.2.3.4"
}
```

### Upgrade Structure

```json
{
  "id": "wooden_pickaxe",
  "level": 5
}
```

## Endpoint Details

### POST /api/progress/sync

Synchronizes player progress with the server. Should be called every 10 seconds or on important events.

**⚠️ Important**: This endpoint includes anti-cheat verification to prevent progress manipulation.

#### Request Body

```json
{
  "level": 12,
  "experience": 12345.0,
  "experienceToNextLevel": 15000.0,
  "coins": 123456.0,
  "diamonds": 50.0,
  "premiumBalance": 10.0,
  "betterCoinBalance": 5.0,
  "totalCoinsEarned": 999999.0,
  "totalDiamondsEarned": 100.0,
  "totalClicks": 12345,
  "totalUpgrades": 20,
  "upgrades": [
    { "id": "wooden_pickaxe", "level": 5 },
    { "id": "stone_pickaxe", "level": 2 }
  ],
  "timestamp": "2024-05-10T12:34:56Z",
  "previousCoins": 123000.0,
  "previousExperience": 12000.0,
  "previousDiamonds": 48.0,
  "previousLevel": 11,
  "previousTimestamp": "2024-05-10T12:34:46Z"
}
```

**Optional fields for verification** (if not provided, server will use last saved state):
- `previousCoins` - Coins before sync
- `previousExperience` - Experience before sync  
- `previousDiamonds` - Diamonds before sync
- `previousLevel` - Level before sync
- `previousTimestamp` - Timestamp before sync

#### Success Response (200 OK)

```json
{
  "success": true,
  "message": "Progress synced successfully",
  "serverState": {
    // Optional: Current server state if includeServerState=true
  }
}
```

#### Error Responses

##### 400 Bad Request
```json
{
  "error": "Invalid request format"
}
```
or
```json
{
  "error": "Invalid values in request"
}
```
or
```json
{
  "error": "Progress verification failed",
  "details": {
    "isValid": false,
    "reason": "Coins gain too high: 1000.00 > 500.00",
    "maxPossibleCoins": 500.0,
    "actualCoinsGain": 1000.0,
    "deltaTime": 10.5
  }
}
```

##### 401 Unauthorized
```json
{
  "error": "Missing token"
}
```

##### 429 Too Many Requests
```json
{
  "error": "Zbyt wiele zapytań do /api/progress/sync – spróbuj ponownie za 0.0 sekund."
}
```

##### 500 Internal Server Error
```json
{
  "error": "Failed to create progress"
}
```
or
```json
{
  "error": "Failed to update progress"
}
```

#### Query Parameters

- `includeServerState=true` - Include current server state in response

### GET /api/progress

Retrieves the current player progress from the server.

#### Success Response (200 OK)

```json
{
  "success": true,
  "progress": {
    "playerName": "Steve",
    "level": 12,
    "experience": 12345.0,
    "experienceToNextLevel": 15000.0,
    "coins": 123456.0,
    "diamonds": 50.0,
    "premiumBalance": 10.0,
    "betterCoinBalance": 5.0,
    "totalCoinsEarned": 999999.0,
    "totalDiamondsEarned": 100.0,
    "totalClicks": 12345,
    "totalUpgrades": 20,
    "upgrades": [
      { "id": "wooden_pickaxe", "level": 5 },
      { "id": "stone_pickaxe", "level": 2 }
    ],
    "lastSync": "2024-05-10T12:34:56Z",
    "lastDevice": "android",
    "lastIp": "1.2.3.4"
  }
}
```

#### Default Values for New Players

If a player has no progress record, the endpoint returns default values:

```json
{
  "success": true,
  "progress": {
    "playerName": "Steve",
    "level": 1,
    "experience": 0,
    "experienceToNextLevel": 100,
    "coins": 0,
    "diamonds": 0,
    "premiumBalance": 0,
    "betterCoinBalance": 0,
    "totalCoinsEarned": 0,
    "totalDiamondsEarned": 0,
    "totalClicks": 0,
    "totalUpgrades": 0,
    "upgrades": [],
    "lastSync": "2024-05-10T12:34:56Z",
    "lastDevice": "android",
    "lastIp": "1.2.3.4"
  }
}
```

### GET /api/leaderboard

Retrieves the top players leaderboard sorted by level and experience.

#### Success Response (200 OK)

```json
{
  "success": true,
  "leaderboard": [
    {
      "playerName": "TopPlayer",
      "level": 50,
      "experience": 99999.0,
      // ... other progress fields
    },
    {
      "playerName": "SecondPlayer",
      "level": 45,
      "experience": 85000.0,
      // ... other progress fields
    }
  ]
}
```

## Usage Examples

### JavaScript/Fetch

#### Sync Progress

```javascript
async function syncProgress(progressData) {
  const response = await fetch('/api/progress/sync', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer ' + token
    },
    body: JSON.stringify(progressData)
  });

  if (response.ok) {
    const result = await response.json();
    console.log('Progress synced:', result.message);
    return result;
  } else {
    const error = await response.json();
    console.error('Sync failed:', error.error);
    throw new Error(error.error);
  }
}

// Usage
const progressData = {
  level: 12,
  experience: 12345.0,
  experienceToNextLevel: 15000.0,
  coins: 123456.0,
  diamonds: 50.0,
  premiumBalance: 10.0,
  betterCoinBalance: 5.0,
  totalCoinsEarned: 999999.0,
  totalDiamondsEarned: 100.0,
  totalClicks: 12345,
  totalUpgrades: 20,
  upgrades: [
    { id: "wooden_pickaxe", level: 5 },
    { id: "stone_pickaxe", level: 2 }
  ],
  timestamp: new Date().toISOString()
};

try {
  await syncProgress(progressData);
} catch (error) {
  console.error('Failed to sync progress:', error);
}
```

#### Get Progress

```javascript
async function getProgress() {
  const response = await fetch('/api/progress', {
    headers: {
      'Authorization': 'Bearer ' + token
    }
  });

  if (response.ok) {
    const result = await response.json();
    console.log('Current progress:', result.progress);
    return result.progress;
  } else {
    const error = await response.json();
    console.error('Failed to get progress:', error.error);
    throw new Error(error.error);
  }
}

// Usage
try {
  const progress = await getProgress();
  // Update game state with progress
  gameState.level = progress.level;
  gameState.coins = progress.coins;
  // ...
} catch (error) {
  console.error('Failed to load progress:', error);
}
```

#### Get Leaderboard

```javascript
async function getLeaderboard() {
  const response = await fetch('/api/leaderboard', {
    headers: {
      'Authorization': 'Bearer ' + token
    }
  });

  if (response.ok) {
    const result = await response.json();
    console.log('Leaderboard:', result.leaderboard);
    return result.leaderboard;
  } else {
    const error = await response.json();
    console.error('Failed to get leaderboard:', error.error);
    throw new Error(error.error);
  }
}

// Usage
try {
  const leaderboard = await getLeaderboard();
  // Display leaderboard in UI
  displayLeaderboard(leaderboard);
} catch (error) {
  console.error('Failed to load leaderboard:', error);
}
```

### Python/Requests

#### Sync Progress

```python
import requests
import json
from datetime import datetime

def sync_progress(progress_data, token):
    response = requests.post(
        'https://bestservers.fun:8443/api/progress/sync',
        headers={
            'Content-Type': 'application/json',
            'Authorization': f'Bearer {token}'
        },
        json=progress_data
    )
    
    if response.status_code == 200:
        result = response.json()
        print(f'Progress synced: {result["message"]}')
        return result
    else:
        error = response.json()
        print(f'Sync failed: {error["error"]}')
        raise Exception(error['error'])

# Usage
progress_data = {
    "level": 12,
    "experience": 12345.0,
    "experienceToNextLevel": 15000.0,
    "coins": 123456.0,
    "diamonds": 50.0,
    "premiumBalance": 10.0,
    "betterCoinBalance": 5.0,
    "totalCoinsEarned": 999999.0,
    "totalDiamondsEarned": 100.0,
    "totalClicks": 12345,
    "totalUpgrades": 20,
    "upgrades": [
        {"id": "wooden_pickaxe", "level": 5},
        {"id": "stone_pickaxe", "level": 2}
    ],
    "timestamp": datetime.now().isoformat()
}

try:
    sync_progress(progress_data, token)
except Exception as e:
    print(f"Failed to sync progress: {e}")
```

## Best Practices

### Synchronization

- **Frequency**: Sync every 10 seconds or on important events (level up, purchase, etc.)
- **Error Handling**: Implement retry logic for failed syncs
- **Offline Support**: Queue syncs when offline and send when connection is restored
- **Conflict Resolution**: Use server state when conflicts occur

### Data Validation

- **Client-side**: Validate data before sending to server
- **Server-side**: Server validates all incoming data
- **Negative Values**: All numeric values must be non-negative
- **Required Fields**: All fields in the sync request are required

### Performance

- **Rate Limiting**: Respect rate limits to avoid being blocked
- **Batch Updates**: Consider batching multiple updates
- **Caching**: Cache progress data locally to reduce API calls
- **Compression**: Use gzip compression for large payloads

### Security

- **Token Management**: Keep JWT tokens secure and renew when needed
- **HTTPS**: Always use HTTPS for API calls
- **Input Validation**: Validate all user inputs
- **Logging**: Monitor API usage for suspicious activity

## Database Schema

The player progress is stored in MongoDB in the `player_progress` collection:

```javascript
{
  "_id": ObjectId("..."),
  "playerName": "Steve",
  "level": 12,
  "experience": 12345.0,
  "experienceToNextLevel": 15000.0,
  "coins": 123456.0,
  "diamonds": 50.0,
  "premiumBalance": 10.0,
  "betterCoinBalance": 5.0,
  "totalCoinsEarned": 999999.0,
  "totalDiamondsEarned": 100.0,
  "totalClicks": 12345,
  "totalUpgrades": 20,
  "upgrades": [
    { "id": "wooden_pickaxe", "level": 5 },
    { "id": "stone_pickaxe", "level": 2 }
  ],
  "lastSync": ISODate("2024-05-10T12:34:56Z"),
  "lastDevice": "android",
  "lastIp": "1.2.3.4",
  "createdAt": ISODate("2024-05-10T12:00:00Z"),
  "updatedAt": ISODate("2024-05-10T12:34:56Z")
}
```

## Notes

- All timestamps are in ISO 8601 format
- The collection is automatically created when the first player progress is saved
- Default values are provided for new players
- Leaderboard is sorted by level (descending) then by experience (descending)
- All API calls are logged for monitoring and debugging
- Rate limiting prevents abuse of the API 

## Anti-Cheat Verification

The sync endpoint includes comprehensive anti-cheat verification to prevent progress manipulation:

### Verification Rules

1. **Maximum Click Rate**: 25 clicks per second
2. **Time Validation**: Delta time cannot exceed 1 hour
3. **Rollback Detection**: Progress cannot decrease
4. **Resource Gain Limits**: Calculated based on:
   - Player level and upgrades
   - Time elapsed between syncs
   - Maximum possible income from clicks and DPS
5. **Level Gain Limits**: Maximum 5 levels per minute
6. **Safety Margin**: 5% tolerance for network delays

### Verification Formula

```
max_clicks = 25 * delta_time_seconds
max_click_income = max_clicks * click_damage * gold_bonus
max_dps_income = dps * delta_time_seconds * gold_bonus
max_possible_income = max_click_income + max_dps_income + 5%_margin
```

### Upgrade Bonuses

| Upgrade | Effect |
|---------|--------|
| wooden_pickaxe | +2.0 damage per level |
| stone_pickaxe | +5.0 damage per level |
| iron_pickaxe | +10.0 damage per level |
| diamond_pickaxe | +25.0 damage per level |
| gold_bonus | +10% gold per level |
| exp_bonus | +5% exp per level |

### Verification Failures

If verification fails, the sync is rejected with detailed error information:

```json
{
  "error": "Progress verification failed",
  "details": {
    "isValid": false,
    "reason": "Coins gain too high: 1000.00 > 500.00",
    "maxPossibleCoins": 500.0,
    "actualCoinsGain": 1000.0,
    "deltaTime": 10.5
  }
}
```

### Best Practices for Clients

1. **Send Previous State**: Include previous state in sync requests for better verification
2. **Regular Syncs**: Sync every 10 seconds to minimize verification complexity
3. **Handle Failures**: Implement retry logic for failed verifications
4. **Monitor Logs**: Check server logs for verification failures
5. **Respect Limits**: Ensure client-side calculations match server expectations 