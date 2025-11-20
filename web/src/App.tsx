import React, { useEffect, useMemo, useState } from "react";
import { Layout, Typography, Space, Select, Tabs, Button, message } from "antd";
import { ReloadOutlined, ThunderboltOutlined } from "@ant-design/icons";
import CommandList from "./components/CommandList";
import ExecutionTable from "./components/ExecutionTable";
import GpuChart from "./components/GpuChart";
import {
  createCommand,
  deleteCommand,
  fetchCommands,
  fetchExecutions,
  fetchGPU,
  fetchNodes,
  runCommand,
} from "./api";
import { Command, Execution, GpuSample, Node } from "./types";

const { Header, Content } = Layout;

function App() {
  const [commands, setCommands] = useState<Command[]>([]);
  const [nodes, setNodes] = useState<Node[]>([]);
  const [selectedNode, setSelectedNode] = useState<string>();
  const [executions, setExecutions] = useState<Execution[]>([]);
  const [gpu, setGpu] = useState<GpuSample[]>([]);
  const [loading, setLoading] = useState(false);

  const nodeOptions = useMemo(
    () => nodes.map((n) => ({ label: n.name, value: n.id })),
    [nodes]
  );

  useEffect(() => {
    refreshAll();
  }, []);

  const refreshAll = async () => {
    setLoading(true);
    try {
      const [cmds, nds, exes, gpuSamples] = await Promise.all([
        fetchCommands(),
        fetchNodes(),
        fetchExecutions(),
        fetchGPU(),
      ]);
      setCommands(cmds);
      setNodes(nds);
      setExecutions(exes);
      setGpu(gpuSamples);
      if (!selectedNode && nds.length > 0) {
        setSelectedNode(nds[0].id);
      }
    } catch (err) {
      console.error(err);
      message.error("Failed to load data from Bastion");
    } finally {
      setLoading(false);
    }
  };

  const refreshExecutionsOnly = async () => {
    try {
      const exes = await fetchExecutions();
      setExecutions(exes);
    } catch (err) {
      message.error("Failed to refresh executions");
    }
  };

  const handleRun = async (commandId: string) => {
    if (!selectedNode) {
      message.warning("Select a node first");
      return;
    }
    try {
      const exec = await runCommand(commandId, selectedNode);
      message.success(`Started on ${selectedNode} (exit ${exec.exit_code})`);
      await refreshExecutionsOnly();
    } catch (err) {
      console.error(err);
      message.error("Failed to trigger command");
    }
  };

  const handleCreate = async (payload: {
    name: string;
    description: string;
    script: string;
    timeout_seconds?: number;
  }) => {
    try {
      const cmd = await createCommand({
        ...payload,
        timeout_seconds: payload.timeout_seconds
          ? Number(payload.timeout_seconds)
          : undefined,
      });
      setCommands((prev) => [...prev, cmd]);
      message.success(`Created command ${cmd.name}`);
    } catch (err) {
      console.error(err);
      message.error("Failed to create command");
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await deleteCommand(id);
      setCommands((prev) => prev.filter((c) => c.id !== id));
      message.success("Deleted command");
    } catch (err) {
      console.error(err);
      message.error("Failed to delete command");
    }
  };

  return (
    <Layout style={{ minHeight: "100vh" }}>
      <Header>
        <Space align="center" style={{ width: "100%", justifyContent: "space-between" }}>
          <Space align="center">
            <ThunderboltOutlined style={{ color: "#00deb3", fontSize: 20 }} />
            <Typography.Title level={4} style={{ margin: 0, color: "#dfe6ee" }}>
              Boundless Bastion
            </Typography.Title>
          </Space>
          <Space align="center">
            <Typography.Text style={{ color: "#9fb2c3" }}>Node</Typography.Text>
            <Select
              style={{ minWidth: 220 }}
              options={nodeOptions}
              value={selectedNode}
              onChange={setSelectedNode}
              placeholder="Select node"
            />
            <Button icon={<ReloadOutlined />} onClick={refreshAll} loading={loading}>
              Reload
            </Button>
          </Space>
        </Space>
      </Header>
      <Content>
        <div className="panel">
          <Tabs
            defaultActiveKey="commands"
            tabBarStyle={{ color: "#ffffff" }}
            items={[
              {
                key: "commands",
                label: <span style={{ color: "#ffffff" }}>Commands</span>,
                children: (
                  <CommandList
                    commands={commands}
                    nodes={nodes}
                    selectedNodeId={selectedNode}
                    onRun={handleRun}
                    onCreate={handleCreate}
                    onDelete={handleDelete}
                    loading={loading}
                  />
                ),
              },
              {
                key: "executions",
                label: <span style={{ color: "#ffffff" }}>Executions</span>,
                children: (
                  <ExecutionTable
                    executions={executions}
                    commands={commands}
                    nodes={nodes}
                    onRefresh={refreshExecutionsOnly}
                  />
                ),
              },
              {
                key: "gpu",
                label: <span style={{ color: "#ffffff" }}>GPU Overview</span>,
                children: <GpuChart samples={gpu} nodes={nodes} onRefresh={refreshAll} loading={loading} />,
              },
            ]}
          />
        </div>
      </Content>
    </Layout>
  );
}

export default App;
