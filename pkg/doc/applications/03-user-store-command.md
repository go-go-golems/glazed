---
Title: Cool Application - Managing User Records with Glazed
Slug: managing-user-records-with-glazed
Short: A comprehensive application example for managing user records using Glazed commands.
Topics:
- glaze
- commands
- application
- yaml
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: Application
---

DISCLAIMER: This was generated with o1-preview and was not validated yet.

## Overview

In this application example, we'll build a user management CLI tool using the `glazed` framework. The application will allow users to generate, list, and delete user records. Commands will be defined both programmatically and via YAML configurations, demonstrating the flexibility and power of `glazed`.

## Features

- **Generate Users**: Create a specified number of user records with customizable fields.
- **List Users**: Display all existing user records in a structured format.
- **Delete Users**: Remove user records by ID.
- **Configuration via YAML**: Define commands and their parameters using YAML files for easy customization.

## Project Structure

```
glazed-user-manager/
├── cmds/
│   ├── generate.go
│   ├── list.go
│   └── delete.go
├── configs/
│   └── commands.yaml
├── main.go
└── go.mod
```

## Step 1: Setting Up the Project

Initialize the project and module.

```bash
mkdir glazed-user-manager
cd glazed-user-manager
go mod init github.com/yourusername/glazed-user-manager
go get github.com/go-go-golems/glazed
```

## Step 2: Defining the User Model

Create a simple in-memory user store.

```go
// user_store.go
package main

import (
    "sync"
)

type User struct {
    ID    int
    Name  string
    Email string
}

type UserStore struct {
    sync.Mutex
    users map[int]User
    nextID int
}

func NewUserStore() *UserStore {
    return &UserStore{
        users: make(map[int]User),
        nextID: 1,
    }
}

func (s *UserStore) AddUser(name, email string) User {
    s.Lock()
    defer s.Unlock()
    user := User{
        ID:    s.nextID,
        Name:  name,
        Email: email,
    }
    s.users[s.nextID] = user
    s.nextID++
    return user
}

func (s *UserStore) ListUsers() []User {
    s.Lock()
    defer s.Unlock()
    users := make([]User, 0, len(s.users))
    for _, user := range s.users {
        users = append(users, user)
    }
    return users
}

func (s *UserStore) DeleteUser(id int) bool {
    s.Lock()
    defer s.Unlock()
    if _, exists := s.users[id]; exists {
        delete(s.users, id)
        return true
    }
    return false
}
```

## Step 3: Creating Commands

### 3.1 Generate Command

Generates user records.

```go
// cmds/generate.go
package cmds

import (
    "context"
    "strconv"

    "github.com/go-go-golems/glazed/pkg/cmds"
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/go-go-golems/glazed/pkg/cmds/parameters"
    "github.com/go-go-golems/glazed/pkg/middlewares"
    "github.com/go-go-golems/glazed/pkg/types"
)

type GenerateCommand struct {
    *cmds.CommandDescription
    store *UserStore
}

func NewGenerateCommand(store *UserStore) (*GenerateCommand, error) {
    return &GenerateCommand{
        CommandDescription: cmds.NewCommandDescription(
            "generate",
            cmds.WithShort("Generate user records"),
            cmds.WithFlags(
                fields.New(
                    "count",
                    fields.TypeInteger,
                    fields.WithHelp("Number of users to generate"),
                    fields.WithDefault(5),
                ),
                fields.New(
                    "verbose",
                    fields.TypeBool,
                    fields.WithHelp("Enable verbose output"),
                    fields.WithDefault(false),
                ),
            ),
        ),
        store: store,
    }, nil
}

func (c *GenerateCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *values.Values, gp middlewares.Processor) error {
    type GenerateSettings struct {
        Count   int  `glazed:"count"`
        Verbose bool `glazed:"verbose"`
    }

    settings := &GenerateSettings{}
    if err := parsedLayers.InitializeStruct("default", settings); err != nil {
        return err
    }

    for i := 0; i < settings.Count; i++ {
        user := c.store.AddUser("User-"+strconv.Itoa(c.store.nextID), "user"+strconv.Itoa(c.store.nextID)+"@example.com")
        row := types.NewRow(
            types.MRP("id", user.ID),
            types.MRP("name", user.Name),
            types.MRP("email", user.Email),
        )

        if settings.Verbose {
            row.Set("status", "Generated successfully")
        }

        if err := gp.AddRow(ctx, row); err != nil {
            return err
        }
    }

    return nil
}
```

### 3.2 List Command

Lists all user records.

```go
// cmds/list.go
package cmds

import (
    "context"

    "github.com/go-go-golems/glazed/pkg/cmds"
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/go-go-golems/glazed/pkg/cmds/parameters"
    "github.com/go-go-golems/glazed/pkg/middlewares"
    "github.com/go-go-golems/glazed/pkg/types"
)

type ListCommand struct {
    *cmds.CommandDescription
    store *UserStore
}

func NewListCommand(store *UserStore) (*ListCommand, error) {
    return &ListCommand{
        CommandDescription: cmds.NewCommandDescription(
            "list",
            cmds.WithShort("List all user records"),
            cmds.WithFlags(),
            cmds.WithArguments(),
        ),
        store: store,
    }, nil
}

func (c *ListCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *values.Values, gp middlewares.Processor) error {
    users := c.store.ListUsers()
    for _, user := range users {
        row := types.NewRow(
            types.MRP("id", user.ID),
            types.MRP("name", user.Name),
            types.MRP("email", user.Email),
        )
        if err := gp.AddRow(ctx, row); err != nil {
            return err
        }
    }
    return nil
}
```

### 3.3 Delete Command

Deletes a user record by ID.

```go
// cmds/delete.go
package cmds

import (
    "context"
    "strconv"

    "github.com/go-go-golems/glazed/pkg/cmds"
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/go-go-golems/glazed/pkg/cmds/parameters"
    "github.com/go-go-golems/glazed/pkg/middlewares"
    "github.com/go-go-golems/glazed/pkg/types"
)

type DeleteCommand struct {
    *cmds.CommandDescription
    store *UserStore
}

func NewDeleteCommand(store *UserStore) (*DeleteCommand, error) {
    return &DeleteCommand{
        CommandDescription: cmds.NewCommandDescription(
            "delete",
            cmds.WithShort("Delete a user record by ID"),
            cmds.WithFlags(),
            cmds.WithArguments(
                fields.New(
                    "id",
                    fields.TypeInteger,
                    fields.WithHelp("ID of the user to delete"),
                    fields.WithRequired(true),
                ),
            ),
        ),
        store: store,
    }, nil
}

func (c *DeleteCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *values.Values, gp middlewares.Processor) error {
    type DeleteSettings struct {
        ID int `glazed:"id"`
    }

    settings := &DeleteSettings{}
    if err := parsedLayers.InitializeStruct("default", settings); err != nil {
        return err
    }

    success := c.store.DeleteUser(settings.ID)
    status := "User deleted successfully"
    if !success {
        status = "User not found"
    }

    row := types.NewRow(
        types.MRP("id", settings.ID),
        types.MRP("status", status),
    )

    if err := gp.AddRow(ctx, row); err != nil {
        return err
    }

    return nil
}
```

## Step 4: Defining Commands in YAML

Create a `commands.yaml` file in the `configs` directory to define commands.

```yaml
---
name: generate
short: Generate user records
flags:
  - name: count
    type: int
    help: Number of users to generate
    default: 5
  - name: verbose
    type: bool
    help: Enable verbose output
    default: false
arguments: []
---
---
name: list
short: List all user records
flags: []
arguments: []
---
---
name: delete
short: Delete a user record by ID
flags: []
arguments:
  - name: id
    type: int
    help: ID of the user to delete
    required: true
---
```

## Step 5: Loading Commands from YAML

Modify your `main.go` to load commands from both YAML and programmatically.

```go
// main.go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/go-go-golems/glazed/pkg/cmds"
    "github.com/go-go-golems/glazed/pkg/cmds/loaders"
    "github.com/go-go-golems/glazed/pkg/cli"
    "github.com/go-go-golems/glazed/pkg/help"
    help_cmd "github.com/go-go-golems/glazed/pkg/help/cmd"
    "github.com/spf13/cobra"
)

func main() {
    store := NewUserStore()

    rootCmd := &cobra.Command{
        Use:   "user-manager",
        Short: "A CLI tool to manage user records",
    }

    // Initialize the help system
    helpSystem := help.NewHelpSystem()
    help_cmd.SetupCobraRootCommand(helpSystem, rootCmd)

    // Create a YAML loader
    yamlLoader := loaders.NewYAMLCommandLoader()

    // Load commands from YAML
    commands, err := yamlLoader.LoadCommands(os.DirFS("configs"), "commands.yaml", nil, nil)
    if err != nil {
        fmt.Println("Error loading commands:", err)
        os.Exit(1)
    }

    // Programmatically create commands
    generateCmd, err := NewGenerateCommand(store)
    if err != nil {
        fmt.Println("Error creating generate command:", err)
        os.Exit(1)
    }

    listCmd, err := NewListCommand(store)
    if err != nil {
        fmt.Println("Error creating list command:", err)
        os.Exit(1)
    }

    deleteCmd, err := NewDeleteCommand(store)
    if err != nil {
        fmt.Println("Error creating delete command:", err)
        os.Exit(1)
    }

    // Add programmatically created commands to the list
    commands = append(commands, generateCmd, listCmd, deleteCmd)

    // Add loaded commands to root
    for _, cmd := range commands {
        cobraCmd, err := cli.BuildCobraCommand(cmd)
        if err != nil {
            fmt.Println("Error building Cobra command:", err)
            os.Exit(1)
        }
        rootCmd.AddCommand(cobraCmd)
    }

    // Execute the root command
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

## Step 6: Running the Application

Build and run your application.

```bash
go build -o user-manager
./user-manager generate --count 3 --verbose
./user-manager list
./user-manager delete --id 2
./user-manager list
```

### Expected Output

```bash
# Generate users
./user-manager generate --count 3 --verbose
ID  Name     Email                  Status
1   User-1   user1@example.com      Generated successfully
2   User-2   user2@example.com      Generated successfully
3   User-3   user3@example.com      Generated successfully

# List users
./user-manager list
ID  Name     Email
1   User-1   user1@example.com
2   User-2   user2@example.com
3   User-3   user3@example.com

# Delete a user
./user-manager delete --id 2
ID  Status
2   User deleted successfully

# List users after deletion
./user-manager list
ID  Name     Email
1   User-1   user1@example.com
3   User-3   user3@example.com
```

## Step 7: Enhancing the Application

### Adding Persistence

Currently, the user records are stored in-memory. To persist data, integrate a database like SQLite.

```go
// Modify UserStore to use SQLite
// Implement methods to interact with the database
```

### Adding More Commands

Expand the application by adding more commands, such as updating user records or searching for users.

```go
// cmds/update.go
package cmds

// Implement UpdateCommand similar to GenerateCommand and DeleteCommand
```

### Improving YAML Configuration

Organize YAML configurations for better scalability, such as separating command definitions into multiple files.

```yaml
# configs/generate.yaml
---
name: generate
short: Generate user records
flags:
  - name: count
    type: int
    help: Number of users to generate
    default: 5
  - name: verbose
    type: bool
    help: Enable verbose output
    default: false
arguments: []
---
```

```yaml
# configs/list.yaml
---
name: list
short: List all user records
flags: []
arguments: []
---
```

```yaml
# configs/delete.yaml
---
name: delete
short: Delete a user record by ID
flags: []
arguments:
  - name: id
    type: int
    help: ID of the user to delete
    required: true
---
```

Update the loader to read all YAML files in the `configs` directory.

## Conclusion

This application demonstrates how to effectively use `glazed` to build a command-line tool with multiple commands, flags, and arguments. By leveraging both programmatic and YAML-based command definitions, you gain flexibility in managing your CLI's functionality. With further enhancements, such as database integration and additional commands, you can expand this tool to meet more complex user management needs.
