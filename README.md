### Калькулятор на Go

### 1. Описание

Сервис принимает JSON-массив инструкций и вычисляет значения переменных параллельно, учитывая зависимости между ними.  
Реализованы два типа операций:

- `calc` — вычисление арифметического выражения (поддерживает `+`, `-`, `*`)
- `print` — вывод значения указанной переменной

### 2. Основные компоненты

#### 2.1. Ядро: `CalculatorService`

##### Метод `Run(...)`

Принимает список инструкций и возвращает результаты для указанных `print` переменных.

```go
func (s *CalculatorService) Run(ctx context.Context, instructions []Instruction) ([]ResultItem, error)
```

##### Что делает:

1. Разбирает инструкции на `calc` и `print`
2. Создаёт задачи (`func() (int64, error)`)
3. Запускает их параллельно через `goroutine`
4. Собирает только те переменные, что указаны в `print`

##### Защита от ошибок:

- Нельзя записать значение дважды → `"variable x already assigned"`
- Нельзя читать неизвестную переменную → `"undefined variable: x"`

##### Используется:

- `sync.WaitGroup` — для ожидания завершения всех задач
- `sync.Mutex` — для безопасного доступа к состоянию
- `map[string]func() (int64, error)` — граф зависимостей

#### 2.2. Парсинг значений: `resolveValue(...)`

Позволяет определить значение из:

- `int64` — если это число
- `string` — если это имя другой переменной

Выполняет рекурсивное вычисление при необходимости.  
Использует один и тот же `tasks` маппинг, чтобы поддерживать граф зависимостей.

### 3. HTTP API

#### POST `/calculate`

Принимает массив инструкций в формате JSON.

Пример:

```json
[
  { "type": "calc", "op": "+", "var": "x", "left": 1, "right": 2 },
  { "type": "print", "var": "x" }
]
```

Ответ:

```json
{
  "items": [{ "var": "x", "value": 3 }]
}
```

Для документации используется **Swagger UI** (`api/swagger.yaml`)  
Запуск: `http://localhost:8090/?url=/v1/swagger.yaml`

### 4. gRPC API

#### Реализация:

- Сервис: `calculator.CalculatorService/Calculate`
- Порт: `50051`
- Прото: `proto/calculator.proto`
- Поддержка: `oneof left_type { int64 left_int; string left_var }`

Пример запроса:

```bash
grpcurl -plaintext -d '{
  "instructions": [
    {"type":"calc","op":"+","var":"x","left_int":1,"right_int":2},
    {"type":"print","var":"x"}
  ]
}' localhost:50051 calculator.CalculatorService/Calculate
```

### 5. Тестирование

#### Юнит-тесты:

- `TestCalculator_Run_SimpleCalc`
- `TestCalculator_Run_VariableDependency`
- `TestCalculator_Run_ReassignVariable`
- `TestCalculator_Run_UndefinedVariable`
- `TestCalculator_Run_ComplexGraph`

#### Как запустить:

```bash
go test -v ./internal/service/
```

#### Покрытие:

```bash
go test -coverprofile=coverage.out ./internal/service/
go tool cover -html=coverage.out
```

### 6. Docker

#### Файлы:

- `Dockerfile` — сборка Go сервиса
- `docker-compose.yml` — запуск HTTP + gRPC + Swagger UI

#### Команда запуска:

```bash
docker-compose up --build
```

#### API:

- HTTP: `http://localhost:8081/calculate`
- gRPC: `localhost:50051`
- Swagger: `http://localhost:8090`
