export function Loading({ label = 'Loading...' }) {
  return (
    <div className="mx-auto max-w-3xl px-4 py-12" role="status" aria-label={label}>
      <div className="animate-pulse space-y-4">
        <div className="h-7 w-40 rounded bg-slate-200 dark:bg-slate-800" />
        <div className="h-36 rounded-2xl bg-slate-200 dark:bg-slate-800" />
      </div>
    </div>
  )
}
