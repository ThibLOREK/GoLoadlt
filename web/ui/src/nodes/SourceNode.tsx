import { memo, useState } from "react";
import { Handle, Position, type NodeProps } from "@xyflow/react";

export type SourceNodeData = {
  sourceType: "csv" | "postgres" | "mysql" | "mssql" | "api";
  label: string;
  config: Record<string, unknown>;
  onChange?: (data: Partial<SourceNodeData>) => void;
};

const inputCls = "w-full px-2 py-1 mb-2 rounded bg-[#0d2137] border border-blue-900 text-white text-xs placeholder-blue-300/40 focus:outline-none focus:border-blue-400";
const labelCls = "block text-[10px] text-blue-300 mb-0.5 uppercase tracking-wide";

export default memo(function SourceNode({ data }: NodeProps<SourceNodeData>) {
  const [open, setOpen] = useState(false);
  const { sourceType = "csv", config = {}, onChange } = data as SourceNodeData;

  const update = (key: string, value: unknown) =>
    onChange?.({ config: { ...config, [key]: value } });

  return (
    <div className="rounded-lg p-3 min-w-[220px] border-2" style={{ background: "#1e3a5f", borderColor: "#2196f3" }}>
      <div className="flex justify-between items-center mb-2">
        <span className="text-xs font-bold text-blue-300 uppercase tracking-wider">⬛ SOURCE</span>
        <button onClick={() => setOpen(o => !o)} className="text-blue-300 hover:text-white text-xs px-1 transition-colors" title="Configurer">
          {open ? "▲" : "⚙"}
        </button>
      </div>

      <select
        value={sourceType}
        onChange={e => onChange?.({ sourceType: e.target.value as SourceNodeData["sourceType"], config: {} })}
        className="w-full px-2 py-1 mb-2 rounded bg-[#0d2137] border border-blue-900 text-white text-xs focus:outline-none focus:border-blue-400"
      >
        <option value="csv">CSV</option>
        <option value="postgres">PostgreSQL</option>
        <option value="mysql">MySQL</option>
        <option value="mssql">SQL Server</option>
        <option value="api">API REST</option>
      </select>

      {open && (
        <div className="border-t border-blue-900 pt-2 mt-1">
          {sourceType === "csv" && (
            <>
              <label className={labelCls}>Chemin fichier</label>
              <input className={inputCls} value={(config.file_path as string) ?? ""} onChange={e => update("file_path", e.target.value)} />
              <label className={labelCls}>Délimiteur</label>
              <input className={inputCls} value={(config.delimiter as string) ?? ","} onChange={e => update("delimiter", e.target.value)} />
            </>
          )}
          {(sourceType === "postgres" || sourceType === "mysql" || sourceType === "mssql") && (
            <>
              <label className={labelCls}>DSN (optionnel)</label>
              <input className={inputCls} value={(config.dsn as string) ?? ""} onChange={e => update("dsn", e.target.value)} />
              <label className={labelCls}>Schéma</label>
              <input className={inputCls} value={(config.schema as string) ?? "public"} onChange={e => update("schema", e.target.value)} />
              <label className={labelCls}>Table</label>
              <input className={inputCls} value={(config.table_name as string) ?? ""} onChange={e => update("table_name", e.target.value)} />
              <label className={labelCls}>WHERE (optionnel)</label>
              <input className={inputCls} value={(config.where as string) ?? ""} onChange={e => update("where", e.target.value)} />
            </>
          )}
          {sourceType === "api" && (
            <>
              <label className={labelCls}>URL</label>
              <input className={inputCls} value={(config.url as string) ?? ""} onChange={e => update("url", e.target.value)} />
              <label className={labelCls}>Chemin données (ex: data)</label>
              <input className={inputCls} value={(config.data_path as string) ?? ""} onChange={e => update("data_path", e.target.value)} />
              <label className={labelCls}>Taille page</label>
              <input type="number" className={inputCls} value={(config.page_size as number) ?? 100} onChange={e => update("page_size", Number(e.target.value))} />
            </>
          )}
        </div>
      )}

      <Handle type="source" position={Position.Right} style={{ background: "#2196f3" }} />
    </div>
  );
});
