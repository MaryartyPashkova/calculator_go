# Используем официальный образ Go
FROM golang:1.24-alpine

# Рабочая директория внутри контейнера
WORKDIR /app

# Копируем go.mod и go.sum для загрузки зависимостей
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем бинарник
RUN go build -o /server cmd/server/main.go cmd/server/grpc_server.go

# Открываем порты
EXPOSE 8081
EXPOSE 50051

# Запускаем сервер
CMD ["/server"]
