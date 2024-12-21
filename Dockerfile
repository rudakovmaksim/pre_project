# создаем базовый образ
FROM golang:1.22

# копируем файлы зависимостей
COPY go.mod go.sum /usr/src/app/
# устанавливаем рабочую директорию
WORKDIR /usr/src/app

# скачиваем зависимости
RUN go mod download

# копируем остальные файлы с хоста в папку контейнера
ADD . /usr/src/app
COPY config.yaml /usr/src/app

# указываем команду, которая будет выполняться при запуске контейнера
CMD ["go", "run", "cmd/pre_project/main.go"]