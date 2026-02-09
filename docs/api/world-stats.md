# World Stats API

## Endpoints

```
GET /stats/:world
```

Retrieves all statistics for the specified Minecraft world. This endpoint is public and protected with lightweight rate limiting.

### Path Parameters

| Parameter | Type   | Description                  |
|-----------|--------|------------------------------|
| `world`   | string | Name of the world to query.  |

### Response

Returns an array of objects mirroring rows from the `questmc_playerstats_world_stats` table, including the player's UUID and in-game name for each statistic row.

```json
[
  {
    "player_uuid": "8d8a76fd-2d94-4de3-92b4-4de0873d3fb3",
    "player_name": "Notch",
    "world": "world_nether",
    "statistic": "minecraft:play_one_minute",
    "value": 12345
  }
]
```

If no rows are found the endpoint returns an empty array.

### Error Responses

- `400 Bad Request` – The `world` parameter was missing.
- `500 Internal Server Error` – The service was unable to connect to the database or read the results.

---

``` 
GET /stats/worlds
```

Returns the list of unique worlds available in the `questmc_playerstats_world_stats` table, ordered alphabetically.

### Response

```json
[
  "world",
  "world_nether",
  "world_the_end"
]
```

If no worlds are present the endpoint returns an empty array.

### Error Responses

- `500 Internal Server Error` – The service was unable to connect to the database or read the results.

## Notes

- Data is sourced from the `questmc_playerstats_world_stats` MariaDB table.
- Values are returned verbatim without aggregation or transformation to match the frontend expectations.
