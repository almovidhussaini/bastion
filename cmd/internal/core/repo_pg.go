package core

import (
    "database/sql"
    "fmt"

    _ "github.com/jackc/pgx/v5/stdlib"
)

// NewPostgresRepos creates repos backed by PostgreSQL using the given DSN.
func NewPostgresRepos(dsn string) (*PostgresCommandRepo, *PostgresNodeRepo, *PostgresExecutionRepo, func(), error) {
    db, err := sql.Open("pgx", dsn)
    if err != nil {
        return nil, nil, nil, nil, fmt.Errorf("open db: %w", err)
    }
    if err := db.Ping(); err != nil {
        db.Close()
        return nil, nil, nil, nil, fmt.Errorf("ping db: %w", err)
    }
    if err := ensureSchema(db); err != nil {
        db.Close()
        return nil, nil, nil, nil, err
    }
    cleanup := func() {
        db.Close()
    }
    return &PostgresCommandRepo{db: db}, &PostgresNodeRepo{db: db}, &PostgresExecutionRepo{db: db}, cleanup, nil
}

func ensureSchema(db *sql.DB) error {
    stmts := []string{
        `CREATE TABLE IF NOT EXISTS commands (
            id TEXT PRIMARY KEY,
            name TEXT NOT NULL,
            description TEXT,
            script TEXT NOT NULL,
            timeout_seconds INTEGER NOT NULL,
            created_at TIMESTAMPTZ NOT NULL
        )`,
        `CREATE TABLE IF NOT EXISTS nodes (
            id TEXT PRIMARY KEY,
            name TEXT NOT NULL,
            address TEXT NOT NULL
        )`,
        `CREATE TABLE IF NOT EXISTS executions (
            id TEXT PRIMARY KEY,
            command_id TEXT NOT NULL,
            node_id TEXT NOT NULL,
            status TEXT NOT NULL,
            started_at TIMESTAMPTZ NOT NULL,
            completed_at TIMESTAMPTZ,
            stdout TEXT,
            stderr TEXT,
            exit_code INTEGER NOT NULL,
            duration_ms BIGINT NOT NULL,
            CONSTRAINT fk_command FOREIGN KEY (command_id) REFERENCES commands(id) ON DELETE CASCADE,
            CONSTRAINT fk_node FOREIGN KEY (node_id) REFERENCES nodes(id) ON DELETE CASCADE
        )`,
    }
    for _, stmt := range stmts {
        if _, err := db.Exec(stmt); err != nil {
            return fmt.Errorf("ensure schema: %w", err)
        }
    }
    return nil
}

type PostgresCommandRepo struct {
    db *sql.DB
}

func (r *PostgresCommandRepo) List() []Command {
    rows, err := r.db.Query(`SELECT id, name, description, script, timeout_seconds, created_at FROM commands ORDER BY created_at DESC`)
    if err != nil {
        return []Command{}
    }
    defer rows.Close()
    var out []Command
    for rows.Next() {
        var c Command
        var desc sql.NullString
        if err := rows.Scan(&c.ID, &c.Name, &desc, &c.Script, &c.TimeoutSeconds, &c.CreatedAt); err == nil {
            c.Description = desc.String
            out = append(out, c)
        }
    }
    return out
}

func (r *PostgresCommandRepo) Get(id string) (Command, bool) {
    var c Command
    var desc sql.NullString
    row := r.db.QueryRow(`SELECT id, name, description, script, timeout_seconds, created_at FROM commands WHERE id=$1`, id)
    if err := row.Scan(&c.ID, &c.Name, &desc, &c.Script, &c.TimeoutSeconds, &c.CreatedAt); err != nil {
        return Command{}, false
    }
    c.Description = desc.String
    return c, true
}

func (r *PostgresCommandRepo) Save(command Command) Command {
    _, _ = r.db.Exec(
        `INSERT INTO commands (id, name, description, script, timeout_seconds, created_at)
         VALUES ($1,$2,$3,$4,$5,$6)
         ON CONFLICT (id) DO UPDATE SET name=EXCLUDED.name, description=EXCLUDED.description, script=EXCLUDED.script, timeout_seconds=EXCLUDED.timeout_seconds`,
        command.ID, command.Name, command.Description, command.Script, command.TimeoutSeconds, command.CreatedAt,
    )
    return command
}

func (r *PostgresCommandRepo) Delete(id string) {
    _, _ = r.db.Exec(`DELETE FROM commands WHERE id=$1`, id)
}

type PostgresNodeRepo struct {
    db *sql.DB
}

func (r *PostgresNodeRepo) List() []Node {
    rows, err := r.db.Query(`SELECT id, name, address FROM nodes ORDER BY name`)
    if err != nil {
        return []Node{}
    }
    defer rows.Close()
    var out []Node
    for rows.Next() {
        var n Node
        if err := rows.Scan(&n.ID, &n.Name, &n.Address); err == nil {
            out = append(out, n)
        }
    }
    return out
}

func (r *PostgresNodeRepo) Get(id string) (Node, bool) {
    var n Node
    row := r.db.QueryRow(`SELECT id, name, address FROM nodes WHERE id=$1`, id)
    if err := row.Scan(&n.ID, &n.Name, &n.Address); err != nil {
        return Node{}, false
    }
    return n, true
}

func (r *PostgresNodeRepo) Save(node Node) Node {
    _, _ = r.db.Exec(
        `INSERT INTO nodes (id, name, address)
         VALUES ($1,$2,$3)
         ON CONFLICT (id) DO UPDATE SET name=EXCLUDED.name, address=EXCLUDED.address`,
        node.ID, node.Name, node.Address,
    )
    return node
}

type PostgresExecutionRepo struct {
    db *sql.DB
}

func (r *PostgresExecutionRepo) List() []Execution {
    rows, err := r.db.Query(`SELECT id, command_id, node_id, status, started_at, completed_at, stdout, stderr, exit_code, duration_ms FROM executions ORDER BY started_at DESC`)
    if err != nil {
        return []Execution{}
    }
    defer rows.Close()
    var out []Execution
    for rows.Next() {
        if exec, ok := scanExecution(rows); ok {
            out = append(out, exec)
        }
    }
    return out
}

func (r *PostgresExecutionRepo) Get(id string) (Execution, bool) {
    row := r.db.QueryRow(`SELECT id, command_id, node_id, status, started_at, completed_at, stdout, stderr, exit_code, duration_ms FROM executions WHERE id=$1`, id)
    exec, ok := scanExecution(row)
    return exec, ok
}

func (r *PostgresExecutionRepo) Save(execution Execution) Execution {
    var completedAt interface{}
    if execution.CompletedAt != nil {
        completedAt = *execution.CompletedAt
    }
    _, _ = r.db.Exec(
        `INSERT INTO executions (id, command_id, node_id, status, started_at, completed_at, stdout, stderr, exit_code, duration_ms)
         VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
         ON CONFLICT (id) DO UPDATE SET status=EXCLUDED.status, completed_at=EXCLUDED.completed_at, stdout=EXCLUDED.stdout, stderr=EXCLUDED.stderr, exit_code=EXCLUDED.exit_code, duration_ms=EXCLUDED.duration_ms`,
        execution.ID, execution.CommandID, execution.NodeID, string(execution.Status), execution.StartedAt, completedAt, execution.Stdout, execution.Stderr, execution.ExitCode, execution.DurationMs,
    )
    return execution
}

type scanner interface {
    Scan(dest ...interface{}) error
}

func scanExecution(row scanner) (Execution, bool) {
    var e Execution
    var completed sql.NullTime
    var status string
    if err := row.Scan(&e.ID, &e.CommandID, &e.NodeID, &status, &e.StartedAt, &completed, &e.Stdout, &e.Stderr, &e.ExitCode, &e.DurationMs); err != nil {
        return Execution{}, false
    }
    e.Status = ExecutionStatus(status)
    if completed.Valid {
        t := completed.Time
        e.CompletedAt = &t
    }
    return e, true
}
