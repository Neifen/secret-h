# Secret-H Voting Helper

A Go and HTMX web application to streamline voting and execution mechanics in the *Secret-H* board game. This app ensures synchronized chancellor voting and clear confirmation of eliminations, replacing physical cards and reducing ambiguity.

## Features
- **Synchronized Voting**: Ensures all players vote for the chancellor simultaneously.
- **Clear Execution Confirmation**: Confirms whether a player has been eliminated, removing uncertainty.
- **Real-Time Updates**: Built with HTMX for dynamic, responsive interactions without page reloads.
- **Dockerized Deployment**: Easy setup and deployment using Docker.

## Prerequisites
- Docker (for running the application)
- A modern web browser for accessing the app

## Getting Started

### Running with Docker
The application is available as a Docker image on Docker Hub.

1. Pull the latest image:
   ```bash
   docker pull docker.io/neifen/secret-h:latest
   ```

2. Run the container, mapping port `8148`:
   ```bash
   docker run -d -p 8148:8148 docker.io/neifen/secret-h:latest
   ```

3. Access the app in your browser at:
   ```
   http://localhost:8148
   ```

### Using Docker Compose
For a more persistent setup, use the provided Docker Compose configuration.

1. Create a `docker-compose.yml` file with the following content:
   ```yaml
   services:
     secret-h:
       image: docker.io/neifen/secret-h:latest
       ports:
         - "8148:8148"
       restart: unless-stopped
   ```

2. Run the application:
   ```bash
   docker-compose up -d
   ```

3. Access the app at:
   ```
   http://localhost:8148
   ```

4. To stop the application:
   ```bash
   docker-compose down
   ```

## Development
To run or modify the application locally without Docker:

1. Clone the repository:
   ```bash
   git clone https://github.com/Neifen/secret-h.git
   cd secret-h
   ```

2. Ensure you have Go (version 1.18 or later) installed.

3. Install dependencies:
   ```bash
   go mod tidy
   ```

4. Run the application:
   ```bash
   go run .
   ```

5. Access the app at:
   ```
   http://localhost:8148
   ```

## License
This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Contact
For issues or suggestions, please open an issue on this repository or contact the maintainer at neifen.b@gmail.com.

Happy gaming with *Secret-H*! 🎲
```