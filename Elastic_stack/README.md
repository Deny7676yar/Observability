# EFK

Т.к. elasticsearch запускается от имени пользователя с id `1000`, заранее создадим папку `elasticsearch` и поменяем ее владельцев:

```bash
mkdir -p elasticsearch && \
    sudo chown -R 1000:1000 elasticsearch
```

Посмотрим на информацию о кластере Elasticsearch. Для начала выберем все существующие индексы:

```bash
http://localhost:9200/_cat/indices
```

Посмотрим на настройки для индекса `fluentd`:

```bash
http://localhost:9200/<здесь-имя-индекса>/_settings
```

Посмотрим на общее состояние кластера:

```bash
http://localhost:9200/_cluster/health
```

Что наблюдаем? Почему?

Поправим это:

```bash
curl -XPUT -H "Content-Type: application/json" -d '{"index" : {"number_of_replicas" : 1 }}' localhost:9200/<здесь-имя-индекса>/_settings
```


Документация про другие API находится здесь: https://www.elastic.co/guide/en/elasticsearch/reference/current/rest-apis.html

# Sentry

```bash
# скачиваем репозиторий с Sentry
git clone https://github.com/getsentry/onpremise.git

# переходим в скачанный репозиторий
cd onpremise

# запускаем предварительную настройку
## в процессе установки будет запрос на создание аккаунта для вашего сервера
./install.sh

# запускаем Sentry
docker-compose up -d
```

Создадим Python проект.

Будем запускать код примера отсюда: https://github.com/getsentry/examples

Перед запуском установим зависимости:

```bash
pip install flask
pip install raven
pip install --upgrade 'sentry-sdk[flask]'
```

Свяжем приложение с нашей инсталяцией Sentry, для этого передадим DSN из проекта в код.

Запустим сервер:

```bash
python3 app.py
```