# Lichess CronJob for automatic tournament generation

Create a tournament and it automatically sends you a template message with the tournament link to an email address.

### Steps to raise the project

1. Create .env file

```sh
    cp .env.example .env
```

2. Get google application key in .env file

3. Get lichess token in .env file

4. Add source mail and destination mail in .env file

5. Write your period in linux crontab format in .env file

6. Execute next command

**For development**

```sh
    docker compose up --build
```

**For production**

```sh
    docker compose -f docker-compose.prod.yml up --build -d
```
