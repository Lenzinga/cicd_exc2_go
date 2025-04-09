# Go REST API with PostgreSQL

![Go CI](https://github.com/Lenzinga/cicd_exc2_go/actions/workflows/go.yml/badge.svg)

This is a basic REST API written in Go that performs CRUD operations on a `products` table using a PostgreSQL database.

## 🚀 Quick Start

```bash
docker-compose up -d        # Start PostgreSQL + Adminer
go run main.go              # Run the API (on :8010)
