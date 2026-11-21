import { ArrowRight, Radio, ShieldCheck, Zap } from 'lucide-react'
import { Link } from 'react-router-dom'
import { useAuth } from '../context/useAuth'

export default function Landing() {
  const { token } = useAuth()
  return (
    <div className="mx-auto max-w-6xl px-4 py-16 sm:px-6 sm:py-24">
      <div className="mx-auto max-w-3xl text-center">
        <div className="mx-auto mb-5 flex w-fit items-center gap-2 rounded-full border border-indigo-200 bg-indigo-50 px-3 py-1 text-sm font-semibold text-indigo-700 dark:border-indigo-800 dark:bg-indigo-950 dark:text-indigo-300">
          <Radio size={15} /> Live audience voting
        </div>
        <h1 className="text-4xl font-black tracking-tight sm:text-6xl">Ask it. Share it. <span className="text-indigo-600">Watch it move.</span></h1>
        <p className="mx-auto mt-5 max-w-2xl text-lg text-slate-600 dark:text-slate-300">Create a poll, send one link, and see every audience vote update across every screen in real time.</p>
        <div className="mt-8 flex flex-col justify-center gap-3 sm:flex-row">
          <Link to={token ? '/create' : '/signup'} className="inline-flex items-center justify-center gap-2 rounded-xl bg-indigo-600 px-5 py-3 font-bold text-white shadow-lg shadow-indigo-600/20 hover:bg-indigo-700">Create a poll <ArrowRight size={18} /></Link>
          {token && <Link to="/dashboard" className="rounded-xl border border-slate-300 px-5 py-3 font-bold hover:bg-white dark:border-slate-700 dark:hover:bg-slate-900">View dashboard</Link>}
        </div>
      </div>
      <div className="mt-16 grid gap-4 md:grid-cols-3">
        {[{ icon: Zap, title: 'Instant updates', text: 'SSE streams every accepted vote to every viewer.' }, { icon: ShieldCheck, title: 'Protected creation', text: 'Only signed-in owners can create and manage polls.' }, { icon: Radio, title: 'One simple link', text: 'Anyone with the link can vote without an account.' }].map(({ icon: Icon, title, text }) => (
          <div key={title} className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm dark:border-slate-800 dark:bg-slate-900">
            <Icon className="text-indigo-600" size={22} />
            <h2 className="mt-4 font-bold">{title}</h2>
            <p className="mt-2 text-sm leading-6 text-slate-600 dark:text-slate-300">{text}</p>
          </div>
        ))}
      </div>
    </div>
  )
}
