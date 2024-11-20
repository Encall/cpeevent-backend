# CPEEVO Backend

This project is the backend for the CPEEVO application, built using the Gin framework and MongoDB. It provides various APIs for managing events, users, posts, and authentication.

## Project Structure

```
.
├── .air.toml
├── .dockerignore
├── .env
├── .env.template
├── .github/
│   └── workflows/
│       └── go.yml
├── .gitignore
├── .vscode/
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── go.sum
├── kubernetes/
│   ├── configmap.yaml
│   ├── deployment.yaml
│   ├── secret.yaml
│   ├── secret.yaml.template
│   └── service.yaml
├── main.go
├── README.md
└── src/
    ├── controllers/
    │   ├── accountController.go
    │   ├── eventController.go
    │   ├── postController.go
    │   └── userController.go
    ├── database/
    │   └── db.go
    ├── helpers/
    │   └── tokenHelper.go
    ├── middleware/
    │   └── authMiddleware.go
    ├── models/
    │   ├── auth.go
    │   └── event.go
    │   └── post.go
    └── routes/
        └── route.go
```

## Getting Started

### Prerequisites

- Go 1.23 or later
- Docker
- Kubernetes
- MongoDB

### Setup

1. **Clone the repository:**

    ```sh
    git clone https://github.com/yourusername/cpeevent-backend.git
    cd cpeevent-backend
    ```

2. **Create the .env file:**

    Copy the .env.template file to .env and fill in your MongoDB connection details.

    ```sh
    cp .env.template .env
    ```

3. **Install dependencies:**

    ```sh
    go mod download
    ```

### Running the Application

#### Using Docker

1. **Build and run the Docker container:**

    ```sh
    docker-compose up --build
    ```

2. **Access the application:**

    Open your browser and go to 
    ```
    http://localhost:8080
    ```

#### Using Kubernetes

Example configuration is for deploying with Oracle Kubernetes Engine.

1. **Apply Kubernetes configurations:**

    ```sh
    kubectl apply -f kubernetes/configmap.yaml
    kubectl apply -f kubernetes/secret.yaml
    kubectl apply -f kubernetes/deployment.yaml
    kubectl apply -f kubernetes/service.yaml
    ```

2. **Access the application:**

    Open your browser and go to the external IP provided by the Kubernetes service.

### Running Locally

1. **Run the application:**

    ```sh
    go run main.go
    ```

2. **Access the application:**

    Open your browser and go to 
    ```
    http://localhost:8080
    ```

## Environment Variables

- `MONGO_URI` - MongoDB connection string
- `DATABASE_NAME` - Name of the MongoDB database
- `SECRET_KEY` - Secret key for JWT
- `GIN_MODE` - Gin mode (`release` or `debug`)
- `ORIGIN_URL` - Allowed origin URL for CORS

## License

This project is licensed under the MIT License. See the LICENSE file for details.

## Acknowledgements

- [Gin Framework](https://github.com/gin-gonic/gin)
- [MongoDB Go Driver](https://github.com/mongodb/mongo-go-driver)
- [Docker](https://www.docker.com/)
- [Kubernetes](https://kubernetes.io/)