# test-task-jwt

В данном проекте реализована часть сервиса аутентификации с использованием access и refresh токенами:
1. маршрут "GET /api/auth" - создание пары access и refresh токенов, обязателен параметр запроса guid.
2. маршрут "GET /api/refresh" - обновление пары access и refresh токенов, в запросе обязательны access токен и refresh токен.

# Особенности проекта:
* Refresh токен передается в cookie, access токен в заголовке "Authorization".
* Логи пишутся в файл data.log, для логирования использован пакет slog.
* Приложение и база данных поднимаются в докер контейнерах, приложение запускается только после запуска базы благодаря ./scripts/wait-for-it.sh.
* Реализован graceful shutdown приложения и базы данных.
* Использован стандарный пакет net/http для запуска сервера.


Для проверки работы авторизации был написан тест, в котором проверяется выдача токенов, получение ошибок:
```
$ go test -v
=== RUN   TestApp
--- PASS: TestApp (0.41s)
PASS
ok      AuthApp/test    0.671s
```

Для отправки warning email при обновлении пары refresh токена реализован интерфейс EmailSender:
```
2024/12/15 13:01:06 warning message succesfully sended to [some email]
2024/12/15 13:01:06 unknown ip get access to refresh operation: unknown ip: [127.0.0.1:49430], expected ip: [255.255.255.255], refresh id: [pO365Cknx9y3opl.x53JM]
```
