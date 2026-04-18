import { memo, useState } from "react";
import { Handle, Position, NodeProps } from "reactflow";
import { Box, Typography, Select, MenuItem, TextField, Divider, IconButton, Collapse } from "@mui/material";
import SettingsIcon from "@mui/icons-material/Settings";

export type TargetNodeData = {
  targetType: "postgres" | "csv";
  label: string;
  config: Record<string, unknown>;
  onChange?: (data: Partial<TargetNodeData>) => void;
};

export default memo(function TargetNode({ data }: NodeProps<TargetNodeData>) {
  const [open, setOpen] = useState(false);
  const { targetType = "postgres", config = {}, onChange } = data;
  const update = (key: string, value: unknown) => onChange?.({ config: { ...config, [key]: value } });

  return (
    <Box sx={{ background: "#3a1e3a", border: "2px solid #9c27b0", borderRadius: 2, minWidth: 220, p: 1.5 }}>
      <Handle type="target" position={Position.Left} style={{ background: "#9c27b0" }} />

      <Box display="flex" justifyContent="space-between" alignItems="center">
        <Typography variant="caption" color="#ce93d8" fontWeight="bold">TARGET</Typography>
        <IconButton size="small" onClick={() => setOpen(o => !o)} sx={{ color: "#ce93d8" }}>
          <SettingsIcon fontSize="small" />
        </IconButton>
      </Box>

      <Select size="small" fullWidth value={targetType}
        onChange={e => onChange?.({ targetType: e.target.value as TargetNodeData["targetType"], config: {} })}
        sx={{ mt: 0.5, mb: 1, bgcolor: "#1a0a1a", color: "white" }}>
        <MenuItem value="postgres">PostgreSQL</MenuItem>
        <MenuItem value="csv">CSV</MenuItem>
      </Select>

      <Collapse in={open}>
        <Divider sx={{ mb: 1, borderColor: "#9c27b0" }} />
        {targetType === "postgres" && (
          <>
            <TextField size="small" fullWidth label="Schéma" value={config.schema ?? "public"}
              onChange={e => update("schema", e.target.value)} sx={inputSx} />
            <TextField size="small" fullWidth label="Table" value={config.table_name ?? ""}
              onChange={e => update("table_name", e.target.value)} sx={inputSx} />
            <TextField size="small" fullWidth label="Taille batch" value={config.batch_size ?? 500}
              onChange={e => update("batch_size", Number(e.target.value))} sx={inputSx} />
          </>
        )}
        {targetType === "csv" && (
          <>
            <TextField size="small" fullWidth label="Chemin fichier" value={config.file_path ?? ""}
              onChange={e => update("file_path", e.target.value)} sx={inputSx} />
            <TextField size="small" fullWidth label="Délimiteur" value={config.delimiter ?? ","}
              onChange={e => update("delimiter", e.target.value)} sx={inputSx} />
          </>
        )}
      </Collapse>
    </Box>
  );
});

const inputSx = { mb: 1, "& input": { color: "white" }, "& label": { color: "#ce93d8" } };
