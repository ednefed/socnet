# Разделение монолита на сервисы
## Протокол взаимодействия
Максимально простой в силу низкой связанности сервисов. Новый сервис предоставляет тот же функционал на эндпоинте /api/v2/dialog/:id. Монолит, дабы не дублировать код по обработке входящих запросов на /dialog/:id, при обращении на эти эндпоинты выполняется проксирование запроса на новый сервис диалогов. При этом:
- путь из запроса меняется с /dialog/:id (монолит) на /api/v2/dialog/:id
- к заголовку исходного запроса (если вдруг таковой не пришёл с "фронта") добавляется X-Request-ID заголовок для организации сквозного логирования события на обоих сервисах
Клиенту возвращается ответ без изменений.

## Пример сквозного логирования за счёт X-Request-ID
Три  запроса:
- cURL с созданием сообщения на старый сервис
- cURL с запросом сообщений диалога на старый сервис
- cURL с запросом сообщений диалога напрямую на новый сервис
```
ednefed@nfd-vm-ubuntu:~/works$ curl -H "Authorization: Bearer $token" -X POST http://localhost:8080/dialog/2 -d '{"message": "hello"}'
{"message":"Message sent"}
ednefed@nfd-vm-ubuntu:~/works$ curl -H "Authorization: Bearer $token" -X GET http://localhost:8080/dialog/2
[{"from_user_id":1001111,"to_user_id":2,"message":"hello","created_at":"2024-10-01T19:25:30Z"}]
ednefed@nfd-vm-ubuntu:~/works$ curl -H "Authorization: Bearer $token" -X GET http://localhost:8081/api/v2/dialog/2
[{"from_user_id":1001111,"to_user_id":2,"message":"hello","created_at":"2024-10-01T19:25:30Z"}]
ednefed@nfd-vm-ubuntu:~/works$
```
Восемь строк в логах):
```
# сообщение в старом сервсие по проксировании POST /dialog/2 на новый сервис, request-id
social_network-api-1             | 2024/10/01 19:25:30 request-id: b0d3e4f9-6f63-4567-8859-44f5488fcae4, proxy: POST /dialog/2 -> dialog_api:8080
# лог от GIN о завершении работы хэндлера POST /api/v2/dialog/2 новым сервисом -- такой вот логгер внутри GIN быстрый, что успел быстрее логгера основного (см. [dialog-api/handlers.go](../dialog-api/handlers.go) строки 55 и 56)
social_network-dialog_api-1      | [GIN] 2024/10/01 - 19:25:30 | 200 |     898.207µs |      172.19.0.1 | POST     "/api/v2/dialog/2"
# сообщение о приёме и обработке запроса новым сервисом, request-id тот же, что и на первом сервисе
social_network-dialog_api-1      | 2024/10/01 19:25:30 request-id: b0d3e4f9-6f63-4567-8859-44f5488fcae4, createDialogMessage: Message sent
# лог от GIN о завершении работы хэндлера POST /dialog/2 старым сервисом
social_network-api-1             | [GIN] 2024/10/01 - 19:25:30 | 200 |      1.9586ms |      172.19.0.1 | POST     "/dialog/2"
# сообщение в старом сервсие по проксировании GET /dialog/2 на новый сервис, request-id
social_network-api-1             | 2024/10/01 19:25:43 request-id: 89292270-13fc-47dd-bb0a-db6bc157409a, proxy: GET /dialog/2 -> dialog_api:8080
# быстрейший лог от GIN о завершении работы хэндлера GET /api/v2/dialog/2
social_network-dialog_api-1      | [GIN] 2024/10/01 - 19:25:43 | 200 |     568.508µs |      172.19.0.1 | GET      "/api/v2/dialog/2"
# сообщение о приёме и обработке запроса новым сервисом, request-id тот же, что и на первом сервисе
social_network-dialog_api-1      | 2024/10/01 19:25:43 request-id: 89292270-13fc-47dd-bb0a-db6bc157409a, getDialogMessages: Messages received
# лог от GIN о завершении работы хэндлера GET /dialog/2 старым сервисом
social_network-api-1             | [GIN] 2024/10/01 - 19:25:43 | 200 |     957.838µs |      172.19.0.1 | GET      "/dialog/2"
# быстрейший лог от GIN о завершении работы хэндлера GET /api/v2/dialog/2
social_network-dialog_api-1      | [GIN] 2024/10/01 - 19:25:50 | 200 |     575.042µs |      172.19.0.1 | GET      "/api/v2/dialog/2"
# сообщение о приёме и обработке запроса новым сервисом, request-id пустой, т.к. никто его не задал на клиенте (мне лень)
social_network-dialog_api-1      | 2024/10/01 19:25:50 request-id: , getDialogMessages: Messages received
```

## Поддержка старых клиентов
Запрос к старому сервису перенаправляется в неизменном (ну или почти, загловок X-Request-ID не в счёт, т.к. нет фронтенда) виде на новый сервис.

## Новые клиенты верно ходят через новый API
Функционал в отдельном сервисе реализован переносом кода из старого сервиса, меняется только эндпоинт (/api/v2/dialog/:id против /dialog/:id).
