import { memo, useState } from "react";
import { Handle, Position, type NodeProps } from "@xyflow/react";

export type TargetNodeData = {
  targetType: "postgres" | "csv" | "mysql" | "mssql";
  label: string;
  config: Record<string, unknown>;
  onChange?: (data: Partial<TargetNodeData>) => void;
};

const inputCls = "w-full px-2 py-1 mb-2 rounded bg-[#1a0a1a] border border-purple-900 text-white text-xs placeholder-purple-300/40 focus:outline-none focus:border-purple-400";
const labelCls = "block text-[10px] text-purple-300 mb-0.5 uppercase tracking-wide";

export default memo(function TargetNode({ data }: NodeProps<TargetNodeData>) {
  const [open, setOpen] = useState(false);
  const { targetType = "postgres", config = {}, onChange } = data as TargetNodeData;
  const update = (key: string, value: unknown) => onChange?.({ config: { ...config, [key]: value } });

  return (
    <div className="rounded-lg p-3 min-w-[220px] border-2" style={{ background: "#3a1e3a", borderColor: "#9c27b0" }}>
      <Handle type="target" position={Position.Left} style={{ background: "#9c27b0" }} />

      <div className="flex justify-between items-center mb-2">
        <span className="text-xs font-bold text-purple-300 uppercase tracking-wider">⬛ TARGET</span>
        <button onClick={() => setOpen(o => !o)} className="text-purple-300 hover:text-white text-xs px-1 transition-colors" title="Configurer">
          {open ? "▲" : "⚙"}
        </button>
      </div>

      <select
        value={targetType}
        onChange={e => onChange?.({ targetType: e.target.value as TargetNodeData["targetType"], config: {} })}
        className="w-full px-2 py-1 mb-2 rounded bg-[#1a0a1a] border border-purple-900 text-white text-xs focus:outline-none focus:border-purple-400"
      >
        <option value="postgres">PostgreSQL</option>
        <option value="mysql">MySQL</option>
        <option value="mssql">SQL Server</option>
        <option value="csv">CSV</option>
      </select>

      {open && (
        <div className="border-t border-purple-900 pt-2 mt-1">
          {(targetType === "postgres" || targetType === "mysql" || targetType === "mssql") && (
            <>
              <label className={labelCls}>Schéma</label>
              <input className={inputCls} value={(config.schema as string) ?? "public"} onChange={e => update("schema", e.target.value)} />
              <label className={labelCls}>Table</label>
              <input className={inputCls} value={(config.table_name as string) ?? ""} onChange={e => update("table_name", e.target.value)} />
              <label className={labelCls}>Taille batch</label>
              <input type="number" className={inputCls} value={(config.batch_size as number) ?? 500} onChange={e => update("batch_size", Number(e.target.value))} />
            </>
          )}
          {targetType === "csv" && (
            <>
              <label className={labelCls}>Chemin fichier</label>
              <input className={inputCls} value={(config.file_path as string) ?? ""} onChange={e => update("file_path", e.target.value)} />
              <label className={labelCls}>Délimiteur</label>
              <input className={inputCls} value={(config.delimiter as string) ?? ","} onChange={e => update("delimiter", e.target.value)} />
            </>
          )}
        </div>
      )}
    </div>
  );
});
