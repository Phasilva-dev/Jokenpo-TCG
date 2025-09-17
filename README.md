# Jokenpo 🎮✊✋✌️

Projeto desenvolvido em **Go** que implementa um jogo de **Pedra, Papel e Tesoura** com suporte a execução local e em containers via **Docker Compose**.
Inclui scripts de testes simples e mistos para validar a carga do sistema.

---

## 📂 Estrutura do Projeto

* `cmd/` → Contém o código-fonte principal (cliente e servidor).
* `internal/` → Implementação dos módulos internos (rede, lógica do jogo, etc.).
* `network/` → Pacote que implementa a logica de comunicação do servidor.
* `session/` → Pacote que implementa a logica do servidor, como criamos partida, como lidamos com a fila e etc.
* `game/` → Pacote que implementa a logica do game.
* `docker-compose.yml` → Configuração principal de containers.
* `docker-compose.simple-test.yml` → Compose para teste simples.
* `docker-compose.mixed-test.yml` → Compose para teste de carga mista.
* `simple_load_test.sh` → Script para rodar teste de carga simples.
* `mixed_load_test.sh` → Script para rodar teste de carga mista.
* `go.mod` / `go.sum` → Dependências Go.
* `LICENSE` → Licença do projeto.

---

## 🚀 Pré-requisitos

Antes de rodar o projeto, certifique-se de ter instalado:

* [Go 1.22+](https://go.dev/dl/)
* [Docker](https://docs.docker.com/get-docker/)
* [Docker Compose](https://docs.docker.com/compose/)

---

## ▶️ Executando Localmente

1. Clone o repositório ou extraia os arquivos:

   ```bash
   git clone <url-do-repo>
   cd Jokenpo
   ```

2. Compile o projeto:

   ```bash
   go build ./...
   ```

3. Execute o servidor:

   ```bash
   go run cmd/server/main.go
   ```

4. Em outro terminal, execute o cliente:

   ```bash
   go run cmd/client/main.go
   ```

> O cliente pedirá o IP do servidor. Se o servidor estiver em outra maquina, insira o ip dela dentro do cmd/client/main.go.

---

## 🐳 Executando com Docker

### Subir o servidor com Docker Compose

```bash
docker compose up --build
```

### Rodar cliente em container

Em outro terminal:

```bash
docker compose run client
```

---

## 🧪 Testes de Carga

* **Teste simples:**

  ```bash
  ./simple_load_test.sh
  ```

* **Teste misto:**

  ```bash
  ./mixed_load_test.sh
  ```

---

## 📖 Como Jogar

1. O cliente conecta ao servidor via **TCP**.
2. O jogador escolhe uma das inúmeras opções que o menu exibe.
3. O servidor processa a jogada e retorna o resultado.

---

## ⚖️ Licença

Este projeto é distribuído sob a licença MIT. Consulte o arquivo [LICENSE](LICENSE) para mais detalhes.

---


