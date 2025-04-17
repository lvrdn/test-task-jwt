# test-task-jwt

В данном проекте реализована часть сервиса аутентификации с использованием access и refresh токенами:
1. маршрут "GET /api/auth" - создание пары access и refresh токенов, обязателен параметр запроса guid.
2. маршрут "GET /api/refresh" - обновление пары access и refresh токенов, в запросе обязательны access токен и refresh токен.

# Особенности проекта:
* Access токен - типа JWT, алгоритм SHA512.
* Refresh токен - произвальная строка, хранится в базе данных в виде bcrypt хеша, передается закодированным base64.
* Refresh токен передается в cookie, access токен в заголовке "Authorization".
* Логи пишутся в файл data.log, для логирования использован пакет slog. Часть логов дублируется в консоль для удобства.
* Приложение и база данных поднимаются в докер контейнерах, приложение запускается только после запуска базы благодаря ./scripts/wait-for-it.sh.
* Реализован graceful shutdown приложения и базы данных.
* Использован стандарный пакет net/http для запуска сервера, регистрации маршрутов и т.д.
* Работа с PostgreSQL осуществляется через интерфейс Storage, что позволяет в дальнейшем применить другую базу данных.
* Для отправки warning email при обновлении пары refresh токена реализован интерфейс EmailSender.

# Запуск:
Для запуска проекта необходимо убедится, что на Вашей машине запущен Docker Engine.
Проверить свободен ли на Вашей машине порт 8086, если нет, то есть возможность запуститься на другом порту, указав его в файле конфигурации app.env.
Также в файле конфигурации можно задать ключ подписи для access токена (ACCESS_KEY), время жизни access токена в минутах (ACCESS_EXP_MINUTES), время жизни refresh токена в минутах (REFRESH_EXP_MINUTES).

Для запуска в текущей директории проекта выполните команду:
```
make app
```

# Тесты (./test):
Для проверки работоспособности приложения были написаны тесты, позволяющие проверить работоспособность приложения и краевые случаи.
Был переработан пакет app для запуска httptest.Server.
Для запуска тестов необходимо предварительно убедиться, что запущен Docker Engine и свободен порт 5432 (для запуска контейнера с PostgreSQL), и выполнить команду:
```
make test
```
После успешного завершения тестирования compose-stack и контейнер с PostgreSQL автоматически удаляться.
Результат выполнения:
```
docker-compose --profile disabled -f test/init_db/docker-compose.yml up -d; 
[+] Running 2/2
 ✔ Network init_db_default  Created                                                                                                                                                                                           0.1s 
 ✔ Container test_postgres  Started                                                                                                                                                                                           0.5s 
sleep 2;
go test ./test;
ok      app/test        0.814s
docker stop test_postgres;
test_postgres
docker-compose -f test/init_db/docker-compose.yml -p init_db down
[+] Running 2/2
 ✔ Container test_postgres  Removed                                                                                                                                                                                           0.0s 
 ✔ Network init_db_default  Removed
```

