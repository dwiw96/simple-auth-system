# Simple Auth System

## Overview
A REST API focusing on user authentication, including features such as Sign Up, Login, Logout, Email Verification, and access and refresh token management.
This project is built with golang and postgres for data storage and redis for caching blocked tokens.

## Features
- **Sign Up**: Register new users and send email verification
- **Login**: Authenticate users and provide access and refresh token.
- **Logout**: Invalidate access token by adding to the blocklist in redis and invalidate refresh token by delete the token from postgres
- **Email Verification**: Verify user email.
- **Email Unverfication**: Delete user data from postgres
- **Access & Refresh Token**: Use JWTs for access token and UUID for refresh token for session management.

## Technologies Used
- **Programming Language**: Golang
- **Database**: Postgresql
- **Cache**: Redis
- **Token**: JWT and UUID
- **Docker**: Postgresql & Redis

## Getting Started

### Instalation
- **Clone the Repository**:
```
git clone https://github.com/dwiw96/simple-auth-system.git
cd go/
```
- **Set up Environment Variable**:
```
open .env file
```
- **Install Dependencies**:
```
go mod tidy
```
- **Run Database Migration**
```
open Makefile
make migrate-up
```
- **Start the App**:
```
go run main.go
```
- **Access the API**:
The default would be be in http://localhost:8080

## Documentation
- **/go/docs/user-api.json**: openAPI spesicifation
- **/go/docs/db.erd**: database entity relationship
- **/go/docs/sequenceDiagam_*...*.pu**: sequence diagram
- **/go/docs/simple-auth-system.postman_collection.json**: postman specification

## License
This `README.md` includes all the key sections for a backend authentication API project, providing an overview of the functionality, setup instructions, API documentation, security considerations, and sequence diagrams. You can adjust details as needed, such as adding links to specific files or diagrams.

