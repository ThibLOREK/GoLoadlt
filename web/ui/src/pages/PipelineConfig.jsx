import React, { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { api } from '../api'

const s = {
  page:    { padding:'2rem', maxWidth:700, margin:'0 auto' },
  back:    { background:'none', border:'none', color:'#38bdf8', cursor:'pointer',
             fontSize:'.9rem', marginBottom:'1.2rem', padding:0 },
  title:   { fontSize:'1.3rem', fontWeight:700, marginBottom:'.4rem' },
  sub:     { color:'#64748b', fontSize:'.85rem', marginBottom:'1.8rem' },
  section: { background:'#1e293b', border:'1px solid #334155', borderRadius:10,
             padding:'1.5rem', marginBottom:'1.2rem' },
  stitle:  { fontWeight:700, fontSize:'.95rem', marginBottom:'1rem', color:'#38bdf8' },
  label:   { display:'block', fontSize:'.82rem', color:'#94a3b8', marginBottom:'.3rem' },
  input:   { width:'100%', padding:'.6rem .8rem', borderRadius:6, border:'1px solid #475569',
             background:'#0f172a', color:'#e2e8f0', fontSize:'.9rem', marginBottom:'1rem' },
  select:  { width:'100%', padding:'.6rem .8rem', borderRadius:6, border:'1px solid #475569',
             background:'#0f172a', color:'#e2e8f0', fontSize:'.9rem', marginBottom:'1rem' },
  row:     { display:'flex', gap:'1rem' },
  half:    { flex:1 },
  btn:     { background:'#38bdf8', border:'none', color:'#0f172a', padding:'.65rem 1.8rem',
             borderRadius:6, cursor:'pointer', fontWeight:700, fontSize:'1rem' },
  cancel:  { background:'#334155', border:'none', color:'#e2e8f0', padding:'.65rem 1.2rem',
             borderRadius:6, cursor:'pointer', fontSize:'1rem', marginRight:'.8rem' },
  toast:   { position:'fixed', top:20, right:20, background:'#1e293b',
             border:'1px solid #38bdf8', padding:'.8rem 1.2rem', borderRadius:8,
             color:'#38bdf8', zIndex:200 },
  hint:    { fontSize:'.78rem', color:'#475569', marginTop:'-.7rem', marginBottom:'1rem' },
}

function CsvSourceForm({ cfg, onChange }) {
  return <>
    <label style={s.label}>Chemin du fichier CSV</label>
    <input style={s.input} placeholder="C:/data/input.csv"
      value={cfg.path || ''} onChange={e => onChange({ ...cfg, path: e.target.value })} />
    <div style={s.row}>
      <div style={s.half}>
        <label style={s.label}>Délimiteur</label>
        <input style={s.input} placeholder="," maxLength={1}
          value={cfg.delimiter || ','} onChange={e => onChange({ ...cfg, delimiter: e.target.value })} />
      </div>
      <div style={s.half}>
        <label style={s.label}>Encodage</label>
        <select style={s.select} value={cfg.encoding || 'utf-8'}
          onChange={e => onChange({ ...cfg, encoding: e.target.value })}>
          <option value="utf-8">UTF-8</option>
          <option value="latin1">Latin-1 (ISO-8859-1)</option>
        </select>
      </div>
    </div>
    <label style={s.label}>
      <input type="checkbox" checked={cfg.has_header !== false}
        onChange={e => onChange({ ...cfg, has_header: e.target.checked })} />
      {' '}Première ligne = en-tête
    </label>
  </>
}

function PostgresSourceForm({ cfg, onChange }) {
  return <>
    <label style={s.label}>DSN PostgreSQL source</label>
    <input style={s.input} placeholder="postgres://user:pass@host:5432/db?sslmode=disable"
      value={cfg.dsn || ''} onChange={e => onChange({ ...cfg, dsn: e.target.value })} />
    <label style={s.label}>Requête SQL</label>
    <input style={s.input} placeholder="SELECT * FROM ma_table WHERE ..."
      value={cfg.query || ''} onChange={e => onChange({ ...cfg, query: e.target.value })} />
  </>
}

function ApiSourceForm({ cfg, onChange }) {
  return <>
    <label style={s.label}>URL de l'API</label>
    <input style={s.input} placeholder="https://api.example.com/data"
      value={cfg.url || ''} onChange={e => onChange({ ...cfg, url: e.target.value })} />
    <label style={s.label}>Méthode</label>
    <select style={s.select} value={cfg.method || 'GET'}
      onChange={e => onChange({ ...cfg, method: e.target.value })}>
      <option>GET</option>
      <option>POST</option>
    </select>
    <label style={s.label}>Header d'authentification (optionnel)</label>
    <input style={s.input} placeholder="Bearer mon-token"
      value={cfg.auth_header || ''} onChange={e => onChange({ ...cfg, auth_header: e.target.value })} />
    <label style={s.label}>Chemin JSON vers le tableau de données</label>
    <input style={s.input} placeholder="data.items  (laisser vide si la réponse est déjà un tableau)"
      value={cfg.json_path || ''} onChange={e => onChange({ ...cfg, json_path: e.target.value })} />
  </>
}

function PostgresTargetForm({ cfg, onChange }) {
  return <>
    <label style={s.label}>DSN PostgreSQL cible</label>
    <input style={s.input} placeholder="postgres://user:pass@host:5432/db?sslmode=disable"
      value={cfg.dsn || ''} onChange={e => onChange({ ...cfg, dsn: e.target.value })} />
    <label style={s.label}>Table cible</label>
    <input style={s.input} placeholder="public.ma_table"
      value={cfg.table || ''} onChange={e => onChange({ ...cfg, table: e.target.value })} />
    <label style={s.label}>Mode d'insertion</label>
    <select style={s.select} value={cfg.mode || 'append'}
      onChange={e => onChange({ ...cfg, mode: e.target.value })}>
      <option value="append">Append — ajouter les lignes</option>
      <option value="truncate">Truncate + Insert — remplacer tout</option>
      <option value="upsert">Upsert — mettre à jour si existe</option>
    </select>
    {cfg.mode === 'upsert' && <>
      <label style={s.label}>Colonne(s) clé pour l'upsert</label>
      <input style={s.input} placeholder="id  (séparées par des virgules si plusieurs)"
        value={cfg.upsert_keys || ''} onChange={e => onChange({ ...cfg, upsert_keys: e.target.value })} />
    </>}
  </>
}

function CsvTargetForm({ cfg, onChange }) {
  return <>
    <label style={s.label}>Chemin du fichier CSV de sortie</label>
    <input style={s.input} placeholder="C:/data/output.csv"
      value={cfg.path || ''} onChange={e => onChange({ ...cfg, path: e.target.value })} />
    <label style={s.label}>Délimiteur</label>
    <input style={s.input} placeholder="," maxLength={1}
      value={cfg.delimiter || ','} onChange={e => onChange({ ...cfg, delimiter: e.target.value })} />
  </>
}

const SOURCE_FORMS = {
  csv:      CsvSourceForm,
  postgres: PostgresSourceForm,
  api:      ApiSourceForm,
}

const TARGET_FORMS = {
  postgres: PostgresTargetForm,
  csv:      CsvTargetForm,
}

export default function PipelineConfig() {
  const { id } = useParams()
  const nav    = useNavigate()
  const [pipeline, setPipeline] = useState(null)
  const [srcCfg,   setSrcCfg]  = useState({})
  const [tgtCfg,   setTgtCfg]  = useState({})
  const [toast,    setToast]   = useState('')

  useEffect(() => {
    api.getPipeline(id).then(p => {
      setPipeline(p)
      setSrcCfg(p.source_config || {})
      setTgtCfg(p.target_config || {})
    }).catch(() => nav('/pipelines'))
  }, [id])

  function notify(msg) {
    setToast(msg)
    setTimeout(() => setToast(''), 3000)
  }

  async function save(e) {
    e.preventDefault()
    try {
      await api.updatePipeline(id, {
        name:          pipeline.name,
        description:   pipeline.description,
        status:        pipeline.status,
        source_type:   pipeline.source_type,
        target_type:   pipeline.target_type,
        source_config: srcCfg,
        target_config: tgtCfg,
      })
      notify('Configuration sauvegardée ✓')
    } catch (err) {
      notify('Erreur : ' + err.message)
    }
  }

  if (!pipeline) return <div style={{ padding:'2rem', color:'#64748b' }}>Chargement…</div>

  const SrcForm = SOURCE_FORMS[pipeline.source_type]
  const TgtForm = TARGET_FORMS[pipeline.target_type]

  return (
    <div style={s.page}>
      {toast && <div style={s.toast}>{toast}</div>}

      <button style={s.back} onClick={() => nav('/pipelines')}>← Retour</button>
      <div style={s.title}>{pipeline.name}</div>
      <div style={s.sub}>{pipeline.source_type} → {pipeline.target_type} · statut : {pipeline.status}</div>

      <form onSubmit={save}>
        <div style={s.section}>
          <div style={s.stitle}>⚙ Configuration de la source ({pipeline.source_type})</div>
          {SrcForm
            ? <SrcForm cfg={srcCfg} onChange={setSrcCfg} />
            : <div style={{ color:'#64748b' }}>Type source non supporté dans l'UI.</div>
          }
        </div>

        <div style={s.section}>
          <div style={s.stitle}>🎯 Configuration de la cible ({pipeline.target_type})</div>
          {TgtForm
            ? <TgtForm cfg={tgtCfg} onChange={setTgtCfg} />
            : <div style={{ color:'#64748b' }}>Type cible non supporté dans l'UI.</div>
          }
        </div>

        <div>
          <button type="button" style={s.cancel} onClick={() => nav('/pipelines')}>Annuler</button>
          <button type="submit" style={s.btn}>💾 Sauvegarder</button>
        </div>
      </form>
    </div>
  )
}