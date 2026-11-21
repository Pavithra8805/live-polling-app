import { useEffect, useRef, useState } from 'react'
import { api, streamUrl } from '../api/client'

export function useLiveResults(slug) {
  const [results, setResults] = useState(null)
  const [status, setStatus] = useState('loading')
  const [error, setError] = useState('')
  const sourceRef = useRef(null)

  useEffect(() => {
    if (!slug) return undefined
    let stopped = false

    const loadFullResults = async () => {
      try {
        const full = await api.getResults(slug)
        if (!stopped) setResults(full)
      } catch (loadError) {
        if (!stopped) setError(loadError.message)
      }
    }

    loadFullResults()
    sourceRef.current = new EventSource(streamUrl(slug), { withCredentials: true })
    sourceRef.current.addEventListener('results', (event) => {
      try {
        const update = JSON.parse(event.data)
        setResults((current) => current ? {
          ...current,
          ...update,
          poll: update.poll ? { ...current.poll, ...update.poll } : current.poll,
          counts: update.counts ? update.counts : current.counts,
          total: update.total ?? current.total,
        } : update)
        if (update.counts) {
          loadFullResults()
        }
        setStatus('live')
        setError('')
      } catch {
        setStatus('reconnecting')
      }
    })
    sourceRef.current.onopen = () => setStatus('live')
    sourceRef.current.onerror = () => {
      setStatus('reconnecting')
      loadFullResults()
    }

    return () => {
      stopped = true
      sourceRef.current?.close()
    }
  }, [slug])

  return { results, status, error }
}
