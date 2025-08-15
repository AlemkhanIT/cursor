# Ecommerce Backend API

A comprehensive Golang ecommerce backend application with user authentication, product management, shopping cart, order processing, payment integration (Stripe), real-time messaging, and product reviews.

## Features

- **User Authentication & Authorization**
  - User registration with email confirmation
  - JWT-based authentication
  - Password reset functionality
  - Email confirmation required for all operations

- **Product Management**
  - CRUD operations for products
  - Product categories and search
  - Stock management
  - Product images

- **Shopping Cart**
  - Add/remove items from cart
  - Update quantities
  - Stock validation

- **Order Management**
  - Create orders from cart
  - Order status tracking (pending, paid, shipped, delivered, cancelled, in_process)
  - Product owners can manage order status
  - Stripe payment integration

- **Product Reviews**
  - 5-star rating system
  - Text comments
  - One review per user per product

- **Real-time Messaging**
  - Private messages between users
  - WebSocket-based real-time communication
  - Message read status
  - Conversation management

## Tech Stack

- **Language**: Go 1.21+
- **Framework**: Gin (HTTP web framework)
- **Database**: PostgreSQL with GORM
- **Authentication**: JWT tokens
- **Email**: SMTP with gomail
- **Payment**: Stripe
- **Real-time**: WebSocket with Gorilla WebSocket
- **Password Hashing**: bcrypt

## Prerequisites

- Go 1.21 or higher
- PostgreSQL database
- SMTP server (Gmail, SendGrid, etc.)
- Stripe account

## Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd ecommerce-app
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Set up environment variables**
   ```bash
   cp env.example .env
   ```
   
   Edit `.env` file with your configuration:
   ```env
   # Database Configuration
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=your_password
   DB_NAME=ecommerce_db

   # JWT Configuration
   JWT_SECRET=your-super-secret-jwt-key-here

   # SMTP Configuration
   SMTP_HOST=smtp.gmail.com
   SMTP_PORT=587
   SMTP_USERNAME=your-email@gmail.com
   SMTP_PASSWORD=your-app-password

   # Stripe Configuration
   STRIPE_SECRET_KEY=sk_test_your_stripe_secret_key
   STRIPE_PUBLISHABLE_KEY=pk_test_your_stripe_publishable_key

   # Server Configuration
   SERVER_PORT=8080
   ```

4. **Create PostgreSQL database**
   ```sql
   CREATE DATABASE ecommerce_db;
   ```

5. **Run the application**
   ```bash
   go run main.go
   ```

The server will start on `http://localhost:8080`

## API Endpoints

### Authentication

- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - User login
- `GET /api/auth/confirm-email` - Confirm email address
- `POST /api/auth/request-password-reset` - Request password reset
- `POST /api/auth/reset-password` - Reset password

### Products

- `GET /api/products` - Get all products (with filters)
- `GET /api/products/:id` - Get product by ID
- `POST /api/products` - Create new product (authenticated)
- `PUT /api/products/:id` - Update product (owner only)
- `DELETE /api/products/:id` - Delete product (owner only)
- `GET /api/products/my` - Get user's products (authenticated)

### Cart

- `GET /api/cart` - Get user's cart (authenticated)
- `POST /api/cart/add` - Add item to cart (authenticated)
- `PUT /api/cart/items/:id` - Update cart item quantity (authenticated)
- `DELETE /api/cart/items/:id` - Remove item from cart (authenticated)
- `DELETE /api/cart` - Clear cart (authenticated)

### Orders

- `POST /api/orders` - Create order from cart (authenticated)
- `GET /api/orders` - Get user's orders (authenticated)
- `GET /api/orders/:id` - Get order by ID (authenticated)
- `GET /api/orders/my-products` - Get orders for user's products (authenticated)
- `PUT /api/orders/:id/status` - Update order status (product owner only)
- `GET /api/orders/confirm-payment` - Confirm Stripe payment

### Reviews

- `POST /api/reviews` - Create product review (authenticated)
- `GET /api/reviews/product/:id` - Get product reviews
- `PUT /api/reviews/:id` - Update review (owner only)
- `DELETE /api/reviews/:id` - Delete review (owner only)
- `GET /api/reviews/my` - Get user's reviews (authenticated)

### Messages

- `POST /api/messages` - Send private message (authenticated)
- `GET /api/messages/conversations` - Get all conversations (authenticated)
- `GET /api/messages/conversation/:user_id` - Get conversation with user (authenticated)
- `GET /api/messages/unread-count` - Get unread message count (authenticated)
- `PUT /api/messages/:id/read` - Mark message as read (authenticated)
- `DELETE /api/messages/:id` - Delete message (sender only)

### WebSocket

- `GET /ws?user_id=123&username=john` - WebSocket connection for real-time messaging

## Request/Response Examples

### Register User
```json
POST /api/auth/register
{
  "email": "user@example.com",
  "password": "password123",
  "first_name": "John",
  "last_name": "Doe"
}
```

### Create Product
```json
POST /api/products
Authorization: Bearer <jwt_token>
{
  "name": "iPhone 15",
  "description": "Latest iPhone model",
  "price": 999.99,
  "stock": 10,
  "category": "Electronics",
  "image_url": "https://example.com/iphone.jpg"
}
```

### Add to Cart
```json
POST /api/cart/add
Authorization: Bearer <jwt_token>
{
  "product_id": 1,
  "quantity": 2
}
```

### Create Order
```json
POST /api/orders
Authorization: Bearer <jwt_token>
{
  "shipping_address": "123 Main St, City, State 12345"
}
```

### Send Message
```json
POST /api/messages
Authorization: Bearer <jwt_token>
{
  "to_user_id": 2,
  "content": "Hello! I'm interested in your product."
}
```

## Database Schema

The application uses the following main entities:
- Users (with email confirmation)
- Products (with stock management)
- Orders and OrderItems
- Cart and CartItems
- Reviews (with ratings)
- Messages (for private communication)

## Security Features

- JWT token authentication
- Password hashing with bcrypt
- Email confirmation required for all operations
- Input validation and sanitization
- CORS configuration
- Rate limiting (can be added)

## Payment Integration

The application integrates with Stripe for payment processing:
- Creates checkout sessions for orders
- Handles payment confirmation
- Updates order status based on payment status

## Real-time Messaging

WebSocket implementation for real-time private messaging:
- STOMP-like protocol over WebSocket
- Private message delivery
- Message read status tracking
- Conversation management

## Development

### Running Tests
```bash
go test ./...
```

### Database Migrations
The application uses GORM auto-migration. Tables are created automatically when the application starts.

### Environment Variables
Make sure to set up all required environment variables in the `.env` file before running the application.

## Production Deployment

1. Set up a production PostgreSQL database
2. Configure production SMTP settings
3. Use production Stripe keys
4. Set a strong JWT secret
5. Configure proper CORS settings
6. Set up SSL/TLS certificates
7. Use a process manager like PM2 or systemd

## License

This project is licensed under the MIT License.
