### Documentação de Utilização do Projeto

Este projeto implementa um rate limiter utilizando Redis para persistência. Abaixo estão as instruções para configurar, executar e utilizar o projeto.

### Pré-requisitos

- [Go](https://golang.org/doc/install) 1.22 ou superior
- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)

### Configuração

1. Clone o repositório:

```sh
git clone https://github.com/seu-usuario/rate-limiter.git
cd rate-limiter
```

2. Crie um arquivo .env na raiz do projeto com o seguinte conteúdo:

```sh
REDIS_ADDR=redis:6379
REDIS_PASSWORD=
REDIS_DB=0
IP_RATE_LIMIT=3
TOKEN_RATE_LIMIT=5
BLOCK_DURATION=30 # in seconds
RATE_LIMIT_WINDOW=3 # in seconds
```

### Executando com Docker Compose

1. Certifique-se de que o Docker e o Docker Compose estão instalados e funcionando corretamente.

2. Execute o comando abaixo para iniciar os serviços:

```sh
docker-compose up --build
```

Isso irá iniciar o Redis e o servidor Go.

### Utilização

O servidor estará disponível na porta `8080`. Você pode testar o rate limiter utilizando `curl`.

#### Exemplo de Requisição sem Token

```sh
curl -i http://localhost:8080/
```

#### Exemplo de Requisição com Token

```sh
curl -i -H "API_KEY: seu_token_aqui" http://localhost:8080/
```

#### Exemplo de Resposta

```http
HTTP/1.1 200 OK
X-RateLimit-Limit: 3
X-RateLimit-Remaining: 2
X-RateLimit-Reset: 3
Content-Length: 13
Content-Type: text/plain; charset=utf-8

Hello, World!
```

#### Exemplo de Resposta ao Exceder o Limite

```http
HTTP/1.1 429 Too Many Requests
Content-Length: 83
Content-Type: text/plain; charset=utf-8

you have reached the maximum number of requests or actions allowed within a certain time frame
```