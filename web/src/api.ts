import axios from "axios";
import { Command, Execution, GpuSample, Node } from "./types";

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE || "http://localhost:8080",
});

export async function fetchCommands(): Promise<Command[]> {
  const res = await api.get<Command[]>("/api/v1/commands");
  return res.data;
}

export async function fetchNodes(): Promise<Node[]> {
  const res = await api.get<Node[]>("/api/v1/nodes");
  return res.data;
}

export async function fetchExecutions(): Promise<Execution[]> {
  const res = await api.get<Execution[]>("/api/v1/executions");
  return res.data;
}

export async function createCommand(payload: {
  name: string;
  description: string;
  script: string;
  timeout_seconds?: number;
}): Promise<Command> {
  const res = await api.post<Command>("/api/v1/commands", payload);
  return res.data;
}

export async function runCommand(commandId: string, nodeId: string): Promise<Execution> {
  const res = await api.post<Execution>("/api/v1/execute", {
    command_id: commandId,
    node_id: nodeId,
  });
  return res.data;
}

export async function fetchGPU(): Promise<GpuSample[]> {
  const res = await api.get<GpuSample[]>("/api/v1/gpu");
  return res.data;
}

export { api };
