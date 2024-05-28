# Gophermart Loyalty System

## Overview

Gophermart is an HTTP API-based loyalty system designed to manage user registrations, authentication, order tracking, and reward points accumulation for users. The system enables users to register, log in, submit order numbers for reward calculations, and manage their loyalty points.

## Features

- **User Registration and Authentication**: Users can register and log in using their credentials.
- **Order Submission**: Users can submit order numbers to the system for reward calculation.
- **Order Tracking**: Users can track the status of their submitted orders and view reward points.
- **Loyalty Points Management**: Users can view their current balance and withdraw points to use for new orders.
- **Integration with External Accrual System**: The system interacts with an external service to calculate reward points for submitted orders.

## API Endpoints

### User Registration

**Endpoint**: `POST /api/user/register`

Registers a new user.

**Request Body**:
```json
{
    "login": "<login>",
    "password": "<password>"
}
```

**Responses**:
- `200 OK`: User successfully registered and authenticated.
- `400 Bad Request`: Invalid request format.
- `409 Conflict`: Login already in use.
- `500 Internal Server Error`: Server error.

### User Login

**Endpoint**: `POST /api/user/login`

Authenticates a user.

**Request Body**:
```json
{
    "login": "<login>",
    "password": "<password>"
}
```

**Responses**:
- `200 OK`: User successfully authenticated.
- `400 Bad Request`: Invalid request format.
- `401 Unauthorized`: Invalid login/password.
- `500 Internal Server Error`: Server error.

### Submit Order

**Endpoint**: `POST /api/user/orders`

Submits an order number for reward calculation.

**Request Body** (text/plain):
```
12345678903
```

**Responses**:
- `200 OK`: Order number already submitted by the user.
- `202 Accepted`: New order number accepted for processing.
- `400 Bad Request`: Invalid request format.
- `401 Unauthorized`: User not authenticated.
- `409 Conflict`: Order number already submitted by another user.
- `422 Unprocessable Entity`: Invalid order number format.
- `500 Internal Server Error`: Server error.

### Get User Orders

**Endpoint**: `GET /api/user/orders`

Retrieves the list of submitted orders, their statuses, and reward information.

**Responses**:
- `200 OK`: Successfully retrieved order list.
- `204 No Content`: No data available.
- `401 Unauthorized`: User not authenticated.
- `500 Internal Server Error`: Server error.

### Get User Balance

**Endpoint**: `GET /api/user/balance`

Retrieves the current loyalty points balance and total points withdrawn.

**Responses**:
- `200 OK`: Successfully retrieved balance.
- `401 Unauthorized`: User not authenticated.
- `500 Internal Server Error`: Server error.

### Withdraw Points

**Endpoint**: `POST /api/user/balance/withdraw`

Requests to withdraw points for a new order.

**Request Body**:
```json
{
    "order": "2377225624",
    "sum": 751
}
```

**Responses**:
- `200 OK`: Successfully processed request.
- `401 Unauthorized`: User not authenticated.
- `402 Payment Required`: Insufficient funds.
- `422 Unprocessable Entity`: Invalid order number.
- `500 Internal Server Error`: Server error.

### Get Withdrawals

**Endpoint**: `GET /api/user/withdrawals`

Retrieves information about withdrawals made by the user.

**Responses**:
- `200 OK`: Successfully retrieved withdrawal information.
- `204 No Content`: No withdrawals available.
- `401 Unauthorized`: User not authenticated.
- `500 Internal Server Error`: Server error.

## Installation

1. Clone the repository:
   ```sh
   git clone https://github.com/yourusername/gophermart.git
   cd gophermart
   ```

2. Set up the environment variables:
   ```sh
   export RUN_ADDRESS=":8080"
   export DATABASE_URI="postgres://user:password@localhost:5432/gophermart"
   export ACCRUAL_SYSTEM_ADDRESS="http://localhost:8081"
   ```

3. Build and run the application:
   ```sh
   go build -o gophermart cmd/gophermart/main.go
   ./gophermart
   ```

4. Alternatively, use Docker:
   ```sh
   docker-compose up
   ```

## Configuration

The service supports configuration via environment variables or command-line flags:

- `RUN_ADDRESS`: Address and port for the service (default: `:8080`).
- `DATABASE_URI`: Connection URI for the PostgreSQL database.
- `ACCRUAL_SYSTEM_ADDRESS`: Address of the external accrual system.

## Project Structure

- `.github`: GitHub workflows and issue templates.
- `cmd`: Entry points for the application.
- `internal`: Internal application code including handlers, services, and database interactions.
- `migrations`: Database migration scripts.
- `docker-compose.yml`: Docker Compose configuration.
- `go.mod`, `go.sum`: Go module dependencies.
- `README.md`: Project documentation.
- `SPECIFICATION.md`: Detailed project specification.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## License

This project is licensed under the MIT License.

 