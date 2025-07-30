---
Title: Database Integration Tutorial
Slug: database-tutorial
Short: Step-by-step guide for integrating with databases
SectionType: Tutorial
Topics:
- database
- integration
- sql
Commands:
- query
- connect
Flags:
- --database-url
- --timeout
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
Order: 1
---

# Database Integration Tutorial

This tutorial shows you how to integrate with databases using the query command.

## Prerequisites

- Database server running
- Valid connection credentials

## Step 1: Connect to Database

First, establish a connection:

```bash
query connect --database-url="postgresql://user:pass@localhost/db"
```

## Step 2: Execute Queries

Run your first query:

```bash
query --sql="SELECT * FROM users LIMIT 10"
```

## Advanced Usage

For complex queries, use the timeout flag:

```bash
query --sql="SELECT COUNT(*) FROM large_table" --timeout=30s
```
