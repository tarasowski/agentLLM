package main

import (
  "bufio"
  "context"
  "fmt"
  "os"
  // add later other providers
  "github.com/anthropics/anthropic-sdk-go"
)

func main() {
  client := anthropic.NewClient()

  scanner := bufio.NewScanner(os.Stdin)
  getUserMessage := func() (string, bool) {
    if !scanner.Scan() {
      return "", false
    }
    return scanner.Text(), true
  }

  agent := NewAgent(&client, getUserMessage)
  err := agent.Run(context.TODO())
  if err != nil {
    fmt.Printf("Error: %s\n", err.Error())
  }
}

func NewAgent(client * anthropic.Client, getUserMessage func() (string, bool)) *Agent {
  return &Agent{
    client:        client,
    getUserMessage: getUserMessage,
  }
}

type Agent struct {
  client        *anthropic.Client
  getUserMessage func() (string, bool)
}
