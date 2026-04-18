import { useCallback, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import ReactFlow, {
  addEdge, Background, Controls, MiniMap,
  useEdgesState, useNodesState, Connection, Node,
} from "reactflow";
import "reactflow/dist/style.css";
import { Box, Typography, Paper, Stack, Button, TextField, Alert, Snackbar } from "@mui/material";
import { useMutation } from "@tanstack/react-query";
import SourceNode, { SourceNodeData } from "../nodes/SourceNode";
import TransformNode, { TransformNodeData } from "../nodes/TransformNode";
import TargetNode, { TargetNodeData } from "../nodes/TargetNode";
import { pipelinesApi } from "../api/client";

const nodeTypes = { source: SourceNode, transform: TransformNode, target: TargetNode };

const buildInitialNodes = (
  onSourceChange: (d: Partial<SourceNodeData>) => void,
  onTransformChange: (d: Partial<TransformNodeData>) => void,
  onTargetChange: (d: Partial<TargetNodeData>) => void,
): Node[] => [
  {
    id: "source", type: "source", position: { x: 60, y: 180 },
    data: { label: "Source", sourceType: "csv", config: {}, onChange: onSourceChange },
  },
  {
    id: "transform", type: "transform", position: { x: 360, y: 160 },
    data: { label: "Transform", steps: [], onChange: onTransformChange },
  },
  {
    id: "target", type: "target", position: { x: 680, y: 180 },
    data: { label: "Target", targetType: "postgres", config: {}, onChange: onTargetChange },
  },
];

const INITIAL_EDGES = [
  { id: "e1", source: "source", target: "transform", animated: true },
  { id: "e2", source: "transform", target: "target", animated: true },
];

export default function PipelineDesigner() {
  const { id } = useParams();
  const navigate = useNavigate();
  const isNew = id === "new";

  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [success, setSuccess] = useState(false);

  const [sourceData, setSourceData] = useState<SourceNodeData>({ label: "Source", sourceType: "csv", config: {} });
  const [transformData, setTransformData] = useState<TransformNodeData>({ label: "Transform", steps: [] });
  const [targetData, setTargetData] = useState<TargetNodeData>({ label: "Target", targetType: "postgres", config: {} });

  const handleSourceChange = useCallback((d: Partial<SourceNodeData>) => setSourceData(prev => ({ ...prev, ...d })), []);
  const handleTransformChange = useCallback((d: Partial<TransformNodeData>) => setTransformData(prev => ({ ...prev, ...d })), []);
  const handleTargetChange = useCallback((d: Partial<TargetNodeData>) => setTargetData(prev => ({ ...prev, ...d })), []);

  const initialNodes = buildInitialNodes(handleSourceChange, handleTransformChange, handleTargetChange);
  const [nodes, , onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(INITIAL_EDGES);
  const onConnect = useCallback(
    (c: Connection) => setEdges(eds => addEdge({ ...c, animated: true }, eds)),
    [setEdges],
  );

  const saveMutation = useMutation({
    mutationFn: () => {
      const payload = {
        name,
        description,
        source_type: sourceData.sourceType,
        target_type: targetData.targetType,
        source_config: sourceData.config,
        target_config: targetData.config,
        steps: transformData.steps.map(s => ({ type: s.type, config: s.config })),
      };
      return isNew ? pipelinesApi.create(payload) : pipelinesApi.update(id!, payload);
    },
    onSuccess: (pipeline) => {
      setSuccess(true);
      if (isNew) navigate(`/pipelines/${pipeline.id}/design`);
    },
  });

  return (
    <Box>
      <Typography variant="h5" fontWeight="bold" mb={2}>
        {isNew ? "Nouveau pipeline" : `Designer — ${name || id}`}
      </Typography>

      <Stack direction="row" spacing={2} mb={2} alignItems="center" flexWrap="wrap">
        <TextField label="Nom du pipeline" size="small" value={name}
          onChange={e => setName(e.target.value)} sx={{ minWidth: 200 }} />
        <TextField label="Description" size="small" value={description}
          onChange={e => setDescription(e.target.value)} sx={{ minWidth: 280 }} />
        <Button variant="contained" onClick={() => saveMutation.mutate()}
          disabled={!name || saveMutation.isPending}>
          {saveMutation.isPending ? "Sauvegarde…" : "Sauvegarder"}
        </Button>
        {!isNew && (
          <Button variant="outlined" color="success" onClick={() => pipelinesApi.run(id!)
            .then(() => navigate(`/pipelines/${id}/runs`))}>
            ▶ Lancer
          </Button>
        )}
      </Stack>

      <Paper sx={{ height: 520, position: "relative" }}>
        <ReactFlow
          nodes={nodes.map(n => ({
            ...n,
            data: n.id === "source" ? { ...sourceData, onChange: handleSourceChange }
              : n.id === "transform" ? { ...transformData, onChange: handleTransformChange }
              : { ...targetData, onChange: handleTargetChange },
          }))}
          edges={edges}
          nodeTypes={nodeTypes}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onConnect={onConnect}
          fitView
        >
          <Background color="#333" />
          <Controls />
          <MiniMap nodeColor={(n) =>
            n.type === "source" ? "#2196f3" : n.type === "transform" ? "#4caf50" : "#9c27b0"
          } />
        </ReactFlow>
      </Paper>

      <Snackbar open={success} autoHideDuration={3000} onClose={() => setSuccess(false)}>
        <Alert severity="success">Pipeline sauvegardé !</Alert>
      </Snackbar>
    </Box>
  );
}
