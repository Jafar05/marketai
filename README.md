# MarketAI - AI-powered Marketplace Cards

Сервис автоматической генерации карточек товаров для маркетплейсов с помощью AI.

## Архитектура

- **Frontend**: React + TypeScript + Ant Design
- **Auth Service**: Go + Echo + PostgreSQL + JWT
- **Cards Service**: Go + Echo + PostgreSQL + OpenAI API
- **Database**: PostgreSQL
- **Communication**: gRPC между сервисами

## Быстрый старт

### Предварительные требования

- Docker и Docker Compose
- Go 1.21+
- Node.js 18+
- OpenAI API ключ

### Запуск через Docker Compose

1. Клонируйте репозиторий:
```bash
git clone <repository-url>
cd MarketAI
```

2. Создайте файл `.env` в корне проекта:
```env
OPENAI_API_KEY=your_openai_api_key_here
JWT_SECRET=your_jwt_secret_here
```

3. Запустите все сервисы:
```bash
docker-compose up -d
```

4. Откройте браузер и перейдите на `http://localhost:3000`

### Локальная разработка

#### Backend (Auth Service)

```bash
cd auth
go mod download
go run cmd/auth/main.go
```

#### Backend (Cards Service)

```bash
cd cards
go mod download
go run cmd/cards/main.go
```

#### Frontend

```bash
cd marketai-front
npm install
npm run dev
```

## API Endpoints

### Auth Service (порт 8080)

- `POST /api/v1/register` - Регистрация пользователя
- `POST /api/v1/login` - Авторизация
- `POST /api/v1/validate` - Валидация токена

### Cards Service (порт 8081)

- `POST /api/v1/cards/generate` - Генерация карточки товара
- `GET /api/v1/cards/history` - История карточек пользователя
- `GET /api/v1/cards/:id` - Получение карточки по ID

## Структура проекта

```
MarketAI/
├── auth/                    # Auth микросервис
│   ├── cmd/auth/
│   ├── internal/
│   ├── migrations/
│   └── proto/
├── cards/                   # Cards микросервис
│   ├── cmd/cards/
│   ├── internal/
│   └── migrations/
├── marketai-front/          # React frontend
│   ├── src/
│   │   ├── pages/
│   │   ├── components/
│   │   ├── store/
│   │   └── api/
├── pkg/                     # Общие пакеты
├── configs/                 # Конфигурационные файлы
└── docker-compose.yml
```

## Использование

1. **Регистрация**: Создайте аккаунт на странице регистрации
2. **Авторизация**: Войдите в систему
3. **Генерация карточки**: 
   - Загрузите URL фотографии товара
   - Введите краткое описание
   - Нажмите "Сгенерировать карточку"
4. **История**: Просматривайте все созданные карточки

## Технологии

- **Backend**: Go, Echo, PostgreSQL, gRPC, JWT
- **Frontend**: React, TypeScript, Ant Design, Zustand, Axios
- **AI**: OpenAI GPT-3.5-turbo
- **DevOps**: Docker, Docker Compose

## Лицензия

MIT License