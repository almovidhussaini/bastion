package yamlloader

import (
    "fmt"
    "os"

    "gopkg.in/yaml.v3"

    "github.com/yourorg/boundless-bastion/cmd/internal/core"
)

type commandDocument struct {
    Name           string `yaml:"name"`
    Description    string `yaml:"description"`
    Script         string `yaml:"script"`
    TimeoutSeconds int    `yaml:"timeout_seconds"`
}

func LoadCommandsFromFile(path string) ([]core.Command, error) {
    raw, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("read yaml: %w", err)
    }

    var docs []commandDocument
    if err := yaml.Unmarshal(raw, &docs); err != nil {
        return nil, fmt.Errorf("parse yaml: %w", err)
    }

    commands := make([]core.Command, 0, len(docs))
    for _, d := range docs {
        commands = append(commands, core.Command{
            Name:           d.Name,
            Description:    d.Description,
            Script:         d.Script,
            TimeoutSeconds: d.TimeoutSeconds,
        })
    }
    return commands, nil
}
