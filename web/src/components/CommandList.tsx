import React, { useMemo, useState } from "react";
import {
  Alert,
  Button,
  Card,
  Form,
  Input,
  Modal,
  Popconfirm,
  Space,
  Table,
  Tag,
  Typography,
  message,
} from "antd";
import { EyeOutlined, PlayCircleOutlined, PlusOutlined, DeleteOutlined } from "@ant-design/icons";
import { Command, Node } from "../types";

interface Props {
  commands: Command[];
  nodes: Node[];
  selectedNodeId?: string;
  onRun: (commandId: string) => void;
  onCreate: (payload: {
    name: string;
    description: string;
    script: string;
    timeout_seconds?: number;
  }) => Promise<void> | void;
  onDelete: (id: string) => Promise<void> | void;
  loading?: boolean;
}

const CommandList: React.FC<Props> = ({
  commands,
  nodes,
  selectedNodeId,
  onRun,
  onCreate,
  onDelete,
  loading,
}) => {
  const [open, setOpen] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [viewing, setViewing] = useState<Command | null>(null);
  const [form] = Form.useForm();

  const nodeNames = useMemo(() => {
    const map: Record<string, string> = {};
    nodes.forEach((n) => (map[n.id] = n.name));
    return map;
  }, [nodes]);

  const columns = [
    {
      title: "Name",
      dataIndex: "name",
      key: "name",
      render: (text: string) => (
        <Typography.Text strong style={{ color: "#000" }}>
          {text}
        </Typography.Text>
      ),
    },
    {
      title: "Description",
      dataIndex: "description",
      key: "description",
      render: (text: string) => (
        <Typography.Text strong style={{ color: "#000" }}>
          {text}
        </Typography.Text>
      ),
    },
    {
      title: "Timeout",
      key: "timeout",
      render: (_: unknown, record: Command) => (
        <Tag color="blue" style={{ borderRadius: 12 }}>{`${record.timeout_seconds}s`}</Tag>
      ),
    },
    {
      title: "Actions",
      key: "actions",
      render: (_: unknown, record: Command) => (
        <Space>
          <Button icon={<EyeOutlined />} onClick={() => setViewing(record)}>
            View
          </Button>
          <Button
            icon={<PlayCircleOutlined />}
            type="primary"
            disabled={!selectedNodeId}
            onClick={() => onRun(record.id)}
          >
            Run
          </Button>
          <Popconfirm
            title="Delete command?"
            description="This will remove the command from Bastion."
            okText="Delete"
            okButtonProps={{ danger: true }}
            onConfirm={() => onDelete(record.id)}
          >
            <Button danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  const submit = async () => {
    try {
      const values = await form.validateFields();
      setSubmitting(true);
      await onCreate({
        name: values.name,
        description: values.description,
        script: values.script,
        timeout_seconds: values.timeout_seconds,
      });
      setOpen(false);
      form.resetFields();
    } catch (err) {
      console.error(err);
      message.error("Failed to create command");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Card className="table-card" title="Command Library" extra={
      <Button icon={<PlusOutlined />} type="dashed" onClick={() => setOpen(true)}>
        New Command
      </Button>
    }>
      {!selectedNodeId && (
        <Alert
          type="warning"
          message="Select a node to run commands"
          style={{ marginBottom: 12 }}
          showIcon
        />
      )}
      <Table
        rowKey="id"
        columns={columns}
        dataSource={commands}
        pagination={{ pageSize: 7 }}
        loading={loading}
      />
      <Modal
        title="Create Command"
        open={open}
        onCancel={() => setOpen(false)}
        onOk={submit}
        confirmLoading={submitting}
        width={720}
      >
        <Form layout="vertical" form={form} initialValues={{ timeout_seconds: 300 }}>
          <Form.Item
            name="name"
            label="Name"
            rules={[{ required: true, message: "Name is required" }]}
          >
            <Input placeholder="e.g. Restart service" />
          </Form.Item>
          <Form.Item name="description" label="Description">
            <Input placeholder="What does this command do?" />
          </Form.Item>
          <Form.Item
            name="script"
            label="Bash Script"
            rules={[{ required: true, message: "Script is required" }]}
          >
            <Input.TextArea rows={6} className="mono" placeholder="bash -lc commands" />
          </Form.Item>
          <Form.Item name="timeout_seconds" label="Timeout (seconds)">
            <Input type="number" min={1} />
          </Form.Item>
        </Form>
      </Modal>
      <Modal
        title={viewing ? viewing.name : "Command"}
        open={!!viewing}
        onCancel={() => setViewing(null)}
        footer={null}
        width={720}
      >
        {viewing && (
          <Space direction="vertical" size="middle" style={{ width: "100%" }}>
            <Typography.Text strong>Description:</Typography.Text>
            <Typography.Text>{viewing.description || "(none)"}</Typography.Text>
            <Typography.Text strong>Script:</Typography.Text>
            <div
              style={{
                background: "#0f172a",
                color: "#e2e8f0",
                padding: 12,
                borderRadius: 10,
                fontFamily: "JetBrains Mono, SFMono-Regular, Consolas, monospace",
                whiteSpace: "pre-wrap",
                border: "1px solid rgba(255,255,255,0.08)",
              }}
            >
              {viewing.script}
            </div>
            <Typography.Text>
              Timeout: <Tag color="blue">{viewing.timeout_seconds}s</Tag>
            </Typography.Text>
            <Typography.Text type="secondary">
              ID: {viewing.id} · Created:{" "}
              {viewing.created_at ? new Date(viewing.created_at).toLocaleString() : "n/a"}
            </Typography.Text>
          </Space>
        )}
      </Modal>
    </Card>
  );
};

export default CommandList;
