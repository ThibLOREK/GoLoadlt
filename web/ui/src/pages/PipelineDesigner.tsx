import { useCallback, useState } from "react";
import { useParams } from "react-router-dom";
import ReactFlow, {
  addEdge,
  Background,
  Controls,
  MiniMap,
  useEdgesState,
  useNodesState,
  Connection,
} from "reactflow";
import "reactflow/dist/style.css";
import { Box, Typography, Paper, Stack, Button, TextField, Select, MenuItem } from "@mui/material";

const INITIAL_NODES = [
  { id: "source", type: "input", position: { x: 80, y: 200 }, data: { label: "Source" } },
  { id: "transform", position: { x: 350, y: 200 }, data: { label: "Transform" } },
  { id: "target", type: "output", position: { x: 620, y: 200 }, data: { label: "Target" } },
];
const INITIAL_EDGES = [
  { id: "e1", source: "source", target: "transform", animated: true },
  { id: "e2", source: "transform", target: "target", animated: true },
];

export default function PipelineDesigner() {
  const { id } = useParams();
  const [nodes, , onNodesChange] = useNodesState(INITIAL_NODES);
  const [edges, setEdges, onEdgesChange] = useEdgesState(INITIAL_EDGES);
  const [name, setName] = useState("");
  const [sourceType, setSourceType] = useState("csv");
  const [targetType, setTargetType] = useState("postgres");

  const onConnect = useCallback(
    (connection: Connection) => setEdges(eds => addEdge({ ...connection, animated: true }, eds)),
    [setEdges]
  );

  return (
    <Box>
      <Typography variant="h5" fontWeight="bold" mb={2}>
        Designer {id === "new" ? "(nouveau pipeline)" : `— ${id}`}
      </Typography>
      <Stack direction="row" spacing={2} mb={2} alignItems="center">
        <TextField label="Nom" size="small" value={name} onChange={e => setName(e.target.value)} />
        <Select size="small" value={sourceType} onChange={e => setSourceType(e.target.value)} displayEmpty>
          <MenuItem value="csv">CSV</MenuItem>
          <MenuItem value="postgres">PostgreSQL</MenuItem>
        </Select>
        <Typography>→</Typography>
        <Select size="small" value={targetType} onChange={e => setTargetType(e.target.value)}>
          <MenuItem value="postgres">PostgreSQL</MenuItem>
          <MenuItem value="csv">CSV</MenuItem>
        </Select>
        <Button variant="contained">Sauvegarder</Button>
      </Stack>
      <Paper sx={{ height: 500 }}>
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onConnect={onConnect}
          fitView
        >
          <Background />
          <Controls />
          <MiniMap />
        </ReactFlow>
      </Paper>
    </Box>
  );
}
