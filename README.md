# GREEN-API test form (Go + required WASM)

Верстка формы выполнена по макету из задания:
- поля `idInstance`, `ApiTokenInstance`
- кнопки `getSettings`, `getStateInstance`, `sendMessage`, `sendFileByUrl`
- отдельная область вывода ответа

## Запуск

```bash
make run
```

Откройте `http://localhost:8080`.

## Что внутри

- `cmd/server` — HTTP сервер на Go.
- `internal/greenapi` — клиент вызова методов GREEN-API.
- `web/index.html` + `web/assets/app.css` — верстка формы под макет.
- `cmd/wasm` — Go WASM helper, который собирает и валидирует весь request для `/api/call`.

## Конфиг приложения

Поддерживается файл `configs/config.yml` и env-переменные.

Приложение работает в fail-fast режиме: если конфиг отсутствует или поля невалидны, сервер не стартует.

Поля в `configs/config.yml`:

- `server.host` — хост локального сервера (например `0.0.0.0` или `127.0.0.1`)
- `server.port` — порт локального сервера (например `8080`)
- `green_api.base_url` — base URL GREEN-API (например `https://api.green-api.com`)

Env-переопределения (имеют приоритет над файлом):

- `APP_CONFIG_PATH` — путь к конфигу (по умолчанию `configs/config.yml`)
- `APP_SERVER_HOST` — переопределяет `server.host`
- `APP_SERVER_PORT` — переопределяет `server.port`
- `GREEN_API_BASE_URL` — переопределяет `green_api.base_url`

Требования валидации:

- `server.host` — обязательно
- `server.port` — обязательно, диапазон `1..65535`
- `green_api.base_url` — обязательно, `http/https` URL

Без собранного WASM интерфейс блокирует кнопки вызова методов.

## Сборка

```bash
make build
```
