
# Respite REST API Library

Respite is a Go library that auto‑generates RESTful APIs from your domain models. It integrates seamlessly with Keycloak or other IDPs for authentication, supports role‑based access control (RBAC), and is fully configurable via environment variables.

## Features

- Automatic REST API Generation: Dynamically exposes CRUD endpoints derived from your domain models, reducing boilerplate code.
- Keycloak Integration: Provides seamless authentication via Keycloak or custom Identity Providers (IDPs), ensuring secure access.
- Role-Based Access Control (RBAC): Define granular permissions for various roles, enhancing security and access management.
- Environment-Driven Configuration: Leverage environment variables for all configurations, promoting flexibility and ease of deployment.
- Modular Design: Built with extensibility in mind, allowing easy customization and integration into various projects.

## Installation

```bash
go get github.com/dzahariev/respite
```

## Getting Started

1. Set your env vars (e.g., using a `.env`).
2. Run the migrations to create DB tables.
3. Launch the API server with `go run main.go`.
4. Send requests to `http://localhost:8800/api/{resource}`.

## Usage Example

```
package main

import (
	"context"
	"log"

	"github.com/dzahariev/respite/api"
	"github.com/dzahariev/respite/auth"
	"github.com/dzahariev/respite/basemodel"
	"github.com/dzahariev/respite/cfg"
	"github.com/dzahariev/solei/model"
	"github.com/sethvargo/go-envconfig"
)

func main() {
	ctx := context.Background()

	// Load configuration from environment variables
	var loggerCfg cfg.Logger
	if err := envconfig.Process(ctx, &loggerCfg); err != nil {
		log.Fatal(err)
	}

	var databaseCfg cfg.DataBase
	if err := envconfig.Process(ctx, &databaseCfg); err != nil {
		log.Fatal(err)
	}

	var serverCfg cfg.Server
	if err := envconfig.Process(ctx, &serverCfg); err != nil {
		log.Fatal(err)
	}

	var authCfg cfg.Keycloak
	if err := envconfig.Process(ctx, &authCfg); err != nil {
		log.Fatal(err)
	}

	// Define your domain model objects
	objects := []basemodel.Object{
		&model.Category{},
		&model.Meal{},
		&model.Order{},
		&model.OrderItem{},
	}

	// Map roles to permissions for access control
	rolesToPermissions := map[string][]string{
		"Customer": {
			"user.read",
			"address.read",
			"address.write",
			"meal.read",
			"category.read",
			"order.read",
			"order.write",
			"orderitem.read",
			"orderitem.write",
		},
		"Chef": {
			"user.read",
			"orderitem.global",
			"orderitem.read",
			"orderitem.write",
		},
		"Courier": {
			"user.read",
			"order.global",
			"order.read",
			"order.write",
		},
		"Owner": {
			"user.read",
			"meal.read",
			"meal.write",
			"category.read",
			"category.write",
			"order.global",
			"order.read",
			"orderitem.global",
			"orderitem.read",
		},
	}

	// Initialize authentication client
	authClient := auth.NewClient(authCfg)

	// Create the API server instance
	server, err := api.NewServer(serverCfg, loggerCfg, databaseCfg, objects, authClient, rolesToPermissions)
	if err != nil {
		log.Fatal(err)
	}

	// Run the server
	server.Run()
}
```

### Configuration

### Configuration

You can configure Respite via environment variables:

| Env Var                   | Description                                   |
|---------------------------|-----------------------------------------------|
| `LOG_LEVEL`, `LOG_FORMAT` | Logger behavior                              |
| `DB_HOST`, `DB_PORT`, …   | PostgreSQL connection info                   |
| `AUTH_URL`, `AUTH_REALM`, … | Keycloak / IDP config                        |
| `SERVER_PORT`, `SERVER_API_PATH`, … | HTTP server settings                          |


### Example Environment Variables

```
# Authentication
AUTH_URL=http://keycloak:8086
AUTH_REALM=myapp
AUTH_CLIENT_ID=myapp-backend-client
AUTH_CLIENT_SECRET=785df81b-170e-4900-8f2d-de46d801606d

# Keycloak Admin (for initial setup)
KEYCLOAK_ADMIN=admin
KEYCLOAK_ADMIN_PASSWORD=admin

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=postgres
DB_USER=postgres
DB_PASSWORD=postgres

# Logger
LOG_LEVEL=debug
LOG_FORMAT=text

# Server
SERVER_API_PATH=api
SERVER_PORT=8800
SERVER_WRITE_TIMEOUT=15s
SERVER_READ_TIMEOUT=15s
SERVER_IDLE_TIMEOUT=60s
SERVER_DEADLINE_ON_INTERRUPT=15s
SERVER_MIN_PAGE_SIZE=10
SERVER_MAX_PAGE_SIZE=500
```
Ensure that sensitive information like AUTH_CLIENT_SECRET and DB_PASSWORD are not hardcoded in public repositories. Consider using .env files or secret management tools for local development.

### Database entries
The database tables should be created following the pattern described below. The created and updated timestamps are filled and updated using DB triggers and the relation between objects is expressed with corresponding ID that points to master record. 

The following SQL statements create tables with triggers for `created_at` and `updated_at` timestamps:
```
-- Function that sets created_at field
CREATE OR REPLACE FUNCTION set_created_at()
    RETURNS TRIGGER
AS
$$
BEGIN
    NEW.created_at = NOW();
    RETURN NEW;
END
$$ LANGUAGE plpgsql;

-- Function that sets updated_at field
CREATE OR REPLACE FUNCTION set_updated_at()
    RETURNS TRIGGER
AS
$$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END
$$ LANGUAGE plpgsql;

...

-- Table for categories
CREATE TABLE categories(
    id uuid PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    name VARCHAR(1024) NOT NULL
);

-- Trigger that sets created_at on categories
CREATE TRIGGER set_created_at_on_categories
    BEFORE INSERT 
    ON categories
    FOR EACH ROW
EXECUTE FUNCTION set_created_at();

-- Trigger that sets created_at on categories
CREATE TRIGGER set_updated_at_on_categories
    BEFORE INSERT OR UPDATE 
    ON categories
    FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- Table for meals
CREATE TABLE meals(
    id uuid PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    name VARCHAR(1024) NOT NULL,
    description VARCHAR(1024) NOT NULL,
    cost real NOT NULL,
    category_id uuid NOT NULL,
    CONSTRAINT fk_category FOREIGN KEY(category_id) REFERENCES categories(id) ON DELETE CASCADE
);

-- Trigger that sets created_at on meals
CREATE TRIGGER set_created_at_on_meals
    BEFORE INSERT 
    ON meals
    FOR EACH ROW
EXECUTE FUNCTION set_created_at();

-- Trigger that sets created_at on meals
CREATE TRIGGER set_updated_at_on_meals
    BEFORE INSERT OR UPDATE 
    ON meals
    FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

```

The owning resources that are linked to User (ownership if the user) should have ID to the user relation. Example here are orders:

```
-- Table for orders
CREATE TABLE orders(
    id uuid PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    price real NOT NULL,
    status VARCHAR(1024) NOT NULL,
    user_id uuid NOT NULL,
    CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

User entity in DataBase should be created with DDL, but the model object that uses this table is provided by the library:
```
-- Table for users
CREATE TABLE users(
    id uuid PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    prefered_user_name VARCHAR(1024) NOT NULL,
    given_name VARCHAR(1024) NOT NULL,
    family_name VARCHAR(1024) NOT NULL,
    email VARCHAR(1024) NOT NULL
);
-- Trigger that sets created_at on users
CREATE TRIGGER set_created_at_on_users
    BEFORE INSERT 
    ON users
    FOR EACH ROW
EXECUTE FUNCTION set_created_at();

-- Trigger that sets created_at on users
CREATE TRIGGER set_updated_at_on_users
    BEFORE INSERT OR UPDATE 
    ON users
    FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
```

### Domain Model

Implement the basemodel.Object interface for your domain entities to have them exposed as REST endpoints automatically. The both entities from DB schema that have a relation between them are created like this:

```
package model

import (
	"context"
	"fmt"
	"html"
	"strings"

	"github.com/dzahariev/respite/basemodel"
)

// Category
type Category struct {
	basemodel.Base
	Name string `json:"name"`
}

func (t *Category) ResourceName() string {
	return "category"
}

func (t *Category) IsGlobal() bool {
	return true
}

// Validate checks structure consistency
func (t *Category) Validate(ctx context.Context) error {
	if t.Name == "" {
		return fmt.Errorf("required Name")
	}

	return nil
}

func (t *Category) Prepare(ctx context.Context) error {
	err := t.BasePrepare(ctx)
	if err != nil {
		return err
	}

	t.Name = html.EscapeString(strings.TrimSpace(t.Name))

	return nil
}
```
```
package model

import (
	"context"
	"fmt"
	"html"
	"strings"

	"github.com/dzahariev/respite/basemodel"
	"github.com/gofrs/uuid/v5"
)

// Meal
type Meal struct {
	basemodel.Base
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Cost        float32   `json:"cost"`
	CategoryID  uuid.UUID `json:"category_id"`
	Category    Category
}

func (t *Meal) ResourceName() string {
	return "meal"
}

func (t *Meal) Preloads() []string {
	return []string{"Category"}
}

// Validate checks structure consistency
func (t *Meal) Validate(ctx context.Context) error {
	if t.Name == "" {
		return fmt.Errorf("required Name")
	}
	if t.Description == "" {
		return fmt.Errorf("required Description")
	}
	if t.Cost == 0 {
		return fmt.Errorf("required Cost")
	}

	return nil
}

func (t *Meal) Prepare(ctx context.Context) error {
	err := t.BasePrepare(ctx)
	if err != nil {
		return err
	}

	t.Name = html.EscapeString(strings.TrimSpace(t.Name))
	t.Description = html.EscapeString(strings.TrimSpace(t.Description))

	return nil
}

```

### Roles and Permissions

Implement Role-Based Access Control (RBAC) by defining roles and assigning specific permissions to control access to various operations and resources within the API.

### API Server Initialization

```
func NewServer(
    serverCfg cfg.Server,
    loggerCfg cfg.Logger,
    databaseCfg cfg.DataBase,
    objects []basemodel.Object,
    authClient auth.Client,
    rolesToPermissions map[string][]string,
) (*Server, error)
```

Creates and configures the REST API server with all components wired.

### Running the Server

Invoke `server.Run()` to start the HTTP server. Note that this is a blocking call and the server will continue running until manually stopped.

### Full example

Still under development, will add link later.

## License

Distributed under the Apache License, Version 2.0. See LICENSE for more information.
