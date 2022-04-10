# pprof

Запустим генератор нагрузки и сразу после pprof на 5 секунд:

```bash
go tool pprof -svg http://localhost:9000/debug/pprof/profile\?seconds\=5 > pprof/pprof.svg
```

Мы получили профиль использования CPU. Получим профиль использования памяти:

```bash
go tool pprof -svg http://localhost:9000/debug/pprof/heap\?seconds\=5 > pprof/pprof-heap.svg
```

# Трейсинг

Запустим генератор нагрузки и начнем собирать данные трейсинга:

```bash
wget -O pprof/trace.out http://localhost:9000/debug/pprof/trace\?seconds\=5
```

После сбора данных запустим анализ:

```bash
go tool trace pprof/trace.out
```

# Сравнение двух профилей:

Запустим профилирование для приложения, неиспользующего кэш:

```bash
go tool pprof http://localhost:9000/debug/pprof/profile\?seconds\=5
```

В stdout будет записано имя файла, содержащего полученный профиль.

Запустим профиль для приложения с кэшированием и передадим имя предыдущего профиля для его сравнения:

```bash
go tool pprof -base <имя-файла-с-профилем> -svg http://localhost:9000/debug/pprof/profile\?seconds\=5 > pprof/pprof-compare.svg
```