import { useCallback, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { ReactFlow, addEdge, Background, Controls, MiniMap, useEdgesState, useNodesState, type Connection, type Node } from "@xyflow/react";
import "@xyflow/react/dist/style.css";
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
  const [saving, setSaving] = useState(false);
  const [success, setSuccess] = useState(false);
  const [error, setError] = useState("");

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

  const handleSave = async () => {
    setSaving(true);
    setError("");
    try {
      const payload = {
        name,
        description,
        source_type: sourceData.sourceType,
        target_type: targetData.targetType,
        source_config: sourceData.config,
        target_config: targetData.config,
        steps: transformData.steps.map(s => ({ type: s.type, config: s.config })),
      };
      const pipeline = isNew ? await pipelinesApi.create(payload) : await pipelinesApi.update(id!, payload);
      setSuccess(true);
      setTimeout(() => setSuccess(false), 3000);
      if (isNew) navigate(`/pipelines/${pipeline.id}/design`);
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Erreur lors de la sauvegarde");
    } finally {
      setSaving(false);
    }
  };

  const handleRun = async () => {
    if (!id || isNew) return;
    try {
      await pipelinesApi.run(id);
      navigate(`/pipelines/${id}/runs`);
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Erreur lors du lancement");
    }
  };

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="flex items-center gap-3 mb-3 flex-wrap">
        <h2 className="text-lg font-bold text-white">
          {isNew ? "Nouveau pipeline" : `Designer — ${name || id}`}
        </h2>
      </div>

      {/* Toolbar */}
      <div className="flex items-center gap-2 mb-3 flex-wrap">
        <input
          type="text"
          placeholder="Nom du pipeline"
          value={name}
          onChange={e => setName(e.target.value)}
          className="px-3 py-1.5 rounded bg-gray-800 border border-gray-600 text-white text-sm focus:outline-none focus:border-blue-400 min-w-[180px]"
        />
        <input
          type="text"
          placeholder="Description"
          value={description}
          onChange={e => setDescription(e.target.value)}
          className="px-3 py-1.5 rounded bg-gray-800 border border-gray-600 text-white text-sm focus:outline-none focus:border-blue-400 min-w-[240px]"
        />
        <button
          onClick={handleSave}
          disabled={!name || saving}
          className="px-4 py-1.5 rounded bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white text-sm font-medium transition-colors"
        >
          {saving ? "Sauvegarde…" : "Sauvegarder"}
        </button>
        {!isNew && (
          <button
            onClick={handleRun}
            className="px-4 py-1.5 rounded border border-green-500 text-green-400 hover:bg-green-900/30 text-sm font-medium transition-colors"
          >
            ▶ Lancer
          </button>
        )}
      </div>

      {/* Feedback messages */}
      {success && (
        <div className="mb-2 px-3 py-2 rounded bg-green-900/40 border border-green-600 text-green-300 text-sm">
          ✓ Pipeline sauvegardé !
        </div>
      )}
      {error && (
        <div className="mb-2 px-3 py-2 rounded bg-red-900/40 border border-red-600 text-red-300 text-sm">
          ✗ {error}
        </div>
      )}

      {/* ReactFlow canvas */}
      <div className="flex-1 rounded overflow-hidden border border-gray-700" style={{ height: 520 }}>
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
      </div>
    </div>
  );
}
