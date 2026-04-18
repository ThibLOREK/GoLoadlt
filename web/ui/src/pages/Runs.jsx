import React, { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { api } from '../api'

const s = {
  page:  { padding:'2rem', maxWidth:900, margin:'0 auto' },
  back:  { background:'none', border:'none', color:'#38bdf8', cursor:'pointer',
           fontSize:'.9rem', marginBottom:'1.2rem', padding:0 },
  title: { fontSize:'1.3rem', fontWeight:700, marginBottom:'1.5rem' },
  card:  { background:'#1e293b', border:'1px solid #334155', borderRadius:10,
           padding:'1rem 1.5rem', marginBottom:'.8rem',
           display:'flex', justifyContent:'space-between', alignItems:'center' },
  id:    { fontSize:'.8rem', color:'#64748b', fontFamily:'monospace' },
  meta:  { fontSize:'.8rem', color:'#64748b', marginTop:'.2rem' },
  badge: { padding:'.25rem .7rem', borderRadius:20, fontSize:'.75rem', fontWeight:600 },
  empty: { textAlign:'center', color:'#475569', marginTop:'4rem', fontSize:'1.1rem' },
  refresh:{ background:'#334155', border:'none', color:'#e2e8f0', padding:'.4rem .9rem',
            borderRadius:6, cursor:'pointer', fontSize:'.85rem', marginLeft:'1rem' },
}

const STATUS_COLORS = {
  pending:   { background:'#1e3a5f', color:'#93c5fd' },
  running:   { background:'#1c3d2e', color:'#4ade80' },
  succeeded: { background:'#166534', color:'#86efac' },
  failed:    { background:'#7c2d12', color:'#fca5a5' },
  cancelled: { background:'#44403c', color:'#d6d3d1' },
}

function fmt(iso) {
  if (!iso) return '—'
  return new Date(iso).toLocaleString('fr-FR')
}

export default function Runs() {
  const { id } = useParams()
  const nav = useNavigate()
  const [runs, setRuns] = useState([])

  async function load() {
    try { setRuns(await api.listRuns(id)) } catch {}
  }

  useEffect(() => { load() }, [id])

  return (
    <div style={s.page}>
      <button style={s.back} onClick={() => nav('/pipelines')}>← Retour</button>
      <div style={{ display:'flex', alignItems:'center', marginBottom:'1.5rem' }}>
        <span style={s.title}>Historique des runs</span>
        <button style={s.refresh} onClick={load}>↻ Actualiser</button>
      </div>

      {runs.length === 0
        ? <div style={s.empty}>Aucun run pour ce pipeline.</div>
        : runs.map(r => (
          <div key={r.id} style={s.card}>
            <div>
              <div style={s.id}>{r.id}</div>
              <div style={s.meta}>
                Démarré : {fmt(r.started_at)} · Terminé : {fmt(r.finished_at)}
              </div>
              {r.error_msg && (
                <div style={{ fontSize:'.8rem', color:'#f87171', marginTop:'.2rem' }}>
                  ⚠ {r.error_msg}
                </div>
              )}
            </div>
            <div style={{ textAlign:'right' }}>
              <span style={{ ...s.badge, ...STATUS_COLORS[r.status] }}>{r.status}</span>
              <div style={{ ...s.meta, marginTop:'.4rem' }}>
                {r.records_read} lus · {r.records_loaded} chargés
              </div>
            </div>
          </div>
        ))
      }
    </div>
  )
}