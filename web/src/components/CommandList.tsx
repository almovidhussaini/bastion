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
import {
  EyeOutlined,
  PlayCircleOutlined,
  PlusOutlined,
  DeleteOutlined,
  EditOutlined,
} from "@ant-design/icons";
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
  onUpdate: (
    id: string,
    payload: {
      name: string;
      description: string;
      script: string;
      timeout_seconds?: number;
    }
  ) => Promise<void> | void;
  onDelete: (id: string) => Promise<void> | void;
  loading?: boolean;
}

const CommandList: React.FC<Props> = ({
  commands,
  nodes,
  selectedNodeId,
  onRun,
  onCreate,
  onUpdate,
  onDelete,
  loading,
}) => {
  const [open, setOpen] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [viewing, setViewing] = useState<Command | null>(null);
  const [editing, setEditing] = useState<Command | null>(null);
  const [updating, setUpdating] = useState(false);
  const [form] = Form.useForm();
  const [editForm] = Form.useForm();

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
        <Typography.Text strong style={{ color: "#e2e8f0" }}>
          {text}
        </Typography.Text>
      ),
    },
    {
      title: "Description",
      dataIndex: "description",
      key: "description",
      render: (text: string) => (
        <Typography.Text strong style={{ color: "#e2e8f0" }}>
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
      align: "center" as const,
      render: (_: unknown, record: Command) => (
        <Space align="center">
          <Button icon={<EyeOutlined />} onClick={() => setViewing(record)}>
            View
          </Button>
          <Button
            icon={<EditOutlined />}
            onClick={() => {
              setEditing(record);
              editForm.setFieldsValue({
                name: record.name,
                description: record.description,
                script: record.script,
                timeout_seconds: record.timeout_seconds,
              });
            }}
          >
            Edit
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

  const submitUpdate = async () => {
    if (!editing) return;
    try {
      const values = await editForm.validateFields();
      setUpdating(true);
      await onUpdate(editing.id, {
        name: values.name,
        description: values.description,
        script: values.script,
        timeout_seconds: values.timeout_seconds,
      });
      setEditing(null);
      editForm.resetFields();
    } catch (err) {
      console.error(err);
      message.error("Failed to update command");
    } finally {
      setUpdating(false);
    }
  };

  return (
    <Card
      className="table-card"
      title={
        <Typography.Text strong style={{ color: "#fff" }}>
          Command Library
        </Typography.Text>
      }
      extra={
        <Button icon={<PlusOutlined />} type="dashed" onClick={() => setOpen(true)}>
          New Command
        </Button>
      }
    >
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
        bordered
        size="middle"
        tableLayout="fixed"
        scroll={{ x: true }}
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
        title={editing ? `Edit ${editing.name}` : "Edit Command"}
        open={!!editing}
        onCancel={() => {
          setEditing(null);
          editForm.resetFields();
        }}
        onOk={submitUpdate}
        confirmLoading={updating}
        width={720}
      >
        <Form layout="vertical" form={editForm}>
          <Form.Item
            name="name"
            label="Name"
            rules={[{ required: true, message: "Name is required" }]}
          >
            <Input placeholder="Command name" />
          </Form.Item>
          <Form.Item name="description" label="Description">
            <Input placeholder="Description" />
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
              ID: {viewing.id} - Created:{" "}
              {viewing.created_at ? new Date(viewing.created_at).toLocaleString() : "n/a"}
            </Typography.Text>
          </Space>
        )}
      </Modal>
    </Card>
  );
};

export default CommandList;
