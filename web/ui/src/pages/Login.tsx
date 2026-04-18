import React, { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../api'

const s = {
  wrap:  { display:'flex', justifyContent:'center', alignItems:'center', height:'100vh' },
  card:  { background:'#1e293b', padding:'2.5rem', borderRadius:12, width:360,
           border:'1px solid #334155' },
  title: { fontSize:'1.5rem', fontWeight:700, marginBottom:'1.5rem', color:'#38bdf8',
           textAlign:'center' },
  label: { display:'block', fontSize:'.85rem', color:'#94a3b8', marginBottom:'.3rem' },
  input: { width:'100%', padding:'.6rem .8rem', borderRadius:6, border:'1px solid #475569',
           background:'#0f172a', color:'#e2e8f0', fontSize:'.95rem', marginBottom:'1rem' },
  btn:   { width:'100%', padding:'.7rem', background:'#38bdf8', border:'none', borderRadius:6,
           color:'#0f172a', fontWeight:700, cursor:'pointer', fontSize:'1rem' },
  err:   { color:'#f87171', fontSize:'.85rem', marginTop:'.75rem', textAlign:'center' },
  sub:   { textAlign:'center', marginTop:'1rem', fontSize:'.85rem', color:'#64748b' },
  lnk:   { color:'#38bdf8', cursor:'pointer' },
}

export default function Login() {
  const [email, setEmail]       = useState('')
  const [password, setPassword] = useState('')
  const [error, setError]       = useState('')
  const [mode, setMode]         = useState('login')
  const nav = useNavigate()

  async function submit(e) {
    e.preventDefault()
    setError('')
    try {
      const fn = mode === 'login' ? api.login : api.register
      const data = await fn(email, password)
      if (data.token) {
        localStorage.setItem('token', data.token)
        nav('/pipelines')
      } else {
        setMode('login')
        setError('Compte créé ! Connecte-toi.')
      }
    } catch (err) {
      setError(err.message)
    }
  }

  return (
    <div style={s.wrap}>
      <div style={s.card}>
        <div style={s.title}>⚡ GoLoadIt</div>
        <form onSubmit={submit}>
          <label style={s.label}>Email</label>
          <input style={s.input} type="email" value={email}
            onChange={e => setEmail(e.target.value)} required />
          <label style={s.label}>Mot de passe</label>
          <input style={s.input} type="password" value={password}
            onChange={e => setPassword(e.target.value)} required />
          <button style={s.btn} type="submit">
            {mode === 'login' ? 'Se connecter' : 'Créer le compte'}
          </button>
        </form>
        {error && <div style={s.err}>{error}</div>}
        <div style={s.sub}>
          {mode === 'login'
            ? <>Pas de compte ? <span style={s.lnk} onClick={() => setMode('register')}>S'inscrire</span></>
            : <>Déjà un compte ? <span style={s.lnk} onClick={() => setMode('login')}>Se connecter</span></>
          }
        </div>
      </div>
    </div>
  )
}