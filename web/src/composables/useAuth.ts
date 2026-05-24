const TOKEN_KEY = 'talkent_access_token'
const REFRESH_KEY = 'talkent_refresh_token'

export function useAuth() {
  function getAccessToken(): string | null {
    return localStorage.getItem(TOKEN_KEY)
  }

  function getRefreshToken(): string | null {
    return localStorage.getItem(REFRESH_KEY)
  }

  function setTokens(accessToken: string, refreshToken: string): void {
    localStorage.setItem(TOKEN_KEY, accessToken)
    localStorage.setItem(REFRESH_KEY, refreshToken)
  }

  function clearTokens(): void {
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(REFRESH_KEY)
  }

  function isAuthenticated(): boolean {
    return getAccessToken() !== null
  }

  async function refreshAccessToken(): Promise<string | null> {
    const refreshToken = getRefreshToken()
    if (!refreshToken) return null

    try {
      const res = await fetch('/api/auth/refresh', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: refreshToken }),
      })
      if (!res.ok) {
        clearTokens()
        return null
      }
      const data = await res.json()
      localStorage.setItem(TOKEN_KEY, data.access_token)
      return data.access_token
    } catch {
      return null
    }
  }

  return { getAccessToken, getRefreshToken, setTokens, clearTokens, isAuthenticated, refreshAccessToken }
}
