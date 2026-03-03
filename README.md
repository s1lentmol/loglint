# loglint

`loglint` — кастомный линтер для Go с интеграцией как module plugin для `golangci-lint`.

## Сборка плагина для golangci-lint

1. Установить `golangci-lint` с командой `custom`:

```bash
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
```

2. Подготовить конфиг сборки:

```bash
cp .custom-gcl.example.yml .custom-gcl.yml
```

3. Собрать кастомный бинарь:

```bash
golangci-lint custom
```

После этого появится `./custom-gcl`.

## Запуск

1. Запустить проверку с примерным конфигом:

```bash
./custom-gcl run -c .golangci.example.yml ./...
```

## Примеры использования

Проверить только один пакет:

```bash
./custom-gcl run -c .golangci.example.yml ./internal/analyzer
```

Проверить несколько пакетов:

```bash
./custom-gcl run -c .golangci.example.yml ./cmd/... ./internal/...
```

Запустить обычный standalone-анализатор без golangci-lint:

```bash
go run ./cmd/loglint ./...
```

## Конфигурация (`.loglint.yml`)

Приоритет источников конфигурации:

1. Явный путь `-config` (standalone) или `settings.config` (golangci plugin).
2. Автопоиск `.loglint.yml` в текущей рабочей директории.
3. Встроенные значения по умолчанию.

Пример:

```yaml
version: 1

rules:
  lowercase_start: true
  english_only: true
  no_special_chars: true
  no_sensitive_data: true

sensitive:
  mode: append # append | override
  keywords:
    - sessionid
    - private_key

ignore:
  paths:
    - "vendor/**"
    - "**/*_generated.go"
```

Готовый шаблон: [`.loglint.example.yml`](.loglint.example.yml)

Standalone с явным конфигом:

```bash
go run ./cmd/loglint -config ./.loglint.yml ./...
```

Фрагмент для `.golangci.yml`:

```yaml
linters:
  enable: [loglint]
  settings:
    custom:
      loglint:
        type: module
        settings:
          config: ".loglint.yml"
```
