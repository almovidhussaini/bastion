package main

import (
    "bytes"
    "context"
    "encoding/json"
    "errors"
    "log"
    "net/http"
    "os"
    "os/exec"
    "strings"
    "time"

    "github.com/yourorg/boundless-bastion/cmd/internal/core"
)

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("ok")) })
    mux.HandleFunc("/api/v1/exec", handleExec)

    port := envOr("DAEMON_PORT", "9081")
    log.Printf("Daemon listening on :%s", port)
    if err := http.ListenAndServe("127.0.0.1:"+port, mux); err != nil {
        log.Fatalf("daemon error: %v", err)
    }
}

func handleExec(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }
    var req core.ExecRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid payload", http.StatusBadRequest)
        return
    }
    timeout := req.TimeoutSeconds
    if timeout <= 0 {
        timeout = 300
    }

    ctx := r.Context()
    if timeout > 0 {
        var cancel context.CancelFunc
        ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
        defer cancel()
    }

    start := time.Now()
    stdout, stderr, exitCode, err := runScript(ctx, req.Script, req.WorkingDir)
    duration := time.Since(start)

    resp := core.ExecResponse{
        Stdout:     stdout,
        Stderr:     stderr,
        ExitCode:   exitCode,
        DurationMs: duration.Milliseconds(),
    }

    if err != nil {
        log.Printf("exec error: %v", err)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}

func runScript(ctx context.Context, script string, workdir string) (string, string, int, error) {
    if strings.TrimSpace(script) == "" {
        return "", "empty script", 1, errors.New("empty script")
    }
    cmd := exec.CommandContext(ctx, "bash", "-lc", script)
    if workdir != "" {
        cmd.Dir = workdir
    }

    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    err := cmd.Run()
    exitCode := 0
    if err != nil {
        if exitErr, ok := err.(*exec.ExitError); ok {
            exitCode = exitErr.ExitCode()
        } else {
            exitCode = 1
        }
        if stderr.Len() == 0 {
            // Surface process start failures (e.g., missing shell) to the caller.
            stderr.WriteString(err.Error())
        }
    }
    return stdout.String(), stderr.String(), exitCode, err
}

func envOr(key, fallback string) string {
    if v := strings.TrimSpace(os.Getenv(key)); v != "" {
        return v
    }
    return fallback
}
