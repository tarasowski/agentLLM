package main

import (
  "bufio"
  "context"
  "fmt"
  "os"
  "encoding/json"
  // add later other providers
  "github.com/anthropics/anthropic-sdk-go"
  "github.com/invopop/jsonschema"
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

  tools := []ToolDefinition{ReadFileDefinition}
  agent := NewAgent(&client, getUserMessage, tools)
  err := agent.Run(context.TODO())
  if err != nil {
    fmt.Printf("Error: %s\n", err.Error())
  }
}

func NewAgent(client * anthropic.Client, 
              getUserMessage func() (string, bool),
              tools []ToolDefinition,
            ) *Agent {
  return &Agent{
    client:        client,
    getUserMessage: getUserMessage,
  }
}

type Agent struct {
  client        *anthropic.Client
  getUserMessage func() (string, bool)
  tools        []ToolDefinition
}

func (a *Agent) Run(ctx context.Context) error {
  conversation := []anthropic.MessageParam{}

  fmt.Println("Chat with Claude (use 'ctrl-c' to quit)")

  for {
    fmt.Print("\u001b[94mYou\u001b[0m: ")
    userInput, ok := a.getUserMessage()
    if !ok {
      break
    }

    userMessage := anthropic.NewUserMessage(anthropic.NewTextBlock(userInput))
    conversation = append(conversation, userMessage)

    message, err  := a.runInference(ctx, conversation)
    if err != nil {
      return err
    }

    conversation = append(conversation, message.ToParam())

    for _, content :=range message.Content{
      switch content.Type{
      case "text":
        fmt.Printf("\u001b[92mClaude\u001b[0m: %s\n", content.Text)
      }
    }

  }
  return nil
}

func (a *Agent) runInference(ctx context.Context, conversation []anthropic.MessageParam) (*anthropic.Message, error) {
  anthropicTools := []anthropic.ToolUnionParam{}

  for _, tool := range a.tools {
    anthropicTools = append(anthropicTools, anthropic.ToolUnionParam{
      OfTool: &anthropic.ToolParam{
        Name:         tool.Name,
        Description:  anthropic.String(tool.Description),
        InputSchema:  tool.InputSchema,
      },
    })
  }

  message, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
    Model: anthropic.ModelClaude3_7SonnetLatest,
    MaxTokens: int64(1024),
    Messages: conversation,
    Tools: anthropicTools,
  })
  return message, err
}


type ToolDefinition struct {
  Name            string `json:"name"`
  Description     string `json:"description"`
  InputSchema     anthropic.ToolInputSchemaParam `json:"input_schema"`
  Function        func(input json.RawMessage) (string, error)
}


var ReadFileDefinition = ToolDefinition{
  Name:       "read_file",
  Description: "Read a file of a given relative file path",
  InputSchema: ReadFileInputSchema,
  Function: ReadFile,
}

type ReadFileInput struct {
  Path string `json:"path" jsonschema_description: "The relative path to the file to read"`
}

var ReadFileInputSchema = GenerateSchema[ReadFileInput]()

func ReadFile(input json.RawMessage) (string, error) {
  readFileInput := ReadFileInput{}
  err := json.Unmarshal(input, &readFileInput)
  if  err != nil {
    panic(err)
  }

  content, err := os.ReadFile(readFileInput.Path)

  if err != nil {
    return "", err
  }

  return string(content), nil

}

func GenerateSchema[T any]() anthropic.ToolInputSchemaParam {
  reflector := jsonschema.Reflector{
    AllowAdditionalProperties: false,
    DoNotReference: true,
  }
  var v T

  schema := reflector.Reflect(v)

  return anthropic.ToolInputSchemaParam{
    Properties: schema.Properties,
  }
}
























