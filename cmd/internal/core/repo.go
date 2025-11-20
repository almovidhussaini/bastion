package core

import "sync"

type CommandRepository interface {
    List() []Command
    Get(id string) (Command, bool)
    Save(command Command) Command
}

type NodeRepository interface {
    List() []Node
    Get(id string) (Node, bool)
    Save(node Node) Node
}

type ExecutionRepository interface {
    List() []Execution
    Get(id string) (Execution, bool)
    Save(execution Execution) Execution
}

type InMemoryCommandRepo struct {
    mu   sync.RWMutex
    data map[string]Command
}

func NewInMemoryCommandRepo() *InMemoryCommandRepo {
    return &InMemoryCommandRepo{data: map[string]Command{}}
}

func (r *InMemoryCommandRepo) List() []Command {
    r.mu.RLock()
    defer r.mu.RUnlock()
    out := make([]Command, 0, len(r.data))
    for _, v := range r.data {
        out = append(out, v)
    }
    return out
}

func (r *InMemoryCommandRepo) Get(id string) (Command, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    v, ok := r.data[id]
    return v, ok
}

func (r *InMemoryCommandRepo) Save(command Command) Command {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.data[command.ID] = command
    return command
}

type InMemoryNodeRepo struct {
    mu   sync.RWMutex
    data map[string]Node
}

func NewInMemoryNodeRepo() *InMemoryNodeRepo {
    return &InMemoryNodeRepo{data: map[string]Node{}}
}

func (r *InMemoryNodeRepo) List() []Node {
    r.mu.RLock()
    defer r.mu.RUnlock()
    out := make([]Node, 0, len(r.data))
    for _, v := range r.data {
        out = append(out, v)
    }
    return out
}

func (r *InMemoryNodeRepo) Get(id string) (Node, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    v, ok := r.data[id]
    return v, ok
}

func (r *InMemoryNodeRepo) Save(node Node) Node {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.data[node.ID] = node
    return node
}

type InMemoryExecutionRepo struct {
    mu   sync.RWMutex
    data map[string]Execution
}

func NewInMemoryExecutionRepo() *InMemoryExecutionRepo {
    return &InMemoryExecutionRepo{data: map[string]Execution{}}
}

func (r *InMemoryExecutionRepo) List() []Execution {
    r.mu.RLock()
    defer r.mu.RUnlock()
    out := make([]Execution, 0, len(r.data))
    for _, v := range r.data {
        out = append(out, v)
    }
    return out
}

func (r *InMemoryExecutionRepo) Get(id string) (Execution, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    v, ok := r.data[id]
    return v, ok
}

func (r *InMemoryExecutionRepo) Save(execution Execution) Execution {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.data[execution.ID] = execution
    return execution
}
