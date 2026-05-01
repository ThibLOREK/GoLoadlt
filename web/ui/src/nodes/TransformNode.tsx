import { memo, useState } from "react";
import { Handle, Position, type NodeProps } from "@xyflow/react";

export type TransformStep = {
  type: "mapper" | "filter" | "caster";
  config: Record<string, unknown>;
};

export type TransformNodeData = {
  label: string;
  steps: TransformStep[];
  onChange?: (data: Partial<TransformNodeData>) => void;
};

const inputCls = "w-full px-2 py-1 mb-1.5 rounded bg-[#0d2118] border border-green-900 text-white text-xs placeholder-green-300/40 focus:outline-none focus:border-green-400";
const labelCls = "block text-[10px] text-green-300 mb-0.5 uppercase tracking-wide";

export default memo(function TransformNode({ data }: NodeProps<TransformNodeData>) {
  const [open, setOpen] = useState(false);
  const { steps = [], onChange } = data as TransformNodeData;

  const addStep = () => onChange?.({ steps: [...steps, { type: "mapper", config: { mapping: {} } }] });
  const removeStep = (i: number) => onChange?.({ steps: steps.filter((_, idx) => idx !== i) });
  const updateStep = (i: number, patch: Partial<TransformStep>) => {
    onChange?.({ steps: steps.map((s, idx) => idx === i ? { ...s, ...patch } : s) });
  };

  return (
    <div className="rounded-lg p-3 min-w-[240px] border-2" style={{ background: "#1a3a2a", borderColor: "#4caf50" }}>
      <Handle type="target" position={Position.Left} style={{ background: "#4caf50" }} />

      <div className="flex justify-between items-center mb-2">
        <span className="text-xs font-bold text-green-300 uppercase tracking-wider">
          ⬛ TRANSFORM ({steps.length} step{steps.length !== 1 ? "s" : ""})
        </span>
        <button onClick={() => setOpen(o => !o)} className="text-green-300 hover:text-white text-xs px-1 transition-colors" title="Configurer">
          {open ? "▲" : "⚙"}
        </button>
      </div>

      {open && (
        <div className="border-t border-green-900 pt-2">
          {steps.map((step, i) => (
            <div key={i} className="mb-2 p-2 rounded" style={{ background: "#0d2118" }}>
              <div className="flex justify-between items-center mb-1.5">
                <select
                  value={step.type}
                  onChange={e => updateStep(i, { type: e.target.value as TransformStep["type"], config: {} })}
                  className="px-2 py-0.5 rounded bg-[#1a3a2a] border border-green-900 text-green-200 text-xs focus:outline-none"
                >
                  <option value="mapper">Mapper</option>
                  <option value="filter">Filter</option>
                  <option value="caster">Caster</option>
                </select>
                <button onClick={() => removeStep(i)} className="text-red-400 hover:text-red-300 text-xs px-1 transition-colors">✕</button>
              </div>

              {step.type === "mapper" && (
                <>
                  <label className={labelCls}>Mapping JSON</label>
                  <textarea
                    rows={2}
                    className={inputCls + " resize-none"}
                    placeholder='{"old_col":"new_col"}'
                    value={typeof step.config.mapping === "object" ? JSON.stringify(step.config.mapping) : "{}"}
                    onChange={e => { try { updateStep(i, { config: { mapping: JSON.parse(e.target.value) } }); } catch {} }}
                  />
                </>
              )}
              {step.type === "filter" && (
                <>
                  <label className={labelCls}>Condition</label>
                  <input className={inputCls} placeholder="ex: amount > 100" value={(step.config.condition as string) ?? ""}
                    onChange={e => updateStep(i, { config: { condition: e.target.value } })} />
                </>
              )}
              {step.type === "caster" && (
                <>
                  <label className={labelCls}>Colonne:type (ex: age:int)</label>
                  <input className={inputCls}
                    value={
                      Array.isArray(step.config.rules) && (step.config.rules as Array<Record<string, string>>).length > 0
                        ? `${(step.config.rules as Array<Record<string, string>>)[0].column}:${(step.config.rules as Array<Record<string, string>>)[0].cast_to}`
                        : ""
                    }
                    onChange={e => {
                      const [col, castTo] = e.target.value.split(":");
                      updateStep(i, { config: { rules: [{ column: col, cast_to: castTo || "string" }] } });
                    }}
                  />
                </>
              )}
            </div>
          ))}
          <button onClick={addStep} className="w-full py-1 rounded border border-green-700 text-green-300 hover:bg-green-900/30 text-xs transition-colors">
            + Ajouter une transformation
          </button>
        </div>
      )}

      <Handle type="source" position={Position.Right} style={{ background: "#4caf50" }} />
    </div>
  );
});
