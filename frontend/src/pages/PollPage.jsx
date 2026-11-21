import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { Copy, Check, QrCode, Radio } from 'lucide-react'
import { QRCodeSVG } from 'qrcode.react'
import { api } from '../api/client'
import { useLiveResults } from '../hooks/useLiveResults'
import { Loading } from '../components/Loading'
import { ResultBars } from '../components/ResultBars'

export default function PollPage() {
  const { slug } = useParams()
  const { results, status, error } = useLiveResults(slug)
  const [selected, setSelected] = useState('')
  const [voted, setVoted] = useState(false)
  const [actionError, setActionError] = useState('')
  const [copied, setCopied] = useState(false)
  const [showQr, setShowQr] = useState(false)

  const submitVote = async () => {
    setActionError('')
    try {
      await api.vote(slug, selected)
      setVoted(true)
    } catch (voteError) {
      setActionError(voteError.message)
    }
  }

  const copyLink = async () => {
    await navigator.clipboard.writeText(window.location.href)
    setCopied(true)
    setTimeout(() => setCopied(false), 1500)
  }

  if (!results && status === 'loading') return <Loading label="Loading poll" />
  if (error && !results) return <div className="mx-auto max-w-xl px-4 py-16"><p role="alert" className="rounded-xl bg-red-50 p-5 text-red-700 dark:bg-red-950 dark:text-red-300">{error}</p></div>
  if (!results) return null

  const poll = results.poll
  const closed = poll.status === 'closed'
  return (
    <div className="mx-auto max-w-3xl px-4 py-10 sm:px-6">
      <div className="mb-5 flex flex-wrap items-center justify-between gap-3">
        <span className={status === 'live' ? 'inline-flex items-center gap-2 rounded-full bg-emerald-50 px-3 py-1 text-sm font-bold text-emerald-700 dark:bg-emerald-950 dark:text-emerald-300' : 'inline-flex items-center gap-2 rounded-full bg-amber-50 px-3 py-1 text-sm font-bold text-amber-700 dark:bg-amber-950 dark:text-amber-300'}>
          <Radio size={14} /> {status === 'live' ? 'Live' : 'Reconnecting'}
        </span>
        <div className="flex gap-2">
          <button type="button" onClick={copyLink} className="inline-flex items-center gap-2 rounded-lg border border-slate-300 px-3 py-2 text-sm font-bold dark:border-slate-700">{copied ? <Check size={16} /> : <Copy size={16} />} {copied ? 'Copied' : 'Copy link'}</button>
          <button type="button" onClick={() => setShowQr((value) => !value)} className="inline-flex items-center gap-2 rounded-lg border border-slate-300 px-3 py-2 text-sm font-bold dark:border-slate-700"><QrCode size={16} /> QR</button>
        </div>
      </div>
      {showQr && <div className="mb-5 grid place-items-center rounded-2xl border border-slate-200 bg-white p-5 dark:border-slate-800 dark:bg-slate-900"><QRCodeSVG value={window.location.href} size={180} includeMargin /></div>}
      <article className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm dark:border-slate-800 dark:bg-slate-900 sm:p-7">
        <h1 className="text-2xl font-black sm:text-3xl">{poll.question}</h1>
        <p className="mt-2 text-sm text-slate-500">{results.total} total vote{results.total === 1 ? '' : 's'}</p>
        {!closed && !voted && (
          <fieldset className="mt-6">
            <legend className="mb-3 text-sm font-bold">Choose one option</legend>
            <div className="space-y-2">
              {poll.options.map((option) => (
                <label key={option.id} className={selected === option.id ? 'flex cursor-pointer items-center gap-3 rounded-xl border-2 border-indigo-500 bg-indigo-50 p-3 dark:bg-indigo-950' : 'flex cursor-pointer items-center gap-3 rounded-xl border border-slate-300 p-3 dark:border-slate-700'}>
                  <input type="radio" name="option" value={option.id} checked={selected === option.id} onChange={() => setSelected(option.id)} />
                  <span className="font-semibold">{option.text}</span>
                </label>
              ))}
            </div>
            {actionError && <p role="alert" className="mt-3 rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700 dark:bg-red-950 dark:text-red-300">{actionError}</p>}
            <button type="button" disabled={!selected} onClick={submitVote} className="mt-4 w-full rounded-xl bg-indigo-600 px-4 py-3 font-bold text-white disabled:cursor-not-allowed disabled:opacity-40">Submit vote</button>
          </fieldset>
        )}
        {voted && !closed && <p role="status" className="mt-5 rounded-xl bg-emerald-50 px-4 py-3 text-sm font-bold text-emerald-700 dark:bg-emerald-950 dark:text-emerald-300">Your vote is in. Results are updating live.</p>}
        {closed && <p className="mt-5 rounded-xl bg-slate-100 px-4 py-3 text-sm font-bold text-slate-700 dark:bg-slate-800 dark:text-slate-300">This poll is closed.</p>}
        <div className="mt-7">
          <ResultBars poll={poll} counts={results.counts} total={results.total} closed={closed} />
        </div>
      </article>
    </div>
  )
}
