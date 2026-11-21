import { Link } from 'react-router-dom'

export default function NotFound() {
  return (
    <div className="mx-auto max-w-xl px-4 py-20 text-center">
      <p className="text-sm font-black text-indigo-600">404</p>
      <h1 className="mt-2 text-3xl font-black">This page wandered off.</h1>
      <Link to="/" className="mt-6 inline-flex rounded-xl bg-indigo-600 px-5 py-3 font-bold text-white">Back home</Link>
    </div>
  )
}
