import React, { useMemo, useState } from "react";
import { Button, Modal, Space, Table, Tag, Typography, Tabs } from "antd";
import { EyeOutlined, ReloadOutlined } from "@ant-design/icons";
import { Command, Execution, Node } from "../types";

interface Props {
  executions: Execution[];
  commands: Command[];
  nodes: Node[];
  onRefresh?: () => void;
}

const ExecutionTable: React.FC<Props> = ({ executions, commands, nodes, onRefresh }) => {
  const [selected, setSelected] = useState<Execution | null>(null);

  const commandLookup = useMemo(() => {
    const map: Record<string, string> = {};
    commands.forEach((c) => (map[c.id] = c.name));
    return map;
  }, [commands]);

  const nodeLookup = useMemo(() => {
    const map: Record<string, string> = {};
    nodes.forEach((n) => (map[n.id] = n.name));
    return map;
  }, [nodes]);

  const statusColor = (status: Execution["status"]): string => {
    switch (status) {
      case "running":
        return "processing";
      case "succeeded":
        return "success";
      case "failed":
        return "error";
      default:
        return "default";
    }
  };

  const columns = [
    {
      title: "Command",
      dataIndex: "command_id",
      render: (id: string) => commandLookup[id] || id,
    },
    {
      title: "Node",
      dataIndex: "node_id",
      render: (id: string) => nodeLookup[id] || id,
    },
    {
      title: "Status",
      dataIndex: "status",
      render: (status: Execution["status"]) => (
        <Tag color={statusColor(status)} style={{ textTransform: "capitalize" }}>
          {status}
        </Tag>
      ),
    },
    {
      title: "Exit",
      dataIndex: "exit_code",
    },
    {
      title: "Duration",
      render: (_: unknown, record: Execution) => `${record.duration_ms} ms`,
    },
    {
      title: "Started",
      dataIndex: "started_at",
      render: (text: string) => new Date(text).toLocaleString(),
    },
    {
      title: "Completed",
      dataIndex: "completed_at",
      render: (text?: string) => (text ? new Date(text).toLocaleString() : "-"),
    },
    {
      title: "Logs",
      key: "logs",
      render: (_: unknown, record: Execution) => (
        <Button icon={<EyeOutlined />} onClick={() => setSelected(record)}>
          View
        </Button>
      ),
    },
  ];

  return (
    <div className="table-card">
      <Space style={{ marginBottom: 12 }}>
        <Typography.Title level={5} style={{ margin: 0, color: "#dfe6ee" }}>
          Execution History
        </Typography.Title>
        <Button icon={<ReloadOutlined />} onClick={onRefresh} size="small">
          Refresh
        </Button>
      </Space>
      <Table rowKey="id" dataSource={executions} columns={columns} pagination={{ pageSize: 8 }} />
      <Modal
        open={!!selected}
        onCancel={() => setSelected(null)}
        width={900}
        title={`Logs for ${commandLookup[selected?.command_id || ""] || selected?.command_id || ""}`}
        footer={null}
      >
        {selected && (
          <Tabs
            items={[
              {
                key: "stdout",
                label: "STDOUT",
                children: <LogBlock text={selected.stdout} />,
              },
              {
                key: "stderr",
                label: "STDERR",
                children: <LogBlock text={selected.stderr} danger />, 
              },
            ]}
          />
        )}
      </Modal>
    </div>
  );
};

const LogBlock: React.FC<{ text: string; danger?: boolean }> = ({ text, danger }) => (
  <div
    style={{
      background: danger ? "#1f1317" : "#0f172a",
      color: danger ? "#ff9aa2" : "#e2e8f0",
      padding: 12,
      borderRadius: 10,
      minHeight: 160,
      fontFamily: "JetBrains Mono, SFMono-Regular, Consolas, monospace",
      whiteSpace: "pre-wrap",
      border: "1px solid rgba(255,255,255,0.08)",
    }}
  >
    {text || "(empty)"}
  </div>
);

export default ExecutionTable;
