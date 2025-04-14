# Сервис для работы с ПВЗ

На ПВЗ несколько раз в день привозят новые товары, которые были заказаны через Авито. Прежде чем их отдавать заказчику, необходимо сначала проверить и внести информацию в базу. Из-за того, что ПВЗ много, а товаров ещё больше, нужно реализовать механизм, позволяющий в разрезе каждого ПВЗ увидеть, сколько раз в день к ним приезжали товары на приёмку и какие товары были получены.

## Описание задачи

[Подробное описание задачи](https://github.com/avito-tech/tech-internship/blob/main/Tech%20Internships/Backend/Backend-trainee-assignment-spring-2025/Backend-trainee-assignment-spring-2025.md)

## Стек технологий

- **Язык программирования:** Go
- **База данных:** PostgreSQL
- **Сбор метрик:** Prometheus
- **Для деплоя используется:** Docker

## Запуск проекта

### Клонируйте репозиторий:

```bash
git clone https://github.com/Maksim-512/pvz_service.git
cd pvz-service
```

### Сборка и запуск с использованием Docker:

```bash
docker-compose up --build
```

### Проект будет доступен по следующим адресам:

- API:        http://localhost:8080
- gRPC:       http://localhost:3000
- Prometheus: http://localhost:9000
- PostgreSQL: http://localhost:5433


## API
Сервис **полностью соответствует** спецификации [OpenAPI](https://github.com/avito-tech/tech-internship/blob/main/Tech%20Internships/Backend/Backend-trainee-assignment-spring-2025/swagger.yaml):


### 1) Авторизация и регистрация
#### POST /dummyLogin

Получение тестового токена для авторизации с ролью пользователя (employee или moderator).

Тело запроса:
```json
{
  "role": "moderator"
}
```

Ответ:
```json
{
  "token": "token"
}
```

#### POST /register

Регистрация нового пользователя с типом (client или moderator).

Тело запроса:
```json
{
    "email": "employee1@yandex.ru",
    "password": "Avito1234",
    "role": "employee"
}
```
Ответ:
```json
{
    "id": "uuid",
    "email": "employee1@yandex.ru",
    "role": "employee"
}
```

#### POST /login

Авторизация пользователя с почтой и паролем, возврат токена для авторизации.

Тело запроса:
```json
{
    "email": "employee1@yandex.ru",
    "password": "Avito1234"
}
```

Ответ:
```json
{
    "token": "token"
}
```

### 2. Работа с ПВЗ
#### POST /pvz

Создание нового ПВЗ (только для модераторов).

Тело запроса:
```json
{
    "city": "Москва"
}
```

Ответ:
```json
{
    "id": "uuid",
    "registrationDate": "datetime",
    "city": "Москва"
}
```

#### GET /pvz

Получение списка ПВЗ с фильтрацией по дате приёмки товаров и пагинацией.

Параметры запроса:
- startDate: Начальная дата диапазона
- endDate: Конечная дата диапазона
- page: Номер страницы
- limit: Размер страницы

Ответ:
```json
{
    "pvzList": [
    {
        "pvz": {
            "id": "uuid",
            "registrationDate": "datetime",
            "city": "Москва"
        },
        "receptions": [
            {
                "reception": {
                    "id": "uuid",
                    "dateTime": "datetime",
                    "pvzID": "uuid",
                    "status": "open"
                },
                "products": [
                    {
                        "id": "uuid",
                        "type": "обувь",
                        "receptionID": "uuid",
                        "dateTime": "datetime"
                    }
                ]
            }
        ]
    }
    ]
}
```

### 3. Работа с приёмками товаров
#### POST /receptions

Создание приёмки товаров для ПВЗ.

Тело запроса:
```json
{
    "pvzID": "uuid"
}
```

Ответ:
```json
{
    "id": "uuid",
    "dateTime": "datetime",
    "pvzID": "uuid",
    "status": "in_progress"
}
```

#### POST /products

Добавление товара в текущую приемку.

Тело запроса:
```json
    {
    "pvzID": "uuid",
    "type": "обувь"
    }
```

Ответ:
```json
{
    "id": "uuid",
    "type": "обувь",
    "receptionId": "uuid",
    "dateTime": "datetime"
}
```

#### POST pvz/{pvzID}/delete_last_product

Удаление товара из приёмки.

Тело запроса:
```json
{
  "pvzID": "uuid"
}
```

Ответ:
```json
{
  "id": "uuid",
  "type": "обувь",
  "receptionId": "uuid",
  "dateTime": "datetime"
}
```

#### POST /pvz/{pvzID}/close_last_reception

Закрытие приёмки товаров.

Тело запроса:
```json
{
  "pvzID": "uuid"
}
```

Ответ:
```json
{
    "id": "uuid",
    "dateTime": "datetime",
    "pvzID": "uuid",
    "status": "close"
}
```

## Тестирование

### Покрытие тестами
Проект покрыт unit-тестами с покрытием 90%.

### Интеграционные тесты

Разработан один интеграционный тест:
- Первым делом создает новый ПВЗ
- Добавляет новую приёмку заказов
- Добавляет 50 товаров в рамках текущей приёмки заказов
- Закрывает приёмку заказов

## Дополнительные функциональности

- gRPC-метод для получения всех ПВЗ
```bash
  grpcurl -plaintext localhost:3000 pvz.v1.PVZService.GetPVZList
```
- **Prometheus для сбора метрик**.
  - **Технические метрики**:
    - Количество запросов
    - Время ответа
  - **Бизнесовые метрики**:
      - Количество созданных ПВЗ
    - Количество созданных приёмок заказов
    - Количество добавленных товаров 

Для запуска:
```bash
curl http://localhost:9000/metrics
```
- Логирование для отслеживания операций (с использованием пакета slog).
