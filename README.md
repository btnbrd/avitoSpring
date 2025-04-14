

# Delivery Point Management Service (ПВЗ-сервис)

Сервис для управления пунктами выдачи заказов (ПВЗ), приёмками и товарами.  
Поддерживает аутентификацию, авторизацию, REST API.  
Реализованы роли `moderator` и `employee`.

---

##  Как запускать

### 1. Склонировать репозиторий

```bash
git clone https://github.com/btnbrd/avitoSpring
cd avitoSpring
```

### 2. Запустить сервис с помощью Docker Compose

```bash
docker-compose up --build
```

Сервис стартует на порте `8080`  
PostgreSQL — на `localhost:5432` (пользователь и пароль указаны в `config.yaml`)

---

##  Как тестировать

### 1. Юнит-тесты

```bash
go test ./... -v
```

### 2. Интеграционный тест



####  Тест приёмки ПВЗ

```bash
go test ./tests/integration -run ^TestPVZIntegration$
```

Этот тест выполняет полный сценарий приёмки товаров:

- Создаёт ПВЗ
- Начинает приёмку
- Добавляет 50 товаров
- Завершает приёмку

---

####  Тест регистрации модератора

```bash
go test ./tests/integration -run ^TestModeratorRegisterLoginAndCreatePVZ$
```




Этот тест проверяет сценарий работы модератора:

- Регистрирует пользователя с ролью *модератор*
- Входит под этим пользователем
- Создаёт ПВЗ

---

# Работа с миграциями

Миграции управляются через отдельный контейнер migrate, который запускает изменения в базе данных PostgreSQL из папки ./migrations

---



##  Покрытие тестами

### Юнит-тесты
- Хендлеры (`handlers`)
- Бизнес-логика (`services`)
- Мидлвар для авторизации и логгирования(`middleware`)
- PGSQL-репозитории  (`storage`)

### Запуск Unit-тестов
```bash
go test ./... -v -coverprofile=coverage.out   
go tool cover -func=coverage.out      
```

Итоговое покрытие 88%

### Интеграционные тест
- Проверяет полный happy path добавления и закрытия приёмки
- Проверяет регистрацию и логин модератора и создание пвз

---

# API

Вот как можно оформить **пользовательское API-руководство** по работе с запросами на основе твоего OpenAPI-описания. Такой документ можно положить в `README.md` или в отдельный `docs/api_usage.md`.

---

#  Руководство по работе с API

##  Авторизация

Все защищённые запросы требуют **JWT-токен**. Получить токен можно через:

###  Быстрый вход (для отладки)
```http
POST /dummyLogin
Content-Type: application/json

{
  "role": "moderator"
}
```

Ответ:
```json
"eyJhbGciOiJIUzI1..."
```

---

###  Регистрация
```http
POST /register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "strong-password",
  "role": "employee"
}
```

---

###  Вход
```http
POST /login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "strong-password"
}
```

Ответ:
```json
"eyJhbGciOiJIUzI1..."
```

---

##  Работа с ПВЗ

### ➕ Создать ПВЗ (модератор)
```http
POST /pvz
Authorization: Bearer <token>
Content-Type: application/json

{
  "city": "Москва"
}
```

---

###  Получить список ПВЗ (с фильтрацией и пагинацией)
```http
GET /pvz?startDate=2024-01-01T00:00:00Z&endDate=2024-12-31T23:59:59Z&page=1&limit=10
Authorization: Bearer <token>
```

---

##  Приемки

###  Начать приемку (сотрудник)
```http
POST /receptions
Authorization: Bearer <token>
Content-Type: application/json

{
  "pvzId": "<uuid>"
}
```

---

###  Закрыть последнюю приемку
```http
POST /pvz/<pvzId>/close_last_reception
Authorization: Bearer <token>
```

---

##  Товары

###  Добавить товар в активную приемку
```http
POST /products
Authorization: Bearer <token>
Content-Type: application/json

{
  "type": "одежда",
  "pvzId": "<uuid>"
}
```

---

###  Удалить последний товар из приемки
```http
POST /pvz/<pvzId>/delete_last_product
Authorization: Bearer <token>
```

---

##  Примечания

- **Модератор** — может создавать ПВЗ.
- **Сотрудник** — может начинать и завершать приемки, добавлять и удалять товары.
- Все действия, кроме `/register`, `/login`, `/dummyLogin` требуют `Authorization: Bearer <JWT>`.

---

# Возникшие трудности
1. В тексте задания указано что у сущности приемка должно быть "Товары, которые были приняты в рамках данной приёмки". Однако в API я это не обнаружил и просто сделал One-To-Many связь приемка-заказы.
2. В процессе разработки текущая архитектура "послойной" организации перестала казаться оптимальной. Наверное стоило выбрать разделение по сущностям.
3. В текущей реализации secret-key для jwt-токенов храниться в файле .ENV, но для реального использования необходимо скрыть его, не стал добавлять в gitignore, чтобы была возможность запустить. Лучше создавать отдельно в случае развертывания сервиса.

---
# Результаты
Были выполнены все необходимые требования(требование по нагрузке не протестировано, расчет на то что проходит за счет постгресс и библиотек), из дополнительных: была добавлена авторизация через почту-пароль, было настроено логирование через zap


