import { useMemo, useState } from 'react'
import { AuthContext } from './authContextObject'

export function AuthProvider({ children }) {
  const [token, setToken] = useState(() => localStorage.getItem('poll_token'))
  const [user, setUser] = useState(() => {
    const saved = localStorage.getItem('poll_user')
    return saved ? JSON.parse(saved) : null
  })

  const loginSession = (nextToken, nextUser) => {
    localStorage.setItem('poll_token', nextToken)
    localStorage.setItem('poll_user', JSON.stringify(nextUser))
    setToken(nextToken)
    setUser(nextUser)
  }

  const logout = () => {
    localStorage.removeItem('poll_token')
    localStorage.removeItem('poll_user')
    setToken(null)
    setUser(null)
  }

  const value = useMemo(() => ({ token, user, loginSession, logout }), [token, user])
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

