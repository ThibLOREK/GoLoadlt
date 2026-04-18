import React, { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../api'

const s = {
  page:   { padding:'2rem', maxWidth:900, margin:'0 auto' },
  header: { display:'flex', justifyContent:'space-between', alignItems:'center',
            marginBottom:'1.5rem' },
  title:  { fontSize:'1.3rem', fontWeight:700 },
  btn:    { background:'#38bdf8', border:'none', color:'#0f172a', padding:'.5rem 1.2rem',
            borderRadius:6, cursor:'pointer', fontWeight:600 },
  card:   { background:'#1e293b', border:'1px solid #334155', borderRadius:10,
            padding:'1.2rem 1.5rem', marginBottom:'1rem',
            display:'flex', justifyContent:'space-between', alignItems:'center' },
  name:   { fontWeight:600, fontSize:'1rem' },
  meta:   { fontSize:'.8rem', color:'#64748b', marginTop:'.2rem' },
  badge:  { padding:'.2rem .6rem', borderRadius:20, fontSize:'.75rem', fontWeight:600 },
  actions:{ display:'flex', gap:'.6rem' },
  run:    { background:'#22c55e', border:'none', color:'#fff', padding:'.4rem .9rem',
            borderRadius:6, cursor:'pointer', fontSize:'.85rem' },
  hist:   { background:'#334155', border:'none', color:'#e2e8f0', padding:'.4rem .9rem',
            borderRadius:6, cursor:'pointer', fontSize:'.85rem' },
  del:    { background:'#ef4444', border:'none', color:'#fff', padding:'.4rem .9rem',
            borderRadius:6, cursor:'pointer', fontSize:'.85rem' },
  modal:  { position:'fixed', inset:0, background:'rgba(0,0,0,.6)',
            display:'flex', justifyContent:'center', alignItems:'center', zIndex:100 },
  mcard:  { background:'#1e293b', padding:'2rem', borderRadius:12, width:440,
            border:'1px solid #334155' },
  mtitle: { fontWeight:700, fontSize:'1.1rem', marginBottom:'1.2rem' },
  label:  { display:'block', fontSize:'.85rem', color:'#94a3b8', marginBottom:'.3rem' },
  input:  { width:'100%', padding:'.6rem .8rem', borderRadius:6, border:'1px solid #475569',
            background:'#0f172a', color:'#e2e8f0', fontSize:'.9rem', marginBottom:'.9rem' },
  select: { width:'100%', padding:'.6rem .8rem', borderRadius:6, border:'1px solid #475569',
            background:'#0f172a', color:'#e2e8f0', fontSize:'.9rem', marginBottom:'.9rem' },
  row:    { display:'flex', gap:'.8rem', justifyContent:'flex-end', marginTop:'.5rem' },
  cancel: { background:'#334155', border:'none', color:'#e2e8f0', padding:'.5rem 1rem',
            borderRadius:6, cursor:'pointer' },
  save:   { background:'#38bdf8', border:'none', color:'#0f172a', padding:'.5rem 1.2rem',
            borderRadius:6, cursor:'pointer', fontWeight:600 },
  empty:  { textAlign:'center', color:'#475569', marginTop:'4rem', fontSize:'1.1rem' },
}

const STATUS_COLORS = {
  draft:   { background:'#334155', color:'#94a3b8' },
  active:  { background:'#166534', color:'#86efac' },
  inactive:{ background:'#7c2d12', color:'#fdba74' },
}

export default function Pipelines() {
  const [pipelines, setPipelines] = useState([])
  const [showModal, setShowModal] = useState(false)
  const [form, setForm]           = useState({ name:'', description:'', source_type:'csv', target_type:'postgres' })
  const [toast, setToast]         = useState('')
  const nav = useNavigate()

  async function load() {
    try { setPipelines(await api.listPipelines()) } catch {}
  }

  useEffect(() => { load() }, [])

  function notify(msg) {
    setToast(msg)
    setTimeout(() => setToast(''), 3000)
  }

  async function create(e) {
    e.preventDefault()
    try {
      await api.createPipeline(form)
      setShowModal(false)
      setForm({ name:'', description:'', source_type:'csv', target_type:'postgres' })
      load()
      notify('Pipeline créé ✓')
    } catch (err) { notify('Erreur : ' + err.message) }
  }

  async function runNow(id) {
    try { await api.runPipeline(id); notify('Run lancé ✓') }
    catch (err) { notify('Erreur : ' + err.message) }
  }

  async function del(id) {
    if (!confirm('Supprimer ce pipeline ?')) return
    try { await api.deletePipeline(id); load(); notify('Supprimé ✓') }
    catch (err) { notify('Erreur : ' + err.message) }
  }

  return (
    <div style={s.page}>
      {toast && (
        <div style={{ position:'fixed', top:20, right:20, background:'#1e293b',
          border:'1px solid #38bdf8', padding:'.8rem 1.2rem', borderRadius:8,
          color:'#38bdf8', zIndex:200 }}>{toast}</div>
      )}

      <div style={s.header}>
        <span style={s.title}>Pipelines</span>
        <button style={s.btn} onClick={() => setShowModal(true)}>+ Nouveau pipeline</button>
      </div>

      {pipelines.length === 0
        ? <div style={s.empty}>Aucun pipeline — crée-en un !</div>
        : pipelines.map(p => (
          <div key={p.id} style={s.card}>
            <div>
              <div style={s.name}>{p.name}</div>
              <div style={s.meta}>{p.description || 'Pas de description'} · {p.source_type} → {p.target_type}</div>
            </div>
            <div style={s.actions}>
              <span style={{ ...s.badge, ...STATUS_COLORS[p.status] }}>{p.status}</span>
              <button style={{ background:'#6366f1', border:'none', color:'#fff', padding:'.4rem .9rem',
                borderRadius:6, cursor:'pointer', fontSize:'.85rem' }}
                onClick={() => nav(`/pipelines/${p.id}/config`)}>
                ⚙ Config
              </button>
              <button style={s.run}  onClick={() => runNow(p.id)}>▶ Run</button>
              <button style={s.hist} onClick={() => nav(`/pipelines/${p.id}/runs`)}>Historique</button>
              <button style={s.del}  onClick={() => del(p.id)}>✕</button>
            </div>
          </div>
        ))
      }

      {showModal && (
        <div style={s.modal} onClick={e => e.target === e.currentTarget && setShowModal(false)}>
          <div style={s.mcard}>
            <div style={s.mtitle}>Nouveau pipeline</div>
            <form onSubmit={create}>
              <label style={s.label}>Nom</label>
              <input style={s.input} value={form.name} required
                onChange={e => setForm({...form, name: e.target.value})} />
              <label style={s.label}>Description</label>
              <input style={s.input} value={form.description}
                onChange={e => setForm({...form, description: e.target.value})} />
              <label style={s.label}>Source</label>
              <select style={s.select} value={form.source_type}
                onChange={e => setForm({...form, source_type: e.target.value})}>
                <option value="csv">CSV</option>
                <option value="postgres">PostgreSQL</option>
                <option value="api">API REST</option>
              </select>
              <label style={s.label}>Cible</label>
              <select style={s.select} value={form.target_type}
                onChange={e => setForm({...form, target_type: e.target.value})}>
                <option value="postgres">PostgreSQL</option>
                <option value="csv">CSV</option>
              </select>
              <div style={s.row}>
                <button type="button" style={s.cancel} onClick={() => setShowModal(false)}>Annuler</button>
                <button type="submit" style={s.save}>Créer</button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}