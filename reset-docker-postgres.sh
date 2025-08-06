#!/bin/bash
docker-compose down
docker volume rm first-server-project_postgres_data
docker-compose up -d

# Fiz essa porra de script porque eu estava rodando isso trezentas vezes :)