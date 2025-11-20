package core

import (
    "bytes"
    "context"
    "crypto/rand"
    "encoding/hex"
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    "sort"
    "strings"
    "time"
)

type BastionService struct {
    commands   CommandRepository
    nodes      NodeRepository
    executions ExecutionRepository
    client     *http.Client
}

func NewBastionService(commands CommandRepository, nodes NodeRepository, executions ExecutionRepository) *BastionService {
    return &BastionService{
        commands:   commands,
        nodes:      nodes,
        executions: executions,
        client: &http.Client{
            Timeout: 60 * time.Second,
        },
    }
}

func (s *BastionService) ListCommands() []Command {
    return s.commands.List()
}

func (s *BastionService) GetCommand(id string) (Command, bool) {
    return s.commands.Get(id)
}

func (s *BastionService) ListNodes() []Node {
    return s.nodes.List()
}

func (s *BastionService) ListExecutions() []Execution {
    list := s.executions.List()
    sort.Slice(list, func(i, j int) bool {
        return list[i].StartedAt.After(list[j].StartedAt)
    })
    return list
}

func (s *BastionService) CreateCommand(input Command) (Command, error) {
    if strings.TrimSpace(input.Name) == "" {
        return Command{}, errors.New("name is required")
    }
    if strings.TrimSpace(input.Script) == "" {
        return Command{}, errors.New("script is required")
    }
    if input.TimeoutSeconds <= 0 {
        input.TimeoutSeconds = 300
    }
    input.ID = randomID("cmd")
    input.CreatedAt = time.Now().UTC()
    return s.commands.Save(input), nil
}

func (s *BastionService) DeleteCommand(id string) error {
    if strings.TrimSpace(id) == "" {
        return errors.New("id is required")
    }
    if _, ok := s.commands.Get(id); !ok {
        return fmt.Errorf("unknown command %s", id)
    }
    s.commands.Delete(id)
    return nil
}

func (s *BastionService) RegisterNode(node Node) Node {
    if node.ID == "" {
        node.ID = randomID("node")
    }
    return s.nodes.Save(node)
}

func (s *BastionService) ExecuteCommand(ctx context.Context, commandID, nodeID string) (Execution, error) {
    cmd, ok := s.commands.Get(commandID)
    if !ok {
        return Execution{}, fmt.Errorf("unknown command %s", commandID)
    }
    node, ok := s.nodes.Get(nodeID)
    if !ok {
        return Execution{}, fmt.Errorf("unknown node %s", nodeID)
    }

    now := time.Now().UTC()
    execRecord := Execution{
        ID:        randomID("exec"),
        CommandID: cmd.ID,
        NodeID:    node.ID,
        Status:    ExecutionRunning,
        StartedAt: now,
    }
    s.executions.Save(execRecord)

    req := ExecRequest{
        Script:         cmd.Script,
        TimeoutSeconds: cmd.TimeoutSeconds,
    }

    payload, err := json.Marshal(req)
    if err != nil {
        return s.failExecution(execRecord, fmt.Sprintf("marshal request: %v", err)), err
    }

    url := strings.TrimRight(node.Address, "/") + "/api/v1/exec"
    httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
    if err != nil {
        return s.failExecution(execRecord, fmt.Sprintf("build request: %v", err)), err
    }
    httpReq.Header.Set("Content-Type", "application/json")

    resp, err := s.client.Do(httpReq)
    if err != nil {
        return s.failExecution(execRecord, fmt.Sprintf("request failed: %v", err)), err
    }
    defer resp.Body.Close()

    var execResp ExecResponse
    if err := json.NewDecoder(resp.Body).Decode(&execResp); err != nil {
        return s.failExecution(execRecord, fmt.Sprintf("decode response: %v", err)), err
    }

    finished := time.Now().UTC()
    execRecord.Stdout = execResp.Stdout
    execRecord.Stderr = execResp.Stderr
    execRecord.ExitCode = execResp.ExitCode
    execRecord.DurationMs = execResp.DurationMs
    execRecord.CompletedAt = &finished
    if execResp.ExitCode == 0 {
        execRecord.Status = ExecutionSucceeded
    } else {
        execRecord.Status = ExecutionFailed
    }
    s.executions.Save(execRecord)
    return execRecord, nil
}

func (s *BastionService) failExecution(execRecord Execution, message string) Execution {
    now := time.Now().UTC()
    execRecord.CompletedAt = &now
    execRecord.Status = ExecutionFailed
    execRecord.Stderr = message
    execRecord.ExitCode = 1
    s.executions.Save(execRecord)
    return execRecord
}

func randomID(prefix string) string {
    b := make([]byte, 6)
    if _, err := rand.Read(b); err != nil {
        return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
    }
    return fmt.Sprintf("%s-%s", prefix, hex.EncodeToString(b))
}
