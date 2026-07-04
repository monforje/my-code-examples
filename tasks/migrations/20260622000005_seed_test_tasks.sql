-- +goose Up

INSERT INTO tags (id, name) VALUES
  ('c2000000-0000-0000-0000-000000000001', 'API Design'),
  ('c2000000-0000-0000-0000-000000000002', 'CI/CD'),
  ('c2000000-0000-0000-0000-000000000003', 'Docker'),
  ('c2000000-0000-0000-0000-000000000004', 'JWT'),
  ('c2000000-0000-0000-0000-000000000005', 'Microservices'),
  ('c2000000-0000-0000-0000-000000000006', 'Performance'),
  ('c2000000-0000-0000-0000-000000000007', 'PostgreSQL'),
  ('c2000000-0000-0000-0000-000000000008', 'REST'),
  ('c2000000-0000-0000-0000-000000000009', 'React'),
  ('c2000000-0000-0000-0000-000000000010', 'Redis'),
  ('c2000000-0000-0000-0000-000000000011', 'Security'),
  ('c2000000-0000-0000-0000-000000000012', 'Testing');

INSERT INTO languages (id, name) VALUES
  ('c3000000-0000-0000-0000-000000000001', 'Go'),
  ('c3000000-0000-0000-0000-000000000002', 'TypeScript');

INSERT INTO tasks (id, task_name, title, description, specification_md_text, task_type, level) VALUES
  ('c1000000-0000-0000-0000-000000000001', 'auth-pages', 'Страницы входа, регистрации и подтверждения email', 'Сверстать auth pages в стиле Codurity: /login, /register, /verify-email.', '# ТЗ: Страницы входа, регистрации и подтверждения email

**Область:** Frontend  
**Тип задачи:** backend/frontend feature  
**Приоритет:** уточняется на планировании

## Контекст

Сверстать auth pages в стиле Codurity: /login, /register, /verify-email.

## Цель

Довести задачу до production-ready состояния: реализовать поведение, покрыть основные сценарии тестами и не сломать существующие контракты.

## Контракты / экраны / API

Login: email/password, заголовок «Вход». Register: email/username/password/confirm_password, заголовок «Регистрация». Verify: email/code + resend.

## Функциональные требования

Убрать старые тексты «Codurity аккаунт» и «new аккаунт»; client validation; API integration с /auth/login, /auth/register, /auth/register/verify; ошибки рядом с формой.

## Нефункциональные требования

- Код должен быть читаемым, типизированным и покрытым тестами.
- Ошибки должны обрабатываться явно, без silent fail.
- Логи не должны содержать секреты, пароли, токены и одноразовые коды.
- Реализация должна быть совместима с существующей архитектурой проекта.

## Критерии приёмки

После register переход на verify; после verify переход на login; confirm password валидируется; токены после verify не выдаются.

## Out of scope

- Полная переработка архитектуры вне текущей задачи.
- Изменение публичных контрактов без отдельного согласования.
- Добавление новых продуктовых сценариев, не описанных в этом ТЗ.
', 'frontend', 'middle');

INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000001', id FROM tags WHERE name = 'React';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000001', id FROM tags WHERE name = 'Testing';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000001', id FROM tags WHERE name = 'REST';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000001', id FROM tags WHERE name = 'Security';

INSERT INTO task_languages (task_id, language_id) SELECT 'c1000000-0000-0000-0000-000000000001', id FROM languages WHERE name = 'TypeScript';

INSERT INTO tasks (id, task_name, title, description, specification_md_text, task_type, level) VALUES
  ('c1000000-0000-0000-0000-000000000002', 'auth-registration-verification', 'Регистрация и подтверждение email', 'Реализовать регистрацию без автоматической выдачи токенов: /auth/register, /auth/register/verify, /auth/register/verification/resend.', '# ТЗ: Регистрация и подтверждение email

**Область:** Auth-сервис  
**Тип задачи:** backend/frontend feature  
**Приоритет:** уточняется на планировании

## Контекст

Реализовать регистрацию без автоматической выдачи токенов: /auth/register, /auth/register/verify, /auth/register/verification/resend.

## Цель

Довести задачу до production-ready состояния: реализовать поведение, покрыть основные сценарии тестами и не сломать существующие контракты.

## Контракты / экраны / API

POST /auth/register: email, username, password -> user_id, email, status=pending_verification; POST /auth/register/verify: email, code -> status=verified; POST /auth/register/verification/resend: email -> status=sent.

## Функциональные требования

Нормализовать email; валидировать пароль и username; хранить код в хешированном виде; ограничить попытки; публиковать событие на отправку письма; после verify публиковать identity.created.

## Нефункциональные требования

- Код должен быть читаемым, типизированным и покрытым тестами.
- Ошибки должны обрабатываться явно, без silent fail.
- Логи не должны содержать секреты, пароли, токены и одноразовые коды.
- Реализация должна быть совместима с существующей архитектурой проекта.

## Критерии приёмки

Нельзя войти до verify; просроченный/повторный код не работает; resend rate-limited; ошибки в едином формате.

## Out of scope

- Полная переработка архитектуры вне текущей задачи.
- Изменение публичных контрактов без отдельного согласования.
- Добавление новых продуктовых сценариев, не описанных в этом ТЗ.
', 'backend', 'middle');

INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000002', id FROM tags WHERE name = 'Security';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000002', id FROM tags WHERE name = 'JWT';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000002', id FROM tags WHERE name = 'REST';

INSERT INTO task_languages (task_id, language_id) SELECT 'c1000000-0000-0000-0000-000000000002', id FROM languages WHERE name = 'Go';

INSERT INTO tasks (id, task_name, title, description, specification_md_text, task_type, level) VALUES
  ('c1000000-0000-0000-0000-000000000003', 'password-recovery-ui', 'UI восстановления пароля', 'Реализовать flow восстановления пароля на русском языке.', '# ТЗ: UI восстановления пароля

**Область:** Frontend  
**Тип задачи:** backend/frontend feature  
**Приоритет:** уточняется на планировании

## Контекст

Реализовать flow восстановления пароля на русском языке.

## Цель

Довести задачу до production-ready состояния: реализовать поведение, покрыть основные сценарии тестами и не сломать существующие контракты.

## Контракты / экраны / API

Страницы: /password-recovery, /password-recovery/verify, /password-recovery/reset.

## Функциональные требования

Заголовок «Восстановить пароль»; убрать «Password Recovery»; forgot показывает нейтральный ответ; reset_token держать только в runtime state; password confirmation.

## Нефункциональные требования

- Код должен быть читаемым, типизированным и покрытым тестами.
- Ошибки должны обрабатываться явно, без silent fail.
- Логи не должны содержать секреты, пароли, токены и одноразовые коды.
- Реализация должна быть совместима с существующей архитектурой проекта.

## Критерии приёмки

Пользователь не видит, существует email или нет; после reset переход на login; ошибки отображаются единообразно.

## Out of scope

- Полная переработка архитектуры вне текущей задачи.
- Изменение публичных контрактов без отдельного согласования.
- Добавление новых продуктовых сценариев, не описанных в этом ТЗ.
', 'frontend', 'junior');

INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000003', id FROM tags WHERE name = 'React';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000003', id FROM tags WHERE name = 'Testing';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000003', id FROM tags WHERE name = 'REST';

INSERT INTO task_languages (task_id, language_id) SELECT 'c1000000-0000-0000-0000-000000000003', id FROM languages WHERE name = 'TypeScript';

INSERT INTO tasks (id, task_name, title, description, specification_md_text, task_type, level) VALUES
  ('c1000000-0000-0000-0000-000000000004', 'session-login-refresh-logout', 'Login, refresh и logout', 'Реализовать сессионный flow с access token в Bearer Authorization и refresh token только в HttpOnly cookie.', '# ТЗ: Login, refresh и logout

**Область:** Auth-сервис  
**Тип задачи:** backend/frontend feature  
**Приоритет:** уточняется на планировании

## Контекст

Реализовать сессионный flow с access token в Bearer Authorization и refresh token только в HttpOnly cookie.

## Цель

Довести задачу до production-ready состояния: реализовать поведение, покрыть основные сценарии тестами и не сломать существующие контракты.

## Контракты / экраны / API

POST /auth/login: email, password -> access_token, expires_in, user + Set-Cookie refresh_token; POST /auth/refresh: cookie -> новый access_token + rotation cookie; POST /auth/logout -> 204.

## Функциональные требования

Login разрешён только verified пользователям; refresh token хранить сервер-side как хеш; делать refresh rotation; logout отзывает текущую сессию; reuse старого refresh логировать как security event.

## Нефункциональные требования

- Код должен быть читаемым, типизированным и покрытым тестами.
- Ошибки должны обрабатываться явно, без silent fail.
- Логи не должны содержать секреты, пароли, токены и одноразовые коды.
- Реализация должна быть совместима с существующей архитектурой проекта.

## Критерии приёмки

Refresh никогда не возвращается в JSON; после logout refresh не работает; параллельный refresh не ломает сессию.

## Out of scope

- Полная переработка архитектуры вне текущей задачи.
- Изменение публичных контрактов без отдельного согласования.
- Добавление новых продуктовых сценариев, не описанных в этом ТЗ.
', 'backend', 'middle');

INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000004', id FROM tags WHERE name = 'Security';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000004', id FROM tags WHERE name = 'JWT';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000004', id FROM tags WHERE name = 'REST';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000004', id FROM tags WHERE name = 'Testing';

INSERT INTO task_languages (task_id, language_id) SELECT 'c1000000-0000-0000-0000-000000000004', id FROM languages WHERE name = 'Go';

INSERT INTO tasks (id, task_name, title, description, specification_md_text, task_type, level) VALUES
  ('c1000000-0000-0000-0000-000000000005', 'header-avatar-theme', 'Header, avatar dropdown и theme switcher', 'Реализовать header для гостя и авторизованного пользователя.', '# ТЗ: Header, avatar dropdown и theme switcher

**Область:** Frontend  
**Тип задачи:** backend/frontend feature  
**Приоритет:** уточняется на планировании

## Контекст

Реализовать header для гостя и авторизованного пользователя.

## Цель

Довести задачу до production-ready состояния: реализовать поведение, покрыть основные сценарии тестами и не сломать существующие контракты.

## Контракты / экраны / API

Гость: logo, theme switcher, «Войти», «Создать аккаунт». Auth: logo, круглый avatar, dropdown: «Профиль», «Настройки», theme switcher, «Выход».

## Функциональные требования

Theme switch не делает reload и не закрывает dropdown; avatar hover круглый и не выходит за avatar; в dark theme logo белый; border-line одинаковая.

## Нефункциональные требования

- Код должен быть читаемым, типизированным и покрытым тестами.
- Ошибки должны обрабатываться явно, без silent fail.
- Логи не должны содержать секреты, пароли, токены и одноразовые коды.
- Реализация должна быть совместима с существующей архитектурой проекта.

## Критерии приёмки

Dropdown остаётся открыт при смене темы; кнопки гостя видны на белом фоне; avatar чуть больше текущего.

## Out of scope

- Полная переработка архитектуры вне текущей задачи.
- Изменение публичных контрактов без отдельного согласования.
- Добавление новых продуктовых сценариев, не описанных в этом ТЗ.
', 'frontend', 'middle');

INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000005', id FROM tags WHERE name = 'React';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000005', id FROM tags WHERE name = 'Testing';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000005', id FROM tags WHERE name = 'Performance';

INSERT INTO task_languages (task_id, language_id) SELECT 'c1000000-0000-0000-0000-000000000005', id FROM languages WHERE name = 'TypeScript';

INSERT INTO tasks (id, task_name, title, description, specification_md_text, task_type, level) VALUES
  ('c1000000-0000-0000-0000-000000000006', 'password-recovery', 'Восстановление пароля', 'Реализовать безопасный password recovery без раскрытия существования email.', '# ТЗ: Восстановление пароля

**Область:** Auth-сервис  
**Тип задачи:** backend/frontend feature  
**Приоритет:** уточняется на планировании

## Контекст

Реализовать безопасный password recovery без раскрытия существования email.

## Цель

Довести задачу до production-ready состояния: реализовать поведение, покрыть основные сценарии тестами и не сломать существующие контракты.

## Контракты / экраны / API

POST /auth/password/forgot; POST /auth/password/forgot/verify; POST /auth/password/forgot/verification/resend; POST /auth/password/reset.

## Функциональные требования

forgot всегда возвращает нейтральный ответ; code имеет TTL и лимит попыток; verify возвращает short-lived reset_token; reset меняет пароль и отзывает все refresh-сессии.

## Нефункциональные требования

- Код должен быть читаемым, типизированным и покрытым тестами.
- Ошибки должны обрабатываться явно, без silent fail.
- Логи не должны содержать секреты, пароли, токены и одноразовые коды.
- Реализация должна быть совместима с существующей архитектурой проекта.

## Критерии приёмки

Нельзя сменить пароль без reset_token; reset_token одноразовый; старые сессии после reset отозваны.

## Out of scope

- Полная переработка архитектуры вне текущей задачи.
- Изменение публичных контрактов без отдельного согласования.
- Добавление новых продуктовых сценариев, не описанных в этом ТЗ.
', 'backend', 'middle');

INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000006', id FROM tags WHERE name = 'Security';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000006', id FROM tags WHERE name = 'REST';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000006', id FROM tags WHERE name = 'Testing';

INSERT INTO task_languages (task_id, language_id) SELECT 'c1000000-0000-0000-0000-000000000006', id FROM languages WHERE name = 'Go';

INSERT INTO tasks (id, task_name, title, description, specification_md_text, task_type, level) VALUES
  ('c1000000-0000-0000-0000-000000000007', 'email-change', 'Изменение email', 'Сделать flow смены email для авторизованного пользователя с подтверждением кода на новый email.', '# ТЗ: Изменение email

**Область:** Auth-сервис  
**Тип задачи:** backend/frontend feature  
**Приоритет:** уточняется на планировании

## Контекст

Сделать flow смены email для авторизованного пользователя с подтверждением кода на новый email.

## Цель

Довести задачу до production-ready состояния: реализовать поведение, покрыть основные сценарии тестами и не сломать существующие контракты.

## Контракты / экраны / API

POST /me/email/change: new_email -> verification_required; POST /me/email/change/verify: new_email, code -> email; POST /me/email/change/verification/resend.

## Функциональные требования

Все endpoints требуют Bearer; новый email нормализовать и проверять на уникальность; текущий email не меняется до verify; после verify публиковать identity.updated.

## Нефункциональные требования

- Код должен быть читаемым, типизированным и покрытым тестами.
- Ошибки должны обрабатываться явно, без silent fail.
- Логи не должны содержать секреты, пароли, токены и одноразовые коды.
- Реализация должна быть совместима с существующей архитектурой проекта.

## Критерии приёмки

Нельзя сменить email без токена; нельзя занять чужой email; users-service получает update event.

## Out of scope

- Полная переработка архитектуры вне текущей задачи.
- Изменение публичных контрактов без отдельного согласования.
- Добавление новых продуктовых сценариев, не описанных в этом ТЗ.
', 'backend', 'middle');

INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000007', id FROM tags WHERE name = 'Security';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000007', id FROM tags WHERE name = 'REST';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000007', id FROM tags WHERE name = 'Microservices';

INSERT INTO task_languages (task_id, language_id) SELECT 'c1000000-0000-0000-0000-000000000007', id FROM languages WHERE name = 'Go';

INSERT INTO tasks (id, task_name, title, description, specification_md_text, task_type, level) VALUES
  ('c1000000-0000-0000-0000-000000000008', 'profile-page', 'Страница профиля', 'Реализовать /profile отдельно от настроек.', '# ТЗ: Страница профиля

**Область:** Frontend  
**Тип задачи:** backend/frontend feature  
**Приоритет:** уточняется на планировании

## Контекст

Реализовать /profile отдельно от настроек.

## Цель

Довести задачу до production-ready состояния: реализовать поведение, покрыть основные сценарии тестами и не сломать существующие контракты.

## Контракты / экраны / API

Sidebar содержит только «Профиль». Content: avatar, username, display_name, bio, save button.

## Функциональные требования

GET/PATCH /profile/me; не показывать user id, дату создания, email verification status, «активен»; PATCH отправляет только изменённые поля.

## Нефункциональные требования

- Код должен быть читаемым, типизированным и покрытым тестами.
- Ошибки должны обрабатываться явно, без silent fail.
- Логи не должны содержать секреты, пароли, токены и одноразовые коды.
- Реализация должна быть совместима с существующей архитектурой проекта.

## Критерии приёмки

Профиль сохраняется без reload; скрытые account-data нигде не отображаются; loading/error states реализованы.

## Out of scope

- Полная переработка архитектуры вне текущей задачи.
- Изменение публичных контрактов без отдельного согласования.
- Добавление новых продуктовых сценариев, не описанных в этом ТЗ.
', 'frontend', 'junior');

INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000008', id FROM tags WHERE name = 'React';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000008', id FROM tags WHERE name = 'Testing';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000008', id FROM tags WHERE name = 'REST';

INSERT INTO task_languages (task_id, language_id) SELECT 'c1000000-0000-0000-0000-000000000008', id FROM languages WHERE name = 'TypeScript';

INSERT INTO tasks (id, task_name, title, description, specification_md_text, task_type, level) VALUES
  ('c1000000-0000-0000-0000-000000000009', 'account-delete', 'Удаление аккаунта', 'Реализовать удаление аккаунта с подтверждением по email-коду.', '# ТЗ: Удаление аккаунта

**Область:** Auth-сервис  
**Тип задачи:** backend/frontend feature  
**Приоритет:** уточняется на планировании

## Контекст

Реализовать удаление аккаунта с подтверждением по email-коду.

## Цель

Довести задачу до production-ready состояния: реализовать поведение, покрыть основные сценарии тестами и не сломать существующие контракты.

## Контракты / экраны / API

DELETE /me -> verification_required; POST /me/delete/verify: code -> 204.

## Функциональные требования

DELETE создаёт код удаления; verify делает soft delete или hard delete по выбранной политике; отозвать все сессии; опубликовать identity.deleted; повторные запросы идемпотентны.

## Нефункциональные требования

- Код должен быть читаемым, типизированным и покрытым тестами.
- Ошибки должны обрабатываться явно, без silent fail.
- Логи не должны содержать секреты, пароли, токены и одноразовые коды.
- Реализация должна быть совместима с существующей архитектурой проекта.

## Критерии приёмки

Без кода аккаунт не удаляется; после удаления нельзя login/refresh; событие удаления доставляется.

## Out of scope

- Полная переработка архитектуры вне текущей задачи.
- Изменение публичных контрактов без отдельного согласования.
- Добавление новых продуктовых сценариев, не описанных в этом ТЗ.
', 'backend', 'middle');

INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000009', id FROM tags WHERE name = 'Security';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000009', id FROM tags WHERE name = 'REST';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000009', id FROM tags WHERE name = 'Testing';

INSERT INTO task_languages (task_id, language_id) SELECT 'c1000000-0000-0000-0000-000000000009', id FROM languages WHERE name = 'Go';

INSERT INTO tasks (id, task_name, title, description, specification_md_text, task_type, level) VALUES
  ('c1000000-0000-0000-0000-000000000010', 'settings-security-page', 'Страница настроек безопасности', 'Реализовать /settings/security отдельно от профиля.', '# ТЗ: Страница настроек безопасности

**Область:** Frontend  
**Тип задачи:** backend/frontend feature  
**Приоритет:** уточняется на планировании

## Контекст

Реализовать /settings/security отдельно от профиля.

## Цель

Довести задачу до production-ready состояния: реализовать поведение, покрыть основные сценарии тестами и не сломать существующие контракты.

## Контракты / экраны / API

Sidebar содержит только «Безопасность». Блоки: смена email, смена пароля, удаление аккаунта.

## Функциональные требования

Drawer/modal для операций должен быть компактным, без больших пробелов; все действия подтверждаются кодом; mobile drawer full width.

## Нефункциональные требования

- Код должен быть читаемым, типизированным и покрытым тестами.
- Ошибки должны обрабатываться явно, без silent fail.
- Логи не должны содержать секреты, пароли, токены и одноразовые коды.
- Реализация должна быть совместима с существующей архитектурой проекта.

## Критерии приёмки

Страница не содержит profile fields; drawer layout ровный; dangerous actions требуют подтверждения.

## Out of scope

- Полная переработка архитектуры вне текущей задачи.
- Изменение публичных контрактов без отдельного согласования.
- Добавление новых продуктовых сценариев, не описанных в этом ТЗ.
', 'frontend', 'middle');

INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000010', id FROM tags WHERE name = 'React';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000010', id FROM tags WHERE name = 'Testing';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000010', id FROM tags WHERE name = 'Security';

INSERT INTO task_languages (task_id, language_id) SELECT 'c1000000-0000-0000-0000-000000000010', id FROM languages WHERE name = 'TypeScript';

INSERT INTO tasks (id, task_name, title, description, specification_md_text, task_type, level) VALUES
  ('c1000000-0000-0000-0000-000000000011', 'design-system-tokens', 'Design System tokens и темы', 'Оформить базовые CSS tokens и документацию дизайн-системы Codurity.', '# ТЗ: Design System tokens и темы

**Область:** Frontend  
**Тип задачи:** backend/frontend feature  
**Приоритет:** уточняется на планировании

## Контекст

Оформить базовые CSS tokens и документацию дизайн-системы Codurity.

## Цель

Довести задачу до production-ready состояния: реализовать поведение, покрыть основные сценарии тестами и не сломать существующие контракты.

## Контракты / экраны / API

Файлы: tokens.css, themes.css, globals.css, design-system.md. Brand base #1758d1 и blue scale от #e8f3ff до #0041a3.

## Функциональные требования

Группы tokens: colors, typography, spacing, radius, shadows, borders, z-index, transitions; light/dark themes; не использовать шрифт логотипа в UI.

## Нефункциональные требования

- Код должен быть читаемым, типизированным и покрытым тестами.
- Ошибки должны обрабатываться явно, без silent fail.
- Логи не должны содержать секреты, пароли, токены и одноразовые коды.
- Реализация должна быть совместима с существующей архитектурой проекта.

## Критерии приёмки

Все компоненты используют CSS variables; тема переключается без reload; tokens задокументированы.

## Out of scope

- Полная переработка архитектуры вне текущей задачи.
- Изменение публичных контрактов без отдельного согласования.
- Добавление новых продуктовых сценариев, не описанных в этом ТЗ.
', 'frontend', 'senior');

INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000011', id FROM tags WHERE name = 'React';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000011', id FROM tags WHERE name = 'Testing';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000011', id FROM tags WHERE name = 'CI/CD';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000011', id FROM tags WHERE name = 'Performance';

INSERT INTO task_languages (task_id, language_id) SELECT 'c1000000-0000-0000-0000-000000000011', id FROM languages WHERE name = 'TypeScript';

INSERT INTO tasks (id, task_name, title, description, specification_md_text, task_type, level) VALUES
  ('c1000000-0000-0000-0000-000000000012', 'users-event-sync', 'Синхронизация профилей по событиям', 'Реализовать worker, который создаёт/обновляет/удаляет профиль по событиям identity.created/updated/deleted.', '# ТЗ: Синхронизация профилей по событиям

**Область:** Users-service  
**Тип задачи:** backend/frontend feature  
**Приоритет:** уточняется на планировании

## Контекст

Реализовать worker, который создаёт/обновляет/удаляет профиль по событиям identity.created/updated/deleted.

## Цель

Довести задачу до production-ready состояния: реализовать поведение, покрыть основные сценарии тестами и не сломать существующие контракты.

## Контракты / экраны / API

NATS subjects: identity.created, identity.updated, identity.deleted. Payload содержит event_id, event_type, occurred_at, user_id, email, username.

## Функциональные требования

Worker отдельным процессом cmd/users/worker; обработка идемпотентная; хранить processed_events; created создаёт профиль с nil optional fields; deleted помечает профиль удалённым.

## Нефункциональные требования

- Код должен быть читаемым, типизированным и покрытым тестами.
- Ошибки должны обрабатываться явно, без silent fail.
- Логи не должны содержать секреты, пароли, токены и одноразовые коды.
- Реализация должна быть совместима с существующей архитектурой проекта.

## Критерии приёмки

Повторная доставка не создаёт дублей; профиль обновляется после смены email; удалённый профиль не отдаётся API.

## Out of scope

- Полная переработка архитектуры вне текущей задачи.
- Изменение публичных контрактов без отдельного согласования.
- Добавление новых продуктовых сценариев, не описанных в этом ТЗ.
', 'backend', 'senior');

INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000012', id FROM tags WHERE name = 'Microservices';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000012', id FROM tags WHERE name = 'PostgreSQL';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000012', id FROM tags WHERE name = 'Testing';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000012', id FROM tags WHERE name = 'Security';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000012', id FROM tags WHERE name = 'Docker';

INSERT INTO task_languages (task_id, language_id) SELECT 'c1000000-0000-0000-0000-000000000012', id FROM languages WHERE name = 'Go';

INSERT INTO tasks (id, task_name, title, description, specification_md_text, task_type, level) VALUES
  ('c1000000-0000-0000-0000-000000000013', 'api-client-auth-refresh', 'API client с Bearer auth', 'Сделать единый HTTP client для access token и refresh flow.', '# ТЗ: API client с Bearer auth

**Область:** Frontend  
**Тип задачи:** backend/frontend feature  
**Приоритет:** уточняется на планировании

## Контекст

Сделать единый HTTP client для access token и refresh flow.

## Цель

Довести задачу до production-ready состояния: реализовать поведение, покрыть основные сценарии тестами и не сломать существующие контракты.

## Контракты / экраны / API

Подставлять Authorization: Bearer; при EXPIRED_AUTH_TOKEN делать POST /auth/refresh; повторять исходный запрос.

## Функциональные требования

Refresh token не читать из JS; защититься от параллельного refresh storm; при refresh failure очистить auth state и отправить на /login.

## Нефункциональные требования

- Код должен быть читаемым, типизированным и покрытым тестами.
- Ошибки должны обрабатываться явно, без silent fail.
- Логи не должны содержать секреты, пароли, токены и одноразовые коды.
- Реализация должна быть совместима с существующей архитектурой проекта.

## Критерии приёмки

Protected requests получают Bearer; истёкший access обновляется; нет бесконечных refresh попыток.

## Out of scope

- Полная переработка архитектуры вне текущей задачи.
- Изменение публичных контрактов без отдельного согласования.
- Добавление новых продуктовых сценариев, не описанных в этом ТЗ.
', 'frontend', 'senior');

INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000013', id FROM tags WHERE name = 'React';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000013', id FROM tags WHERE name = 'Testing';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000013', id FROM tags WHERE name = 'Security';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000013', id FROM tags WHERE name = 'REST';

INSERT INTO task_languages (task_id, language_id) SELECT 'c1000000-0000-0000-0000-000000000013', id FROM languages WHERE name = 'TypeScript';

INSERT INTO tasks (id, task_name, title, description, specification_md_text, task_type, level) VALUES
  ('c1000000-0000-0000-0000-000000000014', 'profile-me-api', 'API профиля текущего пользователя', 'Реализовать GET/PATCH /profile/me для получения и редактирования профиля текущего пользователя.', '# ТЗ: API профиля текущего пользователя

**Область:** Users-service  
**Тип задачи:** backend/frontend feature  
**Приоритет:** уточняется на планировании

## Контекст

Реализовать GET/PATCH /profile/me для получения и редактирования профиля текущего пользователя.

## Цель

Довести задачу до production-ready состояния: реализовать поведение, покрыть основные сценарии тестами и не сломать существующие контракты.

## Контракты / экраны / API

GET /profile/me -> id, email, username, display_name, bio, avatar_url; PATCH /profile/me: display_name, bio, avatar_url -> обновлённый профиль.

## Функциональные требования

email и username read-only; PATCH частичный; display_name 2..64; bio до 300; avatar_url только из разрешённого storage; user_id брать из auth context.

## Нефункциональные требования

- Код должен быть читаемым, типизированным и покрытым тестами.
- Ошибки должны обрабатываться явно, без silent fail.
- Логи не должны содержать секреты, пароли, токены и одноразовые коды.
- Реализация должна быть совместима с существующей архитектурой проекта.

## Критерии приёмки

Нельзя редактировать чужой профиль; read-only поля не меняются; ошибки валидации возвращаются по полям.

## Out of scope

- Полная переработка архитектуры вне текущей задачи.
- Изменение публичных контрактов без отдельного согласования.
- Добавление новых продуктовых сценариев, не описанных в этом ТЗ.
', 'backend', 'junior');

INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000014', id FROM tags WHERE name = 'REST';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000014', id FROM tags WHERE name = 'Testing';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000014', id FROM tags WHERE name = 'API Design';

INSERT INTO task_languages (task_id, language_id) SELECT 'c1000000-0000-0000-0000-000000000014', id FROM languages WHERE name = 'Go';

INSERT INTO tasks (id, task_name, title, description, specification_md_text, task_type, level) VALUES
  ('c1000000-0000-0000-0000-000000000015', 'avatar-generation', 'Генерация дефолтного аватара', 'Сделать детерминированный identicon/avatar generator для первого входа пользователя.', '# ТЗ: Генерация дефолтного аватара

**Область:** Users-service  
**Тип задачи:** backend/frontend feature  
**Приоритет:** уточняется на планировании

## Контекст

Сделать детерминированный identicon/avatar generator для первого входа пользователя.

## Цель

Довести задачу до production-ready состояния: реализовать поведение, покрыть основные сценарии тестами и не сломать существующие контракты.

## Контракты / экраны / API

POST /profile/me/avatar/generate -> avatar_url. Также генерация вызывается при создании профиля.

## Функциональные требования

Аватар детерминирован по user_id; формат SVG; сетка 5x5 или 7x7; цвета из brand palette; SVG без script/external refs; результат сохраняется в storage.

## Нефункциональные требования

- Код должен быть читаемым, типизированным и покрытым тестами.
- Ошибки должны обрабатываться явно, без silent fail.
- Логи не должны содержать секреты, пароли, токены и одноразовые коды.
- Реализация должна быть совместима с существующей архитектурой проекта.

## Критерии приёмки

Один user_id всегда даёт одинаковый avatar; SVG безопасен; endpoint требует авторизацию.

## Out of scope

- Полная переработка архитектуры вне текущей задачи.
- Изменение публичных контрактов без отдельного согласования.
- Добавление новых продуктовых сценариев, не описанных в этом ТЗ.
', 'backend', 'junior');

INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000015', id FROM tags WHERE name = 'REST';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000015', id FROM tags WHERE name = 'Testing';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000015', id FROM tags WHERE name = 'Performance';

INSERT INTO task_languages (task_id, language_id) SELECT 'c1000000-0000-0000-0000-000000000015', id FROM languages WHERE name = 'Go';

INSERT INTO tasks (id, task_name, title, description, specification_md_text, task_type, level) VALUES
  ('c1000000-0000-0000-0000-000000000016', 'protected-routes', 'Protected routes и auth guard', 'Реализовать routing guard для private/public-only страниц.', '# ТЗ: Protected routes и auth guard

**Область:** Frontend  
**Тип задачи:** backend/frontend feature  
**Приоритет:** уточняется на планировании

## Контекст

Реализовать routing guard для private/public-only страниц.

## Цель

Довести задачу до production-ready состояния: реализовать поведение, покрыть основные сценарии тестами и не сломать существующие контракты.

## Контракты / экраны / API

Protected: /profile, /settings/security, /objects, /objects/:id. Public-only: /login, /register, /password-recovery.

## Функциональные требования

При первом открытии проверить auth state; если access истёк, попробовать refresh; во время проверки показать loading state; после logout редирект на login.

## Нефункциональные требования

- Код должен быть читаемым, типизированным и покрытым тестами.
- Ошибки должны обрабатываться явно, без silent fail.
- Логи не должны содержать секреты, пароли, токены и одноразовые коды.
- Реализация должна быть совместима с существующей архитектурой проекта.

## Критерии приёмки

Гость не открывает protected pages; auth user не попадает на login/register; нет redirect loop.

## Out of scope

- Полная переработка архитектуры вне текущей задачи.
- Изменение публичных контрактов без отдельного согласования.
- Добавление новых продуктовых сценариев, не описанных в этом ТЗ.
', 'frontend', 'middle');

INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000016', id FROM tags WHERE name = 'React';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000016', id FROM tags WHERE name = 'Testing';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000016', id FROM tags WHERE name = 'Security';

INSERT INTO task_languages (task_id, language_id) SELECT 'c1000000-0000-0000-0000-000000000016', id FROM languages WHERE name = 'TypeScript';

INSERT INTO tasks (id, task_name, title, description, specification_md_text, task_type, level) VALUES
  ('c1000000-0000-0000-0000-000000000017', 'notifications-email-sender', 'Отправка email-уведомлений', 'Реализовать consumer событий и отправку HTML/text email через SMTP.', '# ТЗ: Отправка email-уведомлений

**Область:** Notifications-service  
**Тип задачи:** backend/frontend feature  
**Приоритет:** уточняется на планировании

## Контекст

Реализовать consumer событий и отправку HTML/text email через SMTP.

## Цель

Довести задачу до production-ready состояния: реализовать поведение, покрыть основные сценарии тестами и не сломать существующие контракты.

## Контракты / экраны / API

Поддержать события verification_code, password_reset_code, password_change_code, email_change_code, account_delete_code.

## Функциональные требования

Выбирать шаблон по event_type; рендерить MJML HTML и text fallback; SMTP config из env; retry при временных ошибках; event_id защищает от двойной отправки.

## Нефункциональные требования

- Код должен быть читаемым, типизированным и покрытым тестами.
- Ошибки должны обрабатываться явно, без silent fail.
- Логи не должны содержать секреты, пароли, токены и одноразовые коды.
- Реализация должна быть совместима с существующей архитектурой проекта.

## Критерии приёмки

Одно событие не отправляется дважды; код не логируется полностью; failed delivery сохраняется для диагностики.

## Out of scope

- Полная переработка архитектуры вне текущей задачи.
- Изменение публичных контрактов без отдельного согласования.
- Добавление новых продуктовых сценариев, не описанных в этом ТЗ.
', 'backend', 'middle');

INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000017', id FROM tags WHERE name = 'Microservices';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000017', id FROM tags WHERE name = 'REST';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000017', id FROM tags WHERE name = 'Docker';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000017', id FROM tags WHERE name = 'Testing';

INSERT INTO task_languages (task_id, language_id) SELECT 'c1000000-0000-0000-0000-000000000017', id FROM languages WHERE name = 'Go';

INSERT INTO tasks (id, task_name, title, description, specification_md_text, task_type, level) VALUES
  ('c1000000-0000-0000-0000-000000000018', 'objects-table-pagination', 'Таблица объектов с cursor pagination', 'Сделать страницу /objects со списком объектов.', '# ТЗ: Таблица объектов с cursor pagination

**Область:** Frontend  
**Тип задачи:** backend/frontend feature  
**Приоритет:** уточняется на планировании

## Контекст

Сделать страницу /objects со списком объектов.

## Цель

Довести задачу до production-ready состояния: реализовать поведение, покрыть основные сценарии тестами и не сломать существующие контракты.

## Контракты / экраны / API

UI: search input, sort select, table/cards, «Создать объект», «Загрузить ещё», empty/loading states.

## Функциональные требования

GET /objects?limit&cursor&q&sort; при поиске сбрасывать cursor; load more добавляет items; не допускать дублей при быстрых кликах.

## Нефункциональные требования

- Код должен быть читаемым, типизированным и покрытым тестами.
- Ошибки должны обрабатываться явно, без silent fail.
- Логи не должны содержать секреты, пароли, токены и одноразовые коды.
- Реализация должна быть совместима с существующей архитектурой проекта.

## Критерии приёмки

Cursor передаётся корректно; search+pagination работают вместе; sort перезагружает список.

## Out of scope

- Полная переработка архитектуры вне текущей задачи.
- Изменение публичных контрактов без отдельного согласования.
- Добавление новых продуктовых сценариев, не описанных в этом ТЗ.
', 'frontend', 'middle');

INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000018', id FROM tags WHERE name = 'React';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000018', id FROM tags WHERE name = 'Testing';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000018', id FROM tags WHERE name = 'REST';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000018', id FROM tags WHERE name = 'API Design';

INSERT INTO task_languages (task_id, language_id) SELECT 'c1000000-0000-0000-0000-000000000018', id FROM languages WHERE name = 'TypeScript';

INSERT INTO tasks (id, task_name, title, description, specification_md_text, task_type, level) VALUES
  ('c1000000-0000-0000-0000-000000000019', 'object-editor-yaml-json', 'Редактор объекта JSON/YAML spec', 'Сделать /objects/new и /objects/:id для редактирования title/spec.', '# ТЗ: Редактор объекта JSON/YAML spec

**Область:** Frontend  
**Тип задачи:** backend/frontend feature  
**Приоритет:** уточняется на планировании

## Контекст

Сделать /objects/new и /objects/:id для редактирования title/spec.

## Цель

Довести задачу до production-ready состояния: реализовать поведение, покрыть основные сценарии тестами и не сломать существующие контракты.

## Контракты / экраны / API

Поля: title, spec editor; режимы JSON/YAML; кнопки «Сохранить», «Отмена», «Форматировать», «Проверить».

## Функциональные требования

Невалидный JSON/YAML не отправлять; YAML конвертировать в object; warning при unsaved changes; PATCH отправляет changed fields.

## Нефункциональные требования

- Код должен быть читаемым, типизированным и покрытым тестами.
- Ошибки должны обрабатываться явно, без silent fail.
- Логи не должны содержать секреты, пароли, токены и одноразовые коды.
- Реализация должна быть совместима с существующей архитектурой проекта.

## Критерии приёмки

Ошибки парсинга видны пользователю; после create переход на detail; данные не теряются без подтверждения.

## Out of scope

- Полная переработка архитектуры вне текущей задачи.
- Изменение публичных контрактов без отдельного согласования.
- Добавление новых продуктовых сценариев, не описанных в этом ТЗ.
', 'frontend', 'senior');

INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000019', id FROM tags WHERE name = 'React';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000019', id FROM tags WHERE name = 'Testing';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000019', id FROM tags WHERE name = 'API Design';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000019', id FROM tags WHERE name = 'REST';

INSERT INTO task_languages (task_id, language_id) SELECT 'c1000000-0000-0000-0000-000000000019', id FROM languages WHERE name = 'TypeScript';

INSERT INTO tasks (id, task_name, title, description, specification_md_text, task_type, level) VALUES
  ('c1000000-0000-0000-0000-000000000020', 'transactional-outbox', 'Transactional outbox для событий', 'Исключить потерю событий между записью в БД и публикацией в NATS.', '# ТЗ: Transactional outbox для событий

**Область:** Backend инфраструктура  
**Тип задачи:** backend/frontend feature  
**Приоритет:** уточняется на планировании

## Контекст

Исключить потерю событий между записью в БД и публикацией в NATS.

## Цель

Довести задачу до production-ready состояния: реализовать поведение, покрыть основные сценарии тестами и не сломать существующие контракты.

## Контракты / экраны / API

Таблица outbox_events: id, event_type, aggregate_id, payload, status, attempts, last_error, created_at, published_at.

## Функциональные требования

Событие пишется в той же транзакции, что и доменное изменение; отдельный publisher читает pending; использовать FOR UPDATE SKIP LOCKED; retry с backoff.

## Нефункциональные требования

- Код должен быть читаемым, типизированным и покрытым тестами.
- Ошибки должны обрабатываться явно, без silent fail.
- Логи не должны содержать секреты, пароли, токены и одноразовые коды.
- Реализация должна быть совместима с существующей архитектурой проекта.

## Критерии приёмки

Если сервис падает после commit, событие не теряется; несколько publisher инстансов не дублируют публикацию.

## Out of scope

- Полная переработка архитектуры вне текущей задачи.
- Изменение публичных контрактов без отдельного согласования.
- Добавление новых продуктовых сценариев, не описанных в этом ТЗ.
', 'backend', 'senior');

INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000020', id FROM tags WHERE name = 'PostgreSQL';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000020', id FROM tags WHERE name = 'Microservices';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000020', id FROM tags WHERE name = 'Docker';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000020', id FROM tags WHERE name = 'CI/CD';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000020', id FROM tags WHERE name = 'Testing';

INSERT INTO task_languages (task_id, language_id) SELECT 'c1000000-0000-0000-0000-000000000020', id FROM languages WHERE name = 'Go';

INSERT INTO tasks (id, task_name, title, description, specification_md_text, task_type, level) VALUES
  ('c1000000-0000-0000-0000-000000000021', 'rate-limit-antibruteforce', 'Rate limiting и anti-bruteforce', 'Защитить auth endpoints от brute-force и spam resend.', '# ТЗ: Rate limiting и anti-bruteforce

**Область:** Auth-сервис  
**Тип задачи:** backend/frontend feature  
**Приоритет:** уточняется на планировании

## Контекст

Защитить auth endpoints от brute-force и spam resend.

## Цель

Довести задачу до production-ready состояния: реализовать поведение, покрыть основные сценарии тестами и не сломать существующие контракты.

## Контракты / экраны / API

Применить к login, register, verification verify/resend, password forgot/reset, email change, account delete verify.

## Функциональные требования

Лимиты по IP, email/user_id и endpoint; хранение счётчиков в Redis; 429 с кодом RATE_LIMITED; лимиты настраиваются через config.

## Нефункциональные требования

- Код должен быть читаемым, типизированным и покрытым тестами.
- Ошибки должны обрабатываться явно, без silent fail.
- Логи не должны содержать секреты, пароли, токены и одноразовые коды.
- Реализация должна быть совместима с существующей архитектурой проекта.

## Критерии приёмки

После превышения лимита endpoint возвращает 429; TTL сбрасывает лимит; тесты покрывают email+IP scope.

## Out of scope

- Полная переработка архитектуры вне текущей задачи.
- Изменение публичных контрактов без отдельного согласования.
- Добавление новых продуктовых сценариев, не описанных в этом ТЗ.
', 'backend', 'senior');

INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000021', id FROM tags WHERE name = 'Security';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000021', id FROM tags WHERE name = 'Performance';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000021', id FROM tags WHERE name = 'Redis';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000021', id FROM tags WHERE name = 'Testing';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000021', id FROM tags WHERE name = 'CI/CD';

INSERT INTO task_languages (task_id, language_id) SELECT 'c1000000-0000-0000-0000-000000000021', id FROM languages WHERE name = 'Go';

INSERT INTO tasks (id, task_name, title, description, specification_md_text, task_type, level) VALUES
  ('c1000000-0000-0000-0000-000000000022', 'toasts-errors-empty-states', 'Единые ошибки, toast и empty states', 'Реализовать общие UI primitives для состояния интерфейса.', '# ТЗ: Единые ошибки, toast и empty states

**Область:** Frontend  
**Тип задачи:** backend/frontend feature  
**Приоритет:** уточняется на планировании

## Контекст

Реализовать общие UI primitives для состояния интерфейса.

## Цель

Довести задачу до production-ready состояния: реализовать поведение, покрыть основные сценарии тестами и не сломать существующие контракты.

## Контракты / экраны / API

Компоненты: ToastProvider, ErrorMessage, FieldError, EmptyState, LoadingState, RetryBlock.

## Функциональные требования

Backend error fields показывать рядом с полями; global errors над формой/toast; тексты на русском; не показывать stack traces; не спамить одинаковыми toast.

## Нефункциональные требования

- Код должен быть читаемым, типизированным и покрытым тестами.
- Ошибки должны обрабатываться явно, без silent fail.
- Логи не должны содержать секреты, пароли, токены и одноразовые коды.
- Реализация должна быть совместима с существующей архитектурой проекта.

## Критерии приёмки

Все формы используют единый renderer; success/error states консистентны; toast не ломает mobile layout.

## Out of scope

- Полная переработка архитектуры вне текущей задачи.
- Изменение публичных контрактов без отдельного согласования.
- Добавление новых продуктовых сценариев, не описанных в этом ТЗ.
', 'frontend', 'junior');

INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000022', id FROM tags WHERE name = 'React';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000022', id FROM tags WHERE name = 'Testing';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000022', id FROM tags WHERE name = 'Performance';

INSERT INTO task_languages (task_id, language_id) SELECT 'c1000000-0000-0000-0000-000000000022', id FROM languages WHERE name = 'TypeScript';

INSERT INTO tasks (id, task_name, title, description, specification_md_text, task_type, level) VALUES
  ('c1000000-0000-0000-0000-000000000023', 'audit-log', 'Аудит критических действий', 'Сохранять audit log для security-sensitive действий.', '# ТЗ: Аудит критических действий

**Область:** Auth-сервис  
**Тип задачи:** backend/frontend feature  
**Приоритет:** уточняется на планировании

## Контекст

Сохранять audit log для security-sensitive действий.

## Цель

Довести задачу до production-ready состояния: реализовать поведение, покрыть основные сценарии тестами и не сломать существующие контракты.

## Контракты / экраны / API

События: registered, email.verified, login.success, login.failed, logout, password.changed, email.changed, account.deleted, refresh.reused.

## Функциональные требования

Таблица audit_logs: id, user_id, event_type, ip, user_agent, metadata, request_id, created_at; не хранить пароли, токены и коды.

## Нефункциональные требования

- Код должен быть читаемым, типизированным и покрытым тестами.
- Ошибки должны обрабатываться явно, без silent fail.
- Логи не должны содержать секреты, пароли, токены и одноразовые коды.
- Реализация должна быть совместима с существующей архитектурой проекта.

## Критерии приёмки

Каждое критическое действие создаёт audit record; по request_id можно связать request log и audit event.

## Out of scope

- Полная переработка архитектуры вне текущей задачи.
- Изменение публичных контрактов без отдельного согласования.
- Добавление новых продуктовых сценариев, не описанных в этом ТЗ.
', 'backend', 'middle');

INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000023', id FROM tags WHERE name = 'Security';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000023', id FROM tags WHERE name = 'PostgreSQL';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000023', id FROM tags WHERE name = 'REST';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000023', id FROM tags WHERE name = 'Testing';

INSERT INTO task_languages (task_id, language_id) SELECT 'c1000000-0000-0000-0000-000000000023', id FROM languages WHERE name = 'Go';

INSERT INTO tasks (id, task_name, title, description, specification_md_text, task_type, level) VALUES
  ('c1000000-0000-0000-0000-000000000024', 'responsive-layout', 'Responsive layout основных страниц', 'Адаптировать ключевые страницы под mobile/tablet/desktop.', '# ТЗ: Responsive layout основных страниц

**Область:** Frontend  
**Тип задачи:** backend/frontend feature  
**Приоритет:** уточняется на планировании

## Контекст

Адаптировать ключевые страницы под mobile/tablet/desktop.

## Цель

Довести задачу до production-ready состояния: реализовать поведение, покрыть основные сценарии тестами и не сломать существующие контракты.

## Контракты / экраны / API

Страницы: auth, profile, settings, objects list, object editor. Breakpoints: <640, 640-1024, >1024.

## Функциональные требования

На mobile sidebar превращается в tabs/collapse; drawer full-screen; tables как cards или horizontal scroll; dropdown не выходит за viewport.

## Нефункциональные требования

- Код должен быть читаемым, типизированным и покрытым тестами.
- Ошибки должны обрабатываться явно, без silent fail.
- Логи не должны содержать секреты, пароли, токены и одноразовые коды.
- Реализация должна быть совместима с существующей архитектурой проекта.

## Критерии приёмки

Нет horizontal overflow на 360px; кнопки доступны пальцем; layout стабилен в light/dark.

## Out of scope

- Полная переработка архитектуры вне текущей задачи.
- Изменение публичных контрактов без отдельного согласования.
- Добавление новых продуктовых сценариев, не описанных в этом ТЗ.
', 'frontend', 'middle');

INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000024', id FROM tags WHERE name = 'React';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000024', id FROM tags WHERE name = 'Testing';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000024', id FROM tags WHERE name = 'Performance';

INSERT INTO task_languages (task_id, language_id) SELECT 'c1000000-0000-0000-0000-000000000024', id FROM languages WHERE name = 'TypeScript';

INSERT INTO tasks (id, task_name, title, description, specification_md_text, task_type, level) VALUES
  ('c1000000-0000-0000-0000-000000000025', 'component-library', 'Библиотека базовых UI-компонентов', 'Собрать набор переиспользуемых компонентов в стиле Codurity.', '# ТЗ: Библиотека базовых UI-компонентов

**Область:** Frontend  
**Тип задачи:** backend/frontend feature  
**Приоритет:** уточняется на планировании

## Контекст

Собрать набор переиспользуемых компонентов в стиле Codurity.

## Цель

Довести задачу до production-ready состояния: реализовать поведение, покрыть основные сценарии тестами и не сломать существующие контракты.

## Контракты / экраны / API

Компоненты: Button, Input, Textarea, Select, Switch, Card, Badge, Avatar, Dropdown, Modal, Drawer, Tabs, Table, Pagination, Toast, Skeleton.

## Функциональные требования

Компоненты используют tokens; имеют states disabled/loading/error; доступны с клавиатуры; ARIA где нужно; есть examples/docs.

## Нефункциональные требования

- Код должен быть читаемым, типизированным и покрытым тестами.
- Ошибки должны обрабатываться явно, без silent fail.
- Логи не должны содержать секреты, пароли, токены и одноразовые коды.
- Реализация должна быть совместима с существующей архитектурой проекта.

## Критерии приёмки

Новые страницы используют библиотеку; keyboard navigation работает; states выглядят консистентно.

## Out of scope

- Полная переработка архитектуры вне текущей задачи.
- Изменение публичных контрактов без отдельного согласования.
- Добавление новых продуктовых сценариев, не описанных в этом ТЗ.
', 'frontend', 'senior');

INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000025', id FROM tags WHERE name = 'React';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000025', id FROM tags WHERE name = 'Testing';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000025', id FROM tags WHERE name = 'CI/CD';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000025', id FROM tags WHERE name = 'Performance';

INSERT INTO task_languages (task_id, language_id) SELECT 'c1000000-0000-0000-0000-000000000025', id FROM languages WHERE name = 'TypeScript';

INSERT INTO tasks (id, task_name, title, description, specification_md_text, task_type, level) VALUES
  ('c1000000-0000-0000-0000-000000000026', 'objects-crud-pagination', 'CRUD объектов с spec и pagination', 'Реализовать хранение объектов id/title/spec в Postgres с cursor pagination.', '# ТЗ: CRUD объектов с spec и pagination

**Область:** Objects API  
**Тип задачи:** backend/frontend feature  
**Приоритет:** уточняется на планировании

## Контекст

Реализовать хранение объектов id/title/spec в Postgres с cursor pagination.

## Цель

Довести задачу до production-ready состояния: реализовать поведение, покрыть основные сценарии тестами и не сломать существующие контракты.

## Контракты / экраны / API

POST /objects; GET /objects?limit&cursor&q&sort; GET /objects/{id}; PATCH /objects/{id}; DELETE /objects/{id}.

## Функциональные требования

spec хранить как jsonb; title 1..120; spec до 128KB; cursor кодирует created_at+id; сортировка default created_at desc, id desc; YAML принимать только отдельным content-type или конвертировать на клиенте.

## Нефункциональные требования

- Код должен быть читаемым, типизированным и покрытым тестами.
- Ошибки должны обрабатываться явно, без silent fail.
- Логи не должны содержать секреты, пароли, токены и одноразовые коды.
- Реализация должна быть совместима с существующей архитектурой проекта.

## Критерии приёмки

Pagination стабильна при одинаковом created_at; PATCH частичный; invalid spec возвращает validation error.

## Out of scope

- Полная переработка архитектуры вне текущей задачи.
- Изменение публичных контрактов без отдельного согласования.
- Добавление новых продуктовых сценариев, не описанных в этом ТЗ.
', 'backend', 'middle');

INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000026', id FROM tags WHERE name = 'REST';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000026', id FROM tags WHERE name = 'PostgreSQL';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000026', id FROM tags WHERE name = 'Testing';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000026', id FROM tags WHERE name = 'API Design';

INSERT INTO task_languages (task_id, language_id) SELECT 'c1000000-0000-0000-0000-000000000026', id FROM languages WHERE name = 'Go';

INSERT INTO tasks (id, task_name, title, description, specification_md_text, task_type, level) VALUES
  ('c1000000-0000-0000-0000-000000000027', 'health-metrics-logs', 'Health checks, metrics и structured logs', 'Добавить базовую observability поддержку для сервисов.', '# ТЗ: Health checks, metrics и structured logs

**Область:** Backend инфраструктура  
**Тип задачи:** backend/frontend feature  
**Приоритет:** уточняется на планировании

## Контекст

Добавить базовую observability поддержку для сервисов.

## Цель

Довести задачу до production-ready состояния: реализовать поведение, покрыть основные сценарии тестами и не сломать существующие контракты.

## Контракты / экраны / API

GET /healthz -> живость процесса; GET /readyz -> postgres/nats/redis checks; GET /metrics -> Prometheus metrics.

## Функциональные требования

Логи JSON; request_id на каждый запрос; метрики http_requests_total, duration, db duration, nats publish/consume, outbox pending; не логировать секреты.

## Нефункциональные требования

- Код должен быть читаемым, типизированным и покрытым тестами.
- Ошибки должны обрабатываться явно, без silent fail.
- Логи не должны содержать секреты, пароли, токены и одноразовые коды.
- Реализация должна быть совместима с существующей архитектурой проекта.

## Критерии приёмки

healthz не зависит от БД; readyz отдаёт 503 при недоступной зависимости; метрики доступны Prometheus.

## Out of scope

- Полная переработка архитектуры вне текущей задачи.
- Изменение публичных контрактов без отдельного согласования.
- Добавление новых продуктовых сценариев, не описанных в этом ТЗ.
', 'backend', 'senior');

INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000027', id FROM tags WHERE name = 'Docker';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000027', id FROM tags WHERE name = 'CI/CD';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000027', id FROM tags WHERE name = 'Performance';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000027', id FROM tags WHERE name = 'PostgreSQL';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000027', id FROM tags WHERE name = 'Testing';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000027', id FROM tags WHERE name = 'Redis';

INSERT INTO task_languages (task_id, language_id) SELECT 'c1000000-0000-0000-0000-000000000027', id FROM languages WHERE name = 'Go';

INSERT INTO tasks (id, task_name, title, description, specification_md_text, task_type, level) VALUES
  ('c1000000-0000-0000-0000-000000000028', 'theme-persistence', 'Переключение и сохранение темы', 'Реализовать theme manager: light/dark/system.', '# ТЗ: Переключение и сохранение темы

**Область:** Frontend  
**Тип задачи:** backend/frontend feature  
**Приоритет:** уточняется на планировании

## Контекст

Реализовать theme manager: light/dark/system.

## Цель

Довести задачу до production-ready состояния: реализовать поведение, покрыть основные сценарии тестами и не сломать существующие контракты.

## Контракты / экраны / API

Ставить data-theme на html/body; сохранять выбор в localStorage; system слушает prefers-color-scheme.

## Функциональные требования

Switcher интегрирован в guest header и avatar dropdown; клик по switcher не закрывает dropdown; смена темы без reload.

## Нефункциональные требования

- Код должен быть читаемым, типизированным и покрытым тестами.
- Ошибки должны обрабатываться явно, без silent fail.
- Логи не должны содержать секреты, пароли, токены и одноразовые коды.
- Реализация должна быть совместима с существующей архитектурой проекта.

## Критерии приёмки

Тема сохраняется после refresh; system реагирует на изменение ОС; dark logo белый.

## Out of scope

- Полная переработка архитектуры вне текущей задачи.
- Изменение публичных контрактов без отдельного согласования.
- Добавление новых продуктовых сценариев, не описанных в этом ТЗ.
', 'frontend', 'junior');

INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000028', id FROM tags WHERE name = 'React';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000028', id FROM tags WHERE name = 'Testing';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000028', id FROM tags WHERE name = 'Performance';

INSERT INTO task_languages (task_id, language_id) SELECT 'c1000000-0000-0000-0000-000000000028', id FROM languages WHERE name = 'TypeScript';

INSERT INTO tasks (id, task_name, title, description, specification_md_text, task_type, level) VALUES
  ('c1000000-0000-0000-0000-000000000029', 'admin-user-search', 'Admin API поиска пользователей', 'Сделать read-only API для поиска пользователей с доступом только для admin role.', '# ТЗ: Admin API поиска пользователей

**Область:** Admin backend  
**Тип задачи:** backend/frontend feature  
**Приоритет:** уточняется на планировании

## Контекст

Сделать read-only API для поиска пользователей с доступом только для admin role.

## Цель

Довести задачу до production-ready состояния: реализовать поведение, покрыть основные сценарии тестами и не сломать существующие контракты.

## Контракты / экраны / API

GET /admin/users?q&email&username&status&limit&cursor; GET /admin/users/{id}.

## Функциональные требования

Проверять Bearer token и role=admin; не отдавать password hash/tokens/codes; поиск case-insensitive; все admin-запросы писать в audit log.

## Нефункциональные требования

- Код должен быть читаемым, типизированным и покрытым тестами.
- Ошибки должны обрабатываться явно, без silent fail.
- Логи не должны содержать секреты, пароли, токены и одноразовые коды.
- Реализация должна быть совместима с существующей архитектурой проекта.

## Критерии приёмки

Без admin role возвращается 403; секретные поля не попадают в response; поиск paginated.

## Out of scope

- Полная переработка архитектуры вне текущей задачи.
- Изменение публичных контрактов без отдельного согласования.
- Добавление новых продуктовых сценариев, не описанных в этом ТЗ.
', 'backend', 'middle');

INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000029', id FROM tags WHERE name = 'REST';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000029', id FROM tags WHERE name = 'Security';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000029', id FROM tags WHERE name = 'PostgreSQL';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000029', id FROM tags WHERE name = 'Testing';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000029', id FROM tags WHERE name = 'API Design';

INSERT INTO task_languages (task_id, language_id) SELECT 'c1000000-0000-0000-0000-000000000029', id FROM languages WHERE name = 'Go';

INSERT INTO tasks (id, task_name, title, description, specification_md_text, task_type, level) VALUES
  ('c1000000-0000-0000-0000-000000000030', 'frontend-e2e-tests', 'E2E-тесты критических frontend-flow', 'Покрыть основные сценарии Playwright-тестами.', '# ТЗ: E2E-тесты критических frontend-flow

**Область:** Frontend QA  
**Тип задачи:** backend/frontend feature  
**Приоритет:** уточняется на планировании

## Контекст

Покрыть основные сценарии Playwright-тестами.

## Цель

Довести задачу до production-ready состояния: реализовать поведение, покрыть основные сценарии тестами и не сломать существующие контракты.

## Контракты / экраны / API

Сценарии: register -> verify, login, theme switch persistence, profile update, settings drawer.

## Функциональные требования

Не зависеть от реального SMTP; API можно мокировать через network; использовать data-testid; запускать e2e отдельно в CI.

## Нефункциональные требования

- Код должен быть читаемым, типизированным и покрытым тестами.
- Ошибки должны обрабатываться явно, без silent fail.
- Логи не должны содержать секреты, пароли, токены и одноразовые коды.
- Реализация должна быть совместима с существующей архитектурой проекта.

## Критерии приёмки

Тесты стабильны локально и в CI; проверены validation errors; theme switch test подтверждает отсутствие reload.

## Out of scope

- Полная переработка архитектуры вне текущей задачи.
- Изменение публичных контрактов без отдельного согласования.
- Добавление новых продуктовых сценариев, не описанных в этом ТЗ.
', 'frontend', 'middle');

INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000030', id FROM tags WHERE name = 'Testing';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000030', id FROM tags WHERE name = 'CI/CD';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000030', id FROM tags WHERE name = 'React';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000030', id FROM tags WHERE name = 'REST';

INSERT INTO task_languages (task_id, language_id) SELECT 'c1000000-0000-0000-0000-000000000030', id FROM languages WHERE name = 'TypeScript';

INSERT INTO tasks (id, task_name, title, description, specification_md_text, task_type, level) VALUES
  ('c1000000-0000-0000-0000-000000000031', 'pizza-api', 'Pizza Ordering API', 'Разработать REST API для оформления заказов пиццы с авторизацией пользователей через JWT.', '# Техническое задание

## Название проекта

Pizza Ordering API

## Цель

Разработать REST API для оформления заказов пиццы с авторизацией пользователей через JWT.

## Технологический стек

### Backend

* Go 1.26.3
* net/http
* PostgreSQL 17.9
* Redis 8.6.3
* pgx
* go-redis
* golang-jwt

### Инфраструктура

* Docker
* Docker Compose

### Архитектура

Hexagonal Architecture (Ports & Adapters)

---

# Функциональные требования

## Авторизация

### Регистрация пользователя

**Endpoint**

POST /auth/register

**Request**

```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Validation**

* email обязателен
* email должен быть уникальным
* password обязателен
* password не менее 8 символов

**Response**

```json
{
  "message": "user registered"
}
```

---

### Авторизация пользователя

**Endpoint**

POST /auth/login

**Request**

```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response**

```json
{
  "access_token": "...",
  "refresh_token": "..."
}
```

---

### JWT

Использовать два типа токенов:

#### Access Token

* срок жизни: 15 минут
* используется для доступа к защищенным эндпоинтам

#### Refresh Token

* срок жизни: 7 дней
* хранится в Redis

---

## Заказы пиццы

### Создание заказа

**Endpoint**

POST /pizza/create/order

**Authorization**

Bearer Access Token

**Request**

```json
{
  "pizza_name": "Pepperoni",
  "size": "large",
  "quantity": 2
}
```

**Validation**

* pizza_name обязателен
* size обязателен
* допустимые значения:

  * small
  * medium
  * large
* quantity > 0

**Response**

```json
{
  "id": "uuid",
  "status": "created"
}
```

---

### Получение списка заказов

**Endpoint**

GET /pizza/orders

**Authorization**

Bearer Access Token

**Response**

```json
[
  {
    "id": "uuid",
    "pizza_name": "Pepperoni",
    "size": "large",
    "quantity": 2,
    "created_at": "2026-01-01T12:00:00Z"
  }
]
```

Пользователь должен видеть только свои заказы.

# Redis

Использовать Redis для хранения refresh токенов.

---

# Архитектура проекта

```text
cmd/
└── app/

internal/
├── domain/
│   ├── user.go
│   └── order.go
│
├── ports/
│   ├── repository.go
│   ├── service.go
│   └── jwt.go
│
├── application/
│   ├── auth/
│   └── pizza/
│
├── adapters/
│   ├── http/
│   │   ├── handlers/
│   │   └── middleware/
│   │
│   ├── postgres/
│   │   └── repositories/
│   │
│   ├── redis/
│   │   └── repositories/
│   │
│   └── jwt/
│
pkg/
configs/
migrations/
```

---

# Docker Compose

Необходимо подготовить docker-compose.yml, содержащий:

### Сервисы

#### app

* golang:1.26.3-alpine

#### postgres

* postgres:17.9-alpine3.23

Параметры:

```env
POSTGRES_DB=pizza
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
```

Порт:

```yaml
5432:5432
```

---

#### redis

* redis:8.6.3-alpine3.23

Порт:

```yaml
6379:6379
```

---

# Переменные окружения

```env
HTTP_PORT=8080

POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_DB=pizza
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres

REDIS_ADDR=redis:6379

JWT_SECRET=super-secret-key

ACCESS_TOKEN_TTL=15m
REFRESH_TOKEN_TTL=168h
```

---

# Нефункциональные требования

* Использовать context.Context во всех слоях.
* Использовать pgx для работы с PostgreSQL.
* Использовать go-redis для работы с Redis.
* Пароли хранить только в виде bcrypt-хеша.
* Возвращать корректные HTTP-коды ошибок.
* JSON-ответы должны быть единообразны.
* Код должен соответствовать принципам Hexagonal Architecture.
* Без использования сторонних HTTP-фреймворков (Gin, Echo, Fiber и т.д.).
* Использовать только net/http.

---

# Критерии приемки

Проект считается выполненным, если:

1. Пользователь может зарегистрироваться.
2. Пользователь может авторизоваться.
3. Генерируются access и refresh токены.
4. Refresh токен сохраняется в Redis.
5. Авторизованный пользователь может создать заказ пиццы.
6. Авторизованный пользователь может получить список своих заказов.
7. PostgreSQL и Redis запускаются через Docker Compose.
8. Приложение запускается через Docker Compose.
9. Код организован согласно Hexagonal Architecture.
10. Все эндпоинты корректно работают через Postman или curl.
11. Проверка на правильность загрузки переменных окружения.', 'backend', 'middle');

INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000031', id FROM tags WHERE name = 'REST';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000031', id FROM tags WHERE name = 'PostgreSQL';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000031', id FROM tags WHERE name = 'Testing';
INSERT INTO task_tags (task_id, tag_id) SELECT 'c1000000-0000-0000-0000-000000000031', id FROM tags WHERE name = 'API Design';

INSERT INTO task_languages (task_id, language_id) SELECT 'c1000000-0000-0000-0000-000000000031', id FROM languages WHERE name = 'Go';

-- +goose Down

DELETE FROM task_languages;
DELETE FROM task_tags;
DELETE FROM tasks;
DELETE FROM languages;
DELETE FROM tags;
