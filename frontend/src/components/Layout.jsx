import { Link, NavLink, useNavigate } from 'react-router-dom'
import { BarChart3, Moon, Sun } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useAuth } from '../context/useAuth'

export function Layout({ children }) {
  const { token, logout } = useAuth()
  const navigate = useNavigate()
  const [dark, setDark] = useState(() => localStorage.getItem('poll_theme') === 'dark')

  useEffect(() => {
    document.documentElement.classList.toggle('dark', dark)
    localStorage.setItem('poll_theme', dark ? 'dark' : 'light')
  }, [dark])

  const signOut = () => {
    logout()
    navigate('/login')
  }

  return (
    <div className="min-h-screen bg-slate-50 text-slate-900 dark:bg-slate-950 dark:text-slate-100">
      <header className="border-b border-slate-200/70 bg-white/90 backdrop-blur dark:border-slate-800 dark:bg-slate-950/90">
        <nav className="mx-auto flex max-w-6xl items-center justify-between px-4 py-3 sm:px-6">
          <Link to="/" className="flex items-center gap-2 font-bold text-indigo-600">
            <span className="grid h-9 w-9 place-items-center rounded-xl bg-indigo-600 text-white"><BarChart3 size={18} /></span>
            LivePulse
          </Link>
          <div className="flex items-center gap-2">
            {token && <NavLink to="/dashboard" className="rounded-lg px-3 py-2 text-sm font-semibold hover:bg-slate-100 dark:hover:bg-slate-800">Dashboard</NavLink>}
            <button type="button" className="grid h-9 w-9 place-items-center rounded-lg border border-slate-200 dark:border-slate-700" onClick={() => setDark((value) => !value)} aria-label="Toggle color theme">
              {dark ? <Sun size={17} /> : <Moon size={17} />}
            </button>
            {token ? (
              <button type="button" onClick={signOut} className="rounded-lg px-3 py-2 text-sm font-semibold text-slate-600 hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-800">Sign out</button>
            ) : (
              <Link to="/login" className="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-semibold text-white hover:bg-indigo-700">Sign in</Link>
            )}
          </div>
        </nav>
      </header>
      <main>{children}</main>
    </div>
  )
}
