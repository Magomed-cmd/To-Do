# Тестовые сценарии (unit)

Ниже — подробный перечень покрытых тестами сценариев по сервисам/пакетам.

## Общие пакеты
- `pkg/events` — `TaskEvent.IsZero`: событие считается пустым без ID или type; валидно при заполненных полях.

## Analytics-service
- Сервис (`internal/service`):
  - `TrackTaskEvent`: отказ при пустом userID; подстановка `OccurredAt` если не задан; корректные дельты для `created/completed/deleted`; ошибка для `UNSPECIFIED` типа.
  - `GetDailyMetrics`: отказ при пустом userID; подстановка текущей даты если не указана.
  - `buildDelta`: табличные проверки маппинга типов событий в дельты и ошибка для неподдерживаемых.
- Адаптер БД (`internal/adapters/database`):
  - Upsert метрик: проверка аргументов и SQL (increments + `GREATEST` для total).
  - Чтение метрик: возврат найденной строки; возврат пустой метрики при `pgx.ErrNoRows`.
  - `normalizeDate`: обрезка времени до полуночи UTC и подстановка текущей даты для zero-value.
- HTTP `/metrics/daily` (`internal/adapters/http/metrics`):
  - Валидация userId/date (400 на нечисловой id и неверный формат даты).
  - Проброс ошибки сервиса (500).
  - Успешный ответ с корректной передачей параметров в сервис.
- HTTP router (`internal/infrastructure/app`):
  - Ошибка при отсутствии `AnalyticsService`.
  - `/health` отвечает 200 на GET/HEAD с именем сервиса.
- Config (`internal/infrastructure/config`):
  - Ошибка валидации при пустых обязательных env.
  - Успешная загрузка: формирование `PostgresURL`, чтение max/min conns, max lifetime.
  - Парс-хелперы: неверные значения дают 0; `valueOrDefault` возвращает fallback.

## Task-service
- DTO (`internal/dto`):
  - `CreateTaskRequest.ToInput`: тримминг описания, дефолты статуса/приоритета.
  - `UpdateTaskRequest.ToInput`: парс/нормализация статуса/приоритета, проброс clear-флагов, тримминг описания.
  - `TaskFilterRequest.ToFilter`: парс status/priority, тримминг search, clamp limit/offset, перенос дат/категории.
  - Респонсы: сбор категорий/комментариев/тасков; парсеры `parseStatus/parsePriority`, `normalizePtr`, `clampLimit/Offset`; `toStatusOrDefault`/`toPriorityOrDefault` дефолтят на pending/medium.
- HTTP common (`internal/adapters/http/common`):
  - `WriteValidationError` → 400; `WriteDomainError` маппит доменные ошибки (404/403/400/500); `ExtractBearerToken` достает токен с префиксом.
- HTTP middleware (`internal/adapters/http/middleware`):
  - JWT: отсутствие/невалидный заголовок → 401; неуспешный парс → 401; успешный парс — сохранение клеймов в контекст; `CurrentUser` достает клеймы.
- Auth (`internal/adapters/auth/jwt`):
  - Генерация/парсинг access и refresh токенов с проверкой экспирации; контроль, что неверный секрет ломает парсинг.
- gRPC клиенты:
  - `analyticsgrpc`: валидация конфигурации (пустой адрес → ошибка); дефолтный timeout при <=0; форвардинг события с корректными полями; `InvalidArgument` от сервера игнорируется, остальные ошибки возвращаются.
  - `usergrpc`: валидация конфигурации; дефолтный timeout; успешное получение пользователя с маппингом в `UserInfo`; преобразование gRPC ошибок: NotFound → `ErrUnknownUser`, PermissionDenied → `ErrForbiddenTaskAccess`, InvalidArgument → `ErrValidationFailed` с сообщением; прочие ошибки пробрасываются.
- Сервис задач (`internal/service`):
  - `CreateTask`: валидация заголовка/статуса/приоритета, тримминг полей, проверка категории, создание, публикация analytics события и уведомления (проверка полезной нагрузки).
  - Валидационные кейсы `CreateTask` (пустой title, неверный статус/приоритет).
  - `UpdateTask`: обновление полей, тримминг, валидация статуса/приоритета, установка/очистка due date и категории.
  - `UpdateTaskStatus`: валидация статуса, изменение, публикация analytics + notification для completed.
  - `DeleteTask`: soft delete, публикация analytics + notification.
  - `ListTasks`: дефолты limit=20/offset=0.
  - Категории: ошибка при пустом имени; успешное создание; `DeleteCategory`/`ListCategories` покрыты через репозиторий-стаб.
  - Комментарии: ошибка на пустой контент; успешное добавление.
  - `ensureUser`: отказ для неактивного пользователя (Forbidden).

## User-service
- Сервис (`internal/service`):
  - `Register`: успешное создание пользователя + выпуск токенов; дубликат email → `ErrUserAlreadyExists`.
  - `Login`: успешный вход с валидным паролем; ошибка при неверном пароле.
  - `RefreshToken`: успешный reissue; ошибка парсинга токена.
  - GitHub OAuth login: создаёт нового пользователя при отсутствии; обновляет существующего.
  - Профиль: `GetProfile`/`UpdateProfile`; предпочтения: `GetPreferences`/`UpdatePreferences`.
  - Управление пользователями: `ListUsers`, `UpdateUserRole`, `UpdateUserStatus`.
- Репозиторий БД (`internal/adapters/database`):
  - `Create`, `GetByID`/`GetByEmail` (NotFound), `Update` (NotFound), `Delete`.
  - `UpsertPreferences` сохранение значений.
  - `WithTransaction` — вызывает вложенный блок.
- HTTP auth (`internal/adapters/http/auth`):
  - `Register`, `Login` (невалидные креды), `Refresh`, `Logout`, `Validate`.
- HTTP admin (`internal/adapters/http/admin`):
  - `ListUsers`, `UpdateRole`, `UpdateStatus`.
- HTTP internal API (`internal/adapters/http/internalapi`):
  - `GetUser`, `ValidateToken`.
- HTTP profile (`internal/adapters/http/profile`):
  - `GetProfile`, `UpdatePreferences`.
- HTTP GitHub (`internal/adapters/http/github`):
  - `Begin` (redirect/state), `Callback` (валидный/неверный state, ошибки обмена).
- HTTP middleware (`internal/adapters/http/middleware`):
  - JWT парсинг, извлечение bearer-токена, проверка ролей (`RequireRoles`) и возврат 403 при несоответствии.
- Auth adapters:
  - GitHub OAuth: генерация state, наличие state в URL, обмен кода на токен + получение профиля (через фейковый транспорт).
  - JWT: генерация/парсинг access/refresh токенов, ошибки парсинга.
- HTTP common (`internal/adapters/http/common`): валидация/доменные ошибки и bearer-токен.

## Notification-service
- Сервис (`internal/service`):
  - `Handle`: отправка писем для `task.created`/`task.deleted`; пропуск неизвестных типов; ошибка при пустом email.
