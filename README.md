# go-battlesnake-server

![Go version](https://img.shields.io/badge/go-1.18-blue)
![License](https://img.shields.io/github/license/haguro/go-battlesnake-server)

`go-battlesnake-server` is a [Battlesnake](https://play.battlesnake.com) HTTP server written in Go. It builds upon the official [Go Starter Snake](https://github.com/BattlesnakeOfficial/starter-snake-go), with a couple of logging improvements and a slightly simpler setup, so you can quickly dive into creating your Battlesnake AI.

## Installation

You will need to have Go 1.18 or higher installed. You can install the package by running the following command:

```bash
go get github.com/your-username/go-battlesnake-server
```

## Usage

To get started, import the package into your Go code, define your move function and pass it to the server along with a [`log.Logger`](https://pkg.go.dev/log#Logger) and few other options. Below is a example **`main.go`** to demonstrate an example usage.

```go name: main.go
package main

import (
    "log"
    "os"

    "github.com/haguro/go-battlesnake-server/server"
)

const (
    version = "0.1"
    author  = "you"
    color   = "#33ccff"
    head    = "default"
    tail    = "default"
)

func main() {
    logOptions := server.LDefault
    if os.Getenv("DEBUG") == "TRUE" {
        logOptions |= server.LDebug
    }

    logger := log.New(os.Stderr, "[my_battlesnake]", log.LstdFlags|log.Lmicroseconds)
    logger.Fatal(
        server.New(
            "8080",
            &server.InfoResponse{
                Author:  author,
                Color:   color,
                Head:    head,
                Tail:    tail,
                Version: version,
            },
            logger,
            logOptions,
            moveSnake,
        ).Start(),
    )
}

func moveSnake(state *server.GameState, l *server.Logger) server.MoveResponse {
    l.Info("Hello from %s", state.You.Name)
    //TODO your snake AI here
    return server.MoveResponse{
        Move:  "left",
        Shout: "",
    }
}

```

Run with `DEBUG` environment variable set to `"TRUE"` to allow debug logging and write raw `GameState` json to the log file:

```bash
DEBUG="TRUE" go run main.go
```

Check out the [Battlesnake documentation](https://docs.battlesnake.com) for more information on implementing your Battlesnake AI.

## Contributing

Contributions are welcome! If you have any ideas, improvements, or bug fixes, please open an issue or submit a pull request.

## License

This project is licensed under the [MIT License](LICENSE).
