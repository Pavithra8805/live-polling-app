import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Plus, Trash2 } from 'lucide-react'
import { api } from '../api/client'
import { ResultBars } from '../components/ResultBars'

export default function CreatePoll() {
  const navigate = useNavigate()
  const [slug, setSlug] = useState('')
  const [question, setQuestion] = useState('')
  const [options, setOptions] = useState(['', ''])
  const [closesAt, setClosesAt] = useState('')
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  const updateOption = (index, value) => setOptions((current) => current.map((item, itemIndex) => itemIndex === index ? value : item))
  const addOption = () => setOptions((current) => current.length < 10 ? [...current, ''] : current)
  const removeOption = (index) => setOptions((current) => current.length > 2 ? current.filter((_, itemIndex) => itemIndex !== index) : current)

  const submit = async (event) => {
    event.preventDefault()
    setBusy(true)
    setError('')
    try {
      const poll = await api.createPoll({
        slug: slug.trim(),
        question: question.trim(),
        options: options.map((option) => option.trim()),
        closesAt: closesAt ? new Date(closesAt).toISOString() : null,
      })
      navigate(`/poll/${poll.slug}`)
    } catch (requestError) {
      setError(requestError.message)
    } finally {
      setBusy(false)
    }
  }

  const previewPoll = {
    slug: slug || 'your-poll',
    question: question || 'Your question will appear here',
    options: options.map((text, index) => ({ id: `option-${index + 1}`, text: text || `Option ${index + 1}` })),
  }
  const previewCounts = Object.fromEntries(previewPoll.options.map((option, index) => [option.id, index === 0 ? 2 : 1]))

  return (
    <div className="mx-auto grid max-w-6xl gap-6 px-4 py-10 sm:px-6 lg:grid-cols-[minmax(0,1fr)_minmax(280px,420px)]">
      <form onSubmit={submit} className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm dark:border-slate-800 dark:bg-slate-900 sm:p-7">
        <h1 className="text-3xl font-black">Create a poll</h1>
        <div className="mt-6 space-y-5">
          <label className="block text-sm font-semibold">Question
            <input required minLength={3} maxLength={200} value={question} onChange={(event) => setQuestion(event.target.value)} placeholder="What should we order?" className="mt-2 w-full rounded-xl border border-slate-300 bg-transparent px-3 py-3 dark:border-slate-700" />
          </label>
          <label className="block text-sm font-semibold">Share slug
            <input required pattern="[a-z0-9]+(?:-[a-z0-9]+)*" value={slug} onChange={(event) => setSlug(event.target.value.toLowerCase())} placeholder="team-lunch" className="mt-2 w-full rounded-xl border border-slate-300 bg-transparent px-3 py-3 dark:border-slate-700" />
            <span className="mt-1 block font-normal text-slate-500">Letters, numbers, and hyphens only.</span>
          </label>
          <fieldset>
            <legend className="text-sm font-semibold">Options</legend>
            <div className="mt-2 space-y-2">
              {options.map((option, index) => (
                <div key={index} className="flex gap-2">
                  <input required minLength={1} maxLength={100} value={option} onChange={(event) => updateOption(index, event.target.value)} aria-label={`Option ${index + 1}`} placeholder={`Option ${index + 1}`} className="min-w-0 flex-1 rounded-xl border border-slate-300 bg-transparent px-3 py-3 dark:border-slate-700" />
                  <button type="button" onClick={() => removeOption(index)} disabled={options.length <= 2} aria-label={`Remove option ${index + 1}`} className="grid w-11 place-items-center rounded-xl border border-slate-300 disabled:opacity-40 dark:border-slate-700"><Trash2 size={16} /></button>
                </div>
              ))}
            </div>
            <button type="button" onClick={addOption} disabled={options.length >= 10} className="mt-3 inline-flex items-center gap-2 rounded-lg px-3 py-2 text-sm font-bold text-indigo-600 disabled:opacity-40"><Plus size={16} /> Add option</button>
          </fieldset>
          <label className="block text-sm font-semibold">Optional close time
            <input type="datetime-local" value={closesAt} onChange={(event) => setClosesAt(event.target.value)} className="mt-2 w-full rounded-xl border border-slate-300 bg-transparent px-3 py-3 dark:border-slate-700" />
          </label>
          {error && <p role="alert" className="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700 dark:bg-red-950 dark:text-red-300">{error}</p>}
          <button disabled={busy} className="w-full rounded-xl bg-indigo-600 px-4 py-3 font-bold text-white hover:bg-indigo-700 disabled:opacity-60">{busy ? 'Creating...' : 'Create & share'}</button>
        </div>
      </form>
      <aside className="h-fit rounded-2xl border border-slate-200 bg-white p-5 dark:border-slate-800 dark:bg-slate-900 lg:sticky lg:top-24">
        <p className="text-xs font-bold uppercase tracking-wider text-indigo-600">Live preview</p>
        <h2 className="mt-3 font-black">{previewPoll.question}</h2>
        <div className="mt-5"><ResultBars poll={previewPoll} counts={previewCounts} total={3} /></div>
      </aside>
    </div>
  )
}
