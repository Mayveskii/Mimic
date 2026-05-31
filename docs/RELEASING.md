# Releasing Mimic

## Branches

- `main` — стабильная ветка. Только через PR.
- `dev` — интеграция. Фича-ветки мёрджатся сюда.

## Pull Requests

1. Каждый PR требует зелёный CI (`lint`, `build`, `test`, `security`)
2. PR в `main` требует review
3. Не мёрджить если хоть один job красный

## Versioning

- Semantic versioning: `vMAJOR.MINOR.PATCH`
- Тег создаётся вручную после мёрджа в `main`
- Тег нельзя пересоздавать. Ошибка → новый тег

## Release Process

1. Убедиться что `main` зелёный
2. Создать annotated tag:
   ```bash
   git tag -a v0.x.x -m "Release v0.x.x"
   git push origin v0.x.x
   ```
3. GoReleaser автоматически:
   - Соберёт `linux/amd64` бинарник
   - Создаст GitHub Release
   - Опубликует Docker образ в GHCR

## Data Pipeline

- Раз в неделю (или вручную) — workflow `data.yml`
- Создаёт ветку `auto/data-sync-YYYYMMDD`
- Открывает PR в `dev`
- Ты решаешь: merge или close
- Никаких автокоммитов в `main`

## npm

- Не используется в текущих релизах
- Будет добавлено когда Mimic готов к широкому выпуску
