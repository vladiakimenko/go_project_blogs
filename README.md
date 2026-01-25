# Blog API

API блог-платформы. JWT аутентификация. Круд по блогам и комментам

## Доступ
- Апи: `http://localhost:8080`
- Адимнка: `http://localhost:8081`
- Свагер: `http://localhost:8082`


## Сервисы
- **Postgres** — база данных
- **Redis** — троттлинг
- **Swagger UI** — документация
- **Adminer** — админка БД

## Эндпоинты

### Auth
- `POST /api/register` — регистрация пользователя
- `POST /api/login` — вход пользователя
- `POST /api/refresh` — обновление токена

### Users
- `GET /api/users/{userID}` — получение профиля пользователя (auth)

### Posts
- `GET /api/posts` — список постов с пагинацией
- `POST /api/posts` — создание поста (auth)
- `GET /api/posts/{postID}` — получение поста
- `PUT /api/posts/{postID}` — обновление поста (auth)
- `DELETE /api/posts/{postID}` — удаление поста (auth)

### Comments
- `GET /api/posts/{postID}/comments` — список комментариев с пагинацией
- `POST /api/posts/{postID}/comments` — создание комментария (auth)
- `PUT /api/posts/{postID}/comments/{commentID}` — обновление комментария (auth)
- `DELETE /api/posts/{postID}/comments/{commentID}` — удаление комментария (auth)


## Пагинация
- Параметры: `limit` и `offset`

## Миграции
- SQL скрипт: `migrations/001_init_schema.sql`
- Автоматически применяются при поднятии контейнера Postgres через Docker

## Запуск
```bash
docker-compose up -d
go run cmd/api/main.go
```
