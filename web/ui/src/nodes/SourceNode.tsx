import { memo, useState } from "react";
import { Handle, Position, NodeProps } from "reactflow";
import {
  Box, Typography, Select, MenuItem, TextField, Divider, IconButton, Collapse,
} from "@mui/material";
import SettingsIcon from "@mui/icons-material/Settings";

export type SourceNodeData = {
  sourceType: "csv" | "postgres" | "api";
  label: string;
  config: Record<string, unknown>;
  onChange?: (data: Partial<SourceNodeData>) => void;
};

export default memo(function SourceNode({ data }: NodeProps<SourceNodeData>) {
  const [open, setOpen] = useState(false);
  const { sourceType = "csv", config = {}, onChange } = data;

  const update = (key: string, value: unknown) =>
    onChange?.({ config: { ...config, [key]: value } });

  return (
    <Box sx={{ background: "#1e3a5f", border: "2px solid #2196f3", borderRadius: 2, minWidth: 220, p: 1.5 }}>
      <Box display="flex" justifyContent="space-between" alignItems="center">
        <Typography variant="caption" color="#90caf9" fontWeight="bold">SOURCE</Typography>
        <IconButton size="small" onClick={() => setOpen(o => !o)} sx={{ color: "#90caf9" }}>
          <SettingsIcon fontSize="small" />
        </IconButton>
      </Box>

      <Select
        size="small" fullWidth
        value={sourceType}
        onChange={e => onChange?.({ sourceType: e.target.value as SourceNodeData["sourceType"], config: {} })}
        sx={{ mt: 0.5, mb: 1, bgcolor: "#0d2137", color: "white" }}
      >
        <MenuItem value="csv">CSV</MenuItem>
        <MenuItem value="postgres">PostgreSQL</MenuItem>
        <MenuItem value="api">API REST</MenuItem>
      </Select>

      <Collapse in={open}>
        <Divider sx={{ mb: 1, borderColor: "#2196f3" }} />
        {sourceType === "csv" && (
          <>
            <TextField size="small" fullWidth label="Chemin fichier" value={config.file_path ?? ""}
              onChange={e => update("file_path", e.target.value)} sx={inputSx} />
            <TextField size="small" fullWidth label="Délimiteur" value={config.delimiter ?? ","}
              onChange={e => update("delimiter", e.target.value)} sx={inputSx} />
          </>
        )}
        {sourceType === "postgres" && (
          <>
            <TextField size="small" fullWidth label="DSN (optionnel)" value={config.dsn ?? ""}
              onChange={e => update("dsn", e.target.value)} sx={inputSx} />
            <TextField size="small" fullWidth label="Schéma" value={config.schema ?? "public"}
              onChange={e => update("schema", e.target.value)} sx={inputSx} />
            <TextField size="small" fullWidth label="Table" value={config.table_name ?? ""}
              onChange={e => update("table_name", e.target.value)} sx={inputSx} />
            <TextField size="small" fullWidth label="WHERE (optionnel)" value={config.where ?? ""}
              onChange={e => update("where", e.target.value)} sx={inputSx} />
          </>
        )}
        {sourceType === "api" && (
          <>
            <TextField size="small" fullWidth label="URL" value={config.url ?? ""}
              onChange={e => update("url", e.target.value)} sx={inputSx} />
            <TextField size="small" fullWidth label="Chemin données (ex: data)" value={config.data_path ?? ""}
              onChange={e => update("data_path", e.target.value)} sx={inputSx} />
            <TextField size="small" fullWidth label="Taille page" value={config.page_size ?? 100}
              onChange={e => update("page_size", Number(e.target.value))} sx={inputSx} />
          </>
        )}
      </Collapse>

      <Handle type="source" position={Position.Right} style={{ background: "#2196f3" }} />
    </Box>
  );
});

const inputSx = { mb: 1, "& input": { color: "white" }, "& label": { color: "#90caf9" } };
