import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { api } from '../api/client'
import { useAuth } from '../context/useAuth'

export default function AuthPage({ mode }) {
  const signup = mode === 'signup'
  const navigate = useNavigate()
  const { loginSession } = useAuth()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  const submit = async (event) => {
    event.preventDefault()
    setBusy(true)
    setError('')
    try {
      const data = signup ? await api.signup(email, password) : await api.login(email, password)
      loginSession(data.token, data.user)
      navigate('/dashboard')
    } catch (requestError) {
      setError(requestError.message)
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="mx-auto max-w-md px-4 py-14">
      <div className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm dark:border-slate-800 dark:bg-slate-900 sm:p-8">
        <h1 className="text-2xl font-black">{signup ? 'Create your account' : 'Welcome back'}</h1>
        <p className="mt-2 text-sm text-slate-600 dark:text-slate-300">{signup ? 'Start creating polls in seconds.' : 'Sign in to manage your polls.'}</p>
        <form className="mt-6 space-y-4" onSubmit={submit}>
          <label className="block text-sm font-semibold">Email
            <input type="email" required autoComplete="email" value={email} onChange={(event) => setEmail(event.target.value)} className="mt-2 w-full rounded-xl border border-slate-300 bg-transparent px-3 py-3 dark:border-slate-700" />
          </label>
          <label className="block text-sm font-semibold">Password
            <input type="password" required minLength={8} autoComplete={signup ? 'new-password' : 'current-password'} value={password} onChange={(event) => setPassword(event.target.value)} className="mt-2 w-full rounded-xl border border-slate-300 bg-transparent px-3 py-3 dark:border-slate-700" />
          </label>
          {error && <p role="alert" className="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700 dark:bg-red-950 dark:text-red-300">{error}</p>}
          <button disabled={busy} className="w-full rounded-xl bg-indigo-600 px-4 py-3 font-bold text-white hover:bg-indigo-700 disabled:opacity-60">{busy ? 'Please wait...' : signup ? 'Create account' : 'Sign in'}</button>
        </form>
        <p className="mt-5 text-center text-sm text-slate-600 dark:text-slate-300">
          {signup ? 'Already have an account?' : 'New to LivePulse?'}{' '}
          <Link className="font-bold text-indigo-600" to={signup ? '/login' : '/signup'}>{signup ? 'Sign in' : 'Create account'}</Link>
        </p>
      </div>
    </div>
  )
}
