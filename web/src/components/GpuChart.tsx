import React, { useMemo } from "react";
import { Button, Card, Space } from "antd";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  Legend,
  ResponsiveContainer,
  CartesianGrid,
} from "recharts";
import { GpuSample, Node } from "../types";

interface Props {
  samples: GpuSample[];
  nodes: Node[];
  onRefresh?: () => void;
  loading?: boolean;
}

const GpuChart: React.FC<Props> = ({ samples, nodes, onRefresh, loading }) => {
  const data = useMemo(() => {
    const grouped: Record<number, any> = {};
    samples.forEach((s) => {
      if (!grouped[s.timestamp]) {
        grouped[s.timestamp] = { timestamp: s.timestamp };
      }
      grouped[s.timestamp][`${s.node_id}_util`] = s.utilization;
      grouped[s.timestamp][`${s.node_id}_mem`] = s.memory_mb;
    });
    return Object.values(grouped).sort((a, b) => a.timestamp - b.timestamp);
  }, [samples]);

  return (
    <Card
      className="table-card"
      title={<span style={{ color: "#fff", fontWeight: 600 }}>GPU Overview</span>}
      extra={<Button onClick={onRefresh} loading={loading}>Refresh</Button>}
    >
      <ResponsiveContainer width="100%" height={320}>
        <LineChart data={data} margin={{ top: 16, right: 24, left: 0, bottom: 0 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="#233046" />
          <XAxis
            dataKey="timestamp"
            tickFormatter={(v) => new Date(v * 1000).toLocaleTimeString([], { hour12: false })}
            stroke="#6c819d"
          />
          <YAxis yAxisId="left" stroke="#6c819d" domain={[0, 100]} tickFormatter={(v) => `${v}%`} />
          <YAxis yAxisId="right" orientation="right" stroke="#6c819d" tickFormatter={(v) => `${v}MB`} />
          <Tooltip
            contentStyle={{ background: "#0d141f", color: "#dfe6ee", border: "1px solid #1f2a3f" }}
            labelFormatter={(v) => new Date(Number(v) * 1000).toLocaleString()}
          />
          <Legend />
          {nodes.map((n) => (
            <Line
              key={`${n.id}-util`}
              yAxisId="left"
              type="monotone"
              dataKey={`${n.id}_util`}
              name={`${n.name} util %`}
              stroke="#00deb3"
              dot={false}
              strokeWidth={2}
            />
          ))}
          {nodes.map((n) => (
            <Line
              key={`${n.id}-mem`}
              yAxisId="right"
              type="monotone"
              dataKey={`${n.id}_mem`}
              name={`${n.name} mem MB`}
              stroke="#94b3ff"
              dot={false}
              strokeDasharray="4 2"
              strokeWidth={2}
            />
          ))}
        </LineChart>
      </ResponsiveContainer>
    </Card>
  );
};

export default GpuChart;
