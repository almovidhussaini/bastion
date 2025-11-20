export type ExecutionStatus = "pending" | "running" | "succeeded" | "failed";

export interface Command {
  id: string;
  name: string;
  description: string;
  script: string;
  timeout_seconds: number;
  created_at: string;
}

export interface Node {
  id: string;
  name: string;
  address: string;
}

export interface Execution {
  id: string;
  command_id: string;
  node_id: string;
  status: ExecutionStatus;
  started_at: string;
  completed_at?: string;
  stdout: string;
  stderr: string;
  exit_code: number;
  duration_ms: number;
}

export interface GpuSample {
  node_id: string;
  timestamp: number;
  utilization: number;
  memory_mb: number;
}
