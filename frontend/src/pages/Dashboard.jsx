import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Plus, ExternalLink, Trash2, Lock } from 'lucide-react'
import { api } from '../api/client'
import { Loading } from '../components/Loading'

export default function Dashboard() {
  const [polls, setPolls] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const navigate = useNavigate()

  const load = async () => {
    setLoading(true)
    try {
      const data = await api.listMine()
      setPolls(data.polls)
    } catch (loadError) {
      setError(loadError.message)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    let active = true
    api.listMine()
      .then((data) => {
        if (active) setPolls(data.polls)
      })
      .catch((loadError) => {
        if (active) setError(loadError.message)
      })
      .finally(() => {
        if (active) setLoading(false)
      })
    return () => { active = false }
  }, [])

  const closePoll = async (slug) => {
    await api.closePoll(slug)
    load()
  }

  const deletePoll = async (slug) => {
    if (!window.confirm('Delete this poll permanently?')) return
    await api.deletePoll(slug)
    setPolls((current) => current.filter((poll) => poll.slug !== slug))
  }

  if (loading) return <Loading label="Loading your polls" />
  return (
    <div className="mx-auto max-w-6xl px-4 py-10 sm:px-6">
      <div className="flex items-end justify-between gap-4">
        <div>
          <p className="text-sm font-bold uppercase tracking-wider text-indigo-600">My polls</p>
          <h1 className="mt-1 text-3xl font-black">Dashboard</h1>
        </div>
        <Link to="/create" className="inline-flex items-center gap-2 rounded-xl bg-indigo-600 px-4 py-3 font-bold text-white hover:bg-indigo-700"><Plus size={18} /> New poll</Link>
      </div>
      {error && <p role="alert" className="mt-6 rounded-xl bg-red-50 p-4 text-red-700 dark:bg-red-950 dark:text-red-300">{error}</p>}
      {polls.length === 0 ? (
        <div className="mt-8 rounded-2xl border border-dashed border-slate-300 p-10 text-center dark:border-slate-700">
          <h2 className="font-bold">No polls yet</h2>
          <p className="mt-2 text-sm text-slate-600 dark:text-slate-300">Create your first poll and share it with an audience.</p>
          <Link to="/create" className="mt-5 inline-flex items-center gap-2 rounded-xl bg-indigo-600 px-4 py-3 font-bold text-white"><Plus size={18} /> Create poll</Link>
        </div>
      ) : (
        <div className="mt-8 grid gap-4 md:grid-cols-2">
          {polls.map((poll) => (
            <article key={poll.id} className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm dark:border-slate-800 dark:bg-slate-900">
              <div className="flex items-start justify-between gap-3">
                <div>
                  <span className={poll.status === 'open' ? 'rounded-full bg-emerald-50 px-2.5 py-1 text-xs font-bold text-emerald-700 dark:bg-emerald-950 dark:text-emerald-300' : 'rounded-full bg-slate-100 px-2.5 py-1 text-xs font-bold text-slate-600 dark:bg-slate-800 dark:text-slate-300'}>{poll.status}</span>
                  <h2 className="mt-3 text-lg font-bold">{poll.question}</h2>
                  <p className="mt-1 text-sm text-slate-500">{poll.options.length} options</p>
                </div>
              </div>
              <div className="mt-5 flex flex-wrap gap-2">
                <button type="button" onClick={() => navigate(`/poll/${poll.slug}`)} className="inline-flex items-center gap-2 rounded-lg border border-slate-300 px-3 py-2 text-sm font-bold dark:border-slate-700"><ExternalLink size={15} /> Open</button>
                {poll.status === 'open' && <button type="button" onClick={() => closePoll(poll.slug)} className="inline-flex items-center gap-2 rounded-lg border border-slate-300 px-3 py-2 text-sm font-bold dark:border-slate-700"><Lock size={15} /> Close</button>}
                <button type="button" onClick={() => deletePoll(poll.slug)} className="inline-flex items-center gap-2 rounded-lg px-3 py-2 text-sm font-bold text-red-600 hover:bg-red-50 dark:hover:bg-red-950"><Trash2 size={15} /> Delete</button>
              </div>
            </article>
          ))}
        </div>
      )}
    </div>
  )
}
