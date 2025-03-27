# GameHub API

A modern RESTful API for accessing comprehensive video game information, including new releases, popular games, and store availability.

## Features

- Comprehensive game information
- Real-time data from multiple sources
- Store availability and direct purchase links
- Rating and review information
- Platform compatibility details
- Metacritic scores integration

## Tech Stack

- Go 1.21+
- Gin Web Framework
- RAWG Video Games Database API
- Environment-based configuration

## Prerequisites

- Go 1.21 or higher
- RAWG API Key ([Get it here](https://rawg.io/apidocs))

## Installation

1. Clone the repository
```bash
git clone https://github.com/yourusername/gamehub.git
cd gamehub
```

2. Install dependencies
```bash
go mod download
```

3. Configure environment variables
```bash
cp .env.example .env
```

4. Add your RAWG API key to `.env`:
```
PORT=8080
RAWG_API_KEY=your_api_key_here
```

## Running the Application

### Development
```bash
go run .
```

### Production
```bash
go build
./gamehub
```

The API will be available at `http://localhost:8080`

## API Endpoints

### Games Listing

| Endpoint | Description |
|----------|-------------|
| `GET /games` | List all games |
| `GET /games/latest` | Get latest releases |
| `GET /games/popular` | Get most popular games |
| `GET /games/metacritic` | Get highest rated games on Metacritic |
| `GET /games/upcoming` | Get upcoming releases |
| `GET /games/search?q={query}` | Search games by name |
| `GET /games/{id}` | Get detailed information about a specific game |

### Response Format

```json
{
  "id": "3498",
  "title": "Grand Theft Auto V",
  "description": "...",
  "background_image": "https://...",
  "genres": ["Action", "Adventure"],
  "rating": 4.48,
  "rating_top": 5,
  "release_date": "2013-09-17",
  "metacritic": 97,
  "playtime": 73,
  "platforms": ["PC", "PlayStation 5", "Xbox Series S/X"],
  "stores": [
    {
      "id": 1,
      "name": "Steam",
      "url": "https://store.steampowered.com/...",
      "image": "https://..."
    }
  ]
}
```

## Error Handling

The API uses standard HTTP status codes:

- 200: Success
- 400: Bad Request
- 404: Not Found
- 500: Internal Server Error

Error responses include a message:
```json
{
  "error": "Error description"
}
```

## Rate Limiting

The API inherits RAWG's rate limiting:
- 20,000 requests per month
- No per-second/minute rate limit

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [RAWG Video Games Database](https://rawg.io/) for providing the game data
- [Gin Web Framework](https://gin-gonic.com/) for the web framework
- All contributors who participate in this project 