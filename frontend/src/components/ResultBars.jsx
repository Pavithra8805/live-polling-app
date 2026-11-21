import { motion } from 'motion/react'

export function ResultBars({ poll, counts, total, closed = false }) {
  const highest = Math.max(0, ...poll.options.map((option) => counts[option.id] || 0))
  return (
    <div className="space-y-4">
      {poll.options.map((option) => {
        const count = counts[option.id] || 0
        const percentage = total > 0 ? Math.round((count / total) * 100) : 0
        const winner = closed && count === highest && highest > 0
        return (
          <div key={option.id} className={winner ? 'rounded-2xl border border-amber-300 bg-amber-50 p-3 dark:bg-amber-950/30' : ''}>
            <div className="mb-2 flex items-center justify-between gap-3">
              <span className="font-semibold">{option.text}</span>
              <span className="shrink-0 text-sm font-bold text-slate-600 dark:text-slate-300">{percentage}% <span className="font-normal">({count})</span></span>
            </div>
            <div className="h-3 overflow-hidden rounded-full bg-slate-200 dark:bg-slate-800">
              <motion.div
                className={winner ? 'h-full rounded-full bg-amber-500' : 'h-full rounded-full bg-indigo-500'}
                initial={{ width: 0 }}
                animate={{ width: `${percentage}%` }}
                transition={{ type: 'spring', stiffness: 140, damping: 22 }}
              />
            </div>
          </div>
        )
      })}
    </div>
  )
}
