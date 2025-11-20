package main

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "os"
    "strings"
    "time"

    "github.com/yourorg/boundless-bastion/cmd/internal/core"
    yamlloader "github.com/yourorg/boundless-bastion/cmd/internal/yaml"
)

type bastionServer struct {
    svc *core.BastionService
}

func main() {
    commandRepo := core.NewInMemoryCommandRepo()
    nodeRepo := core.NewInMemoryNodeRepo()
    execRepo := core.NewInMemoryExecutionRepo()
    svc := core.NewBastionService(commandRepo, nodeRepo, execRepo)

    daemonURL := envOr("DAEMON_URL", "http://localhost:9090")
    nodeRepo.Save(core.Node{ID: "node-local", Name: "Local Daemon", Address: daemonURL})

    if yamlPath := os.Getenv("COMMANDS_FILE"); yamlPath != "" {
        if cmds, err := yamlloader.LoadCommandsFromFile(yamlPath); err != nil {
            log.Printf("failed to load commands from %s: %v", yamlPath, err)
        } else {
            for _, c := range cmds {
                if _, err := svc.CreateCommand(c); err != nil {
                    log.Printf("skip command %s: %v", c.Name, err)
                }
            }
        }
    }

    if len(commandRepo.List()) == 0 {
        svc.CreateCommand(core.Command{
            Name:           "Check GPU",
            Description:    "Print GPU info with nvidia-smi",
            Script:         "nvidia-smi || echo 'nvidia-smi not available'",
            TimeoutSeconds: 60,
        })
        svc.CreateCommand(core.Command{
            Name:           "Docker ps",
            Description:    "List running containers",
            Script:         "docker ps",
            TimeoutSeconds: 60,
        })
    }

    srv := &bastionServer{svc: svc}
    mux := http.NewServeMux()
    mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK); w.Write([]byte("ok")) })
    mux.HandleFunc("/api/v1/commands", srv.handleCommands)
    mux.HandleFunc("/api/v1/nodes", srv.handleNodes)
    mux.HandleFunc("/api/v1/execute", srv.handleExecute)
    mux.HandleFunc("/api/v1/executions", srv.handleExecutions)
    mux.HandleFunc("/api/v1/gpu", srv.handleGPU)

    handler := withCORS(mux)

    port := envOr("BASTION_PORT", "8080")
    log.Printf("Bastion listening on :%s", port)
    if err := http.ListenAndServe(":"+port, handler); err != nil {
        log.Fatalf("server error: %v", err)
    }
}

func (s *bastionServer) handleCommands(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        if id := r.URL.Query().Get("id"); id != "" {
            if cmd, ok := s.svc.GetCommand(id); ok {
                writeJSON(w, http.StatusOK, cmd)
                return
            }
            http.Error(w, "not found", http.StatusNotFound)
            return
        }
        writeJSON(w, http.StatusOK, s.svc.ListCommands())
    case http.MethodPost:
        var payload struct {
            Name           string `json:"name"`
            Description    string `json:"description"`
            Script         string `json:"script"`
            TimeoutSeconds int    `json:"timeout_seconds"`
        }
        if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
            http.Error(w, "invalid payload", http.StatusBadRequest)
            return
        }
        cmd, err := s.svc.CreateCommand(core.Command{
            Name:           payload.Name,
            Description:    payload.Description,
            Script:         payload.Script,
            TimeoutSeconds: payload.TimeoutSeconds,
        })
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
        writeJSON(w, http.StatusCreated, cmd)
    case http.MethodDelete:
        id := r.URL.Query().Get("id")
        if id == "" {
            http.Error(w, "id is required", http.StatusBadRequest)
            return
        }
        if err := s.svc.DeleteCommand(id); err != nil {
            http.Error(w, err.Error(), http.StatusNotFound)
            return
        }
        w.WriteHeader(http.StatusNoContent)
    default:
        w.WriteHeader(http.StatusMethodNotAllowed)
    }
}

func (s *bastionServer) handleNodes(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }
    writeJSON(w, http.StatusOK, s.svc.ListNodes())
}

func (s *bastionServer) handleExecute(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }
    var payload struct {
        CommandID string `json:"command_id"`
        NodeID    string `json:"node_id"`
    }
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        http.Error(w, "invalid payload", http.StatusBadRequest)
        return
    }
    ctx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
    defer cancel()
    execRecord, err := s.svc.ExecuteCommand(ctx, payload.CommandID, payload.NodeID)
    if err != nil {
        log.Printf("execution error: %v", err)
    }
    writeJSON(w, http.StatusOK, execRecord)
}

func (s *bastionServer) handleExecutions(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }
    writeJSON(w, http.StatusOK, s.svc.ListExecutions())
}

func (s *bastionServer) handleGPU(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }
    nodes := s.svc.ListNodes()
    now := time.Now().UTC()
    samples := make([]core.GPUSample, 0, len(nodes))
    for _, n := range nodes {
        samples = append(samples, core.GPUSample{
            NodeID:      n.ID,
            Timestamp:   now.Unix(),
            Utilization: 20 + float64(len(n.ID))*5,
            MemoryMB:    2000 + len(n.ID)*256,
        })
    }
    writeJSON(w, http.StatusOK, samples)
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(payload)
}

func envOr(key, fallback string) string {
    if v := strings.TrimSpace(os.Getenv(key)); v != "" {
        return v
    }
    return fallback
}

func withCORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
        w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusNoContent)
            return
        }
        next.ServeHTTP(w, r)
    })
}
