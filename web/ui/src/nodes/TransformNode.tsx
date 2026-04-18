import { memo, useState } from "react";
import { Handle, Position, NodeProps } from "reactflow";
import {
  Box, Typography, Select, MenuItem, TextField, Divider, IconButton,
  Collapse, Button, Stack,
} from "@mui/material";
import SettingsIcon from "@mui/icons-material/Settings";
import AddIcon from "@mui/icons-material/Add";
import DeleteIcon from "@mui/icons-material/Delete";

export type TransformStep = {
  type: "mapper" | "filter" | "caster";
  config: Record<string, unknown>;
};

export type TransformNodeData = {
  label: string;
  steps: TransformStep[];
  onChange?: (data: Partial<TransformNodeData>) => void;
};

export default memo(function TransformNode({ data }: NodeProps<TransformNodeData>) {
  const [open, setOpen] = useState(false);
  const { steps = [], onChange } = data;

  const addStep = () => onChange?.({ steps: [...steps, { type: "mapper", config: { mapping: {} } }] });

  const removeStep = (i: number) =>
    onChange?.({ steps: steps.filter((_, idx) => idx !== i) });

  const updateStep = (i: number, patch: Partial<TransformStep>) => {
    const updated = steps.map((s, idx) => idx === i ? { ...s, ...patch } : s);
    onChange?.({ steps: updated });
  };

  return (
    <Box sx={{ background: "#1a3a2a", border: "2px solid #4caf50", borderRadius: 2, minWidth: 240, p: 1.5 }}>
      <Handle type="target" position={Position.Left} style={{ background: "#4caf50" }} />

      <Box display="flex" justifyContent="space-between" alignItems="center">
        <Typography variant="caption" color="#a5d6a7" fontWeight="bold">
          TRANSFORM ({steps.length} step{steps.length !== 1 ? "s" : ""})
        </Typography>
        <IconButton size="small" onClick={() => setOpen(o => !o)} sx={{ color: "#a5d6a7" }}>
          <SettingsIcon fontSize="small" />
        </IconButton>
      </Box>

      <Collapse in={open}>
        <Divider sx={{ my: 1, borderColor: "#4caf50" }} />
        {steps.map((step, i) => (
          <Box key={i} sx={{ mb: 1.5, p: 1, bgcolor: "#0d2118", borderRadius: 1 }}>
            <Stack direction="row" justifyContent="space-between" alignItems="center" mb={0.5}>
              <Select size="small" value={step.type}
                onChange={e => updateStep(i, { type: e.target.value as TransformStep["type"], config: {} })}
                sx={{ color: "white", fontSize: 12 }}>
                <MenuItem value="mapper">Mapper</MenuItem>
                <MenuItem value="filter">Filter</MenuItem>
                <MenuItem value="caster">Caster</MenuItem>
              </Select>
              <IconButton size="small" onClick={() => removeStep(i)} sx={{ color: "#ef5350" }}>
                <DeleteIcon fontSize="small" />
              </IconButton>
            </Stack>

            {step.type === "mapper" && (
              <TextField size="small" fullWidth multiline rows={2}
                label="mapping JSON" placeholder='{"old":"new"}'
                value={typeof step.config.mapping === "object"
                  ? JSON.stringify(step.config.mapping) : "{}"}
                onChange={e => {
                  try { updateStep(i, { config: { mapping: JSON.parse(e.target.value) } }); } catch {}
                }}
                sx={inputSx}
              />
            )}
            {step.type === "filter" && (
              <TextField size="small" fullWidth
                label="colonne" value={(step.config.rules as Array<Record<string, string>>)?.[0]?.column ?? ""}
                onChange={e => updateStep(i, { config: { rules: [{ column: e.target.value, operator: "eq", value: "" }] } })}
                sx={inputSx}
              />
            )}
            {step.type === "caster" && (
              <TextField size="small" fullWidth
                label="colonne:type (ex: age:int)"
                value={
                  Array.isArray(step.config.rules) && (step.config.rules as Array<Record<string,string>>).length > 0
                    ? `${(step.config.rules as Array<Record<string,string>>)[0].column}:${(step.config.rules as Array<Record<string,string>>)[0].cast_to}`
                    : ""
                }
                onChange={e => {
                  const [col, castTo] = e.target.value.split(":");
                  updateStep(i, { config: { rules: [{ column: col, cast_to: castTo || "string" }] } });
                }}
                sx={inputSx}
              />
            )}
          </Box>
        ))}
        <Button size="small" startIcon={<AddIcon />} onClick={addStep} sx={{ color: "#a5d6a7" }}>
          Ajouter une transformation
        </Button>
      </Collapse>

      <Handle type="source" position={Position.Right} style={{ background: "#4caf50" }} />
    </Box>
  );
});

const inputSx = { "& input, & textarea": { color: "white", fontSize: 12 }, "& label": { color: "#a5d6a7", fontSize: 12 } };
