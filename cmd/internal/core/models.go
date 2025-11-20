package core

import "time"

type Command struct {
    ID             string    `json:"id"`
    Name           string    `json:"name"`
    Description    string    `json:"description"`
    Script         string    `json:"script"`
    TimeoutSeconds int       `json:"timeout_seconds"`
    CreatedAt      time.Time `json:"created_at"`
}

type Node struct {
    ID      string `json:"id"`
    Name    string `json:"name"`
    Address string `json:"address"`
}

type ExecutionStatus string

const (
    ExecutionPending  ExecutionStatus = "pending"
    ExecutionRunning  ExecutionStatus = "running"
    ExecutionSucceeded ExecutionStatus = "succeeded"
    ExecutionFailed   ExecutionStatus = "failed"
)

type Execution struct {
    ID          string          `json:"id"`
    CommandID   string          `json:"command_id"`
    NodeID      string          `json:"node_id"`
    Status      ExecutionStatus `json:"status"`
    StartedAt   time.Time       `json:"started_at"`
    CompletedAt *time.Time      `json:"completed_at,omitempty"`
    Stdout      string          `json:"stdout"`
    Stderr      string          `json:"stderr"`
    ExitCode    int             `json:"exit_code"`
    DurationMs  int64           `json:"duration_ms"`
}

type ExecRequest struct {
    Script         string `json:"script"`
    TimeoutSeconds int    `json:"timeout_seconds"`
    WorkingDir     string `json:"working_dir,omitempty"`
}

type ExecResponse struct {
    Stdout     string `json:"stdout"`
    Stderr     string `json:"stderr"`
    ExitCode   int    `json:"exit_code"`
    DurationMs int64  `json:"duration_ms"`
}

type GPUSample struct {
    NodeID       string  `json:"node_id"`
    Timestamp    int64   `json:"timestamp"`
    Utilization  float64 `json:"utilization"`
    MemoryMB     int     `json:"memory_mb"`
}
