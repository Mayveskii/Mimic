# Самоотладка MCP — Что мы нашли и пофиксили

**Дата:** 2026-05-18  
**Метод:** Запуск `mimic serve` → интерактивные вызовы инструментов → анализ реального вывода  
**Бюджет:** $0.01 (тесты), $0 (самоотладка)

---

## Найденные баги (реальные, не теоретические)

### BUG #1: SYS_FILE_EXISTS возвращал пустоту

**Как нашел:** Вызвал `SYS_FILE_EXISTS {"path": "README.md"}` → получил `""` (0 символов).

**Почему:** В `core/ops.c:895`:
```c
bool exists = (stat(path, &st) == 0);
(void)exists;  // ← результат выброшен!
return ERR_OK; // ← ничего не записано в packet->result
```

**Фикс:**
```c
snprintf(packet->result, sizeof(packet->result), "exists: %s", exists ? "true" : "false");
packet->result_len = strlen(packet->result);
```

**Результат:** Tier 1 стал 4/4 PASS вместо 3/4.

---

### BUG #2: IO_READ требует fd, не path

**Как нашел:** Вызвал `IO_READ {"path": "core/ops.c"}` → получил 91 символ (пустой fd).

**Почему:** IO_READ требует file descriptor (fd) от IO_OPEN. Модель не может сделать это в одном шаге.

**Решение:** Создан новый инструмент `SYS_FILE_READ`:
- Принимает `path`, `limit`, `offset`
- Внутри: `open() → read() → close()`
- Возвращает содержимое файла

**Результат:** Tier 3 теперь может читать файлы. Тест: `SYS_FILE_READ {"path": "core/ops.c"}` → 4095 символов.

---

### BUG #3: Tier 2 не работает (модель не делает циклы)

**Как нашел:** Попросил модель "проверить git status, затем прочитать Makefile" → модель вызвала только `GIT_STATUS`.

**Почему:** kimi k2.6 (и все LLM) делают **один** tool_call за запрос. Они не умеют внутренние циклы.

**Решение:** Создан `internal/orchestrator/loop.go` — external multi-turn coordinator:
```
User Request
    → Model → tool_call #1 → Execute → Result → Feed back to Model
    → Model → tool_call #2 → Execute → Result → Feed back to Model
    → ...repeat until final answer...
```

---

## Артефакты созданные при отладке

| Файл | Назначение |
|------|-----------|
| `test/battlefield/mcp_debug.py` | Интерактивный отладчик MCP сервера |
| `internal/orchestrator/loop.go` | Multi-turn coordinator для цепочек |
| `core/ops.c` | Новый инструмент `SYS_FILE_READ` (0x8A) |
| `internal/mcp/tool_schemas.go` | JSON Schema для `SYS_FILE_READ` |

---

## Проверка через отладчик

```bash
$ python3 test/battlefield/mcp_debug.py
mcp> call SYS_FILE_READ {"path": "README.md"}
📥 Result: { "content": [{ "text": "# Mimic — Deterministic...", ... }] }
📊 Output length: 3059 chars

mcp> call SYS_FILE_READ {"path": "core/ops.c", "limit": 4096}
📥 Result: { "content": [{ "text": "#include \"ops.h\"\n#include <stdio.h>...", ... }] }
📊 Output length: 4095 chars
```

---

## Следующий шаг

Для полного Tier 2/3 нужен multi-turn loop с реальным LLM API. Компонент создан (`loop.go`), но требует:
1. ModelCaller implementation for OpenRouter
2. Integration with benchmark framework
3. Real-world test with 5+ turn chains

Готов продолжить?
