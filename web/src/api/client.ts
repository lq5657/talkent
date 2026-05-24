import { useAuth } from '../composables/useAuth'

const API_BASE = '/api'

class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

function getAuthHeaders(): Record<string, string> {
  const token = useAuth().getAccessToken()
  if (token) {
    return { Authorization: `Bearer ${token}` }
  }
  return {}
}

let isRefreshing = false
let refreshPromise: Promise<string | null> | null = null

async function tryRefreshToken(): Promise<boolean> {
  if (isRefreshing) {
    const token = await refreshPromise
    return token !== null
  }
  isRefreshing = true
  refreshPromise = useAuth().refreshAccessToken()
  const token = await refreshPromise
  isRefreshing = false
  refreshPromise = null
  return token !== null
}

async function apiRequest<T>(
  method: string,
  path: string,
  body?: unknown,
): Promise<T> {
  const opts: RequestInit = {
    method,
    headers: {
      'Content-Type': 'application/json',
      ...getAuthHeaders(),
    },
  }
  if (body !== undefined) {
    opts.body = JSON.stringify(body)
  }

  let res: Response
  try {
    res = await fetch(`${API_BASE}${path}`, opts)
  } catch (e) {
    if (e instanceof TypeError) {
      throw new ApiError('网络连接失败，请检查后端服务是否启动', 0)
    }
    throw new ApiError('请求失败，请稍后重试', 0)
  }

  if (res.status === 401) {
    const refreshed = await tryRefreshToken()
    if (refreshed) {
      // Retry with new token
      const newHeaders = {
        'Content-Type': 'application/json',
        ...getAuthHeaders(),
      }
      const retryRes = await fetch(`${API_BASE}${path}`, { ...opts, headers: newHeaders })
      if (retryRes.ok) {
        const data = await retryRes.json()
        return data as T
      }
    }
    useAuth().clearTokens()
    window.location.href = '/login'
    throw new ApiError('登录已过期，请重新登录', 401)
  }

  if (!res.ok) {
    let message = '请求失败，请稍后重试'
    try {
      const data = await res.json()
      if (data.error) message = data.error
    } catch {
      // response body not JSON — use fallback message
    }
    throw new ApiError(message, res.status)
  }

  const data = await res.json()
  return data as T
}

function apiGet<T>(path: string): Promise<T> {
  return apiRequest<T>('GET', path)
}

function apiPost<T>(path: string, body?: unknown): Promise<T> {
  return apiRequest<T>('POST', path, body)
}

// --- Types ---

export interface TrainingGoal {
  name: string
  description: string
}

export interface Dimension {
  name: string
  description: string
}

export interface RecommendGoalsResponse {
  source: string
  goals: TrainingGoal[]
}

export interface RecommendDimensionsRequest {
  role_type: string
  goals: TrainingGoal[]
  mode: string
  role_desc: string
}

export interface RecommendDimensionsResponse {
  source: string
  dimensions: Dimension[]
}

export interface CreateSessionRequest {
  role_description: string
  scenario: string
  role_type: string
  goals: TrainingGoal[]
  dimensions: Dimension[]
  round_limit: number
}

export interface CreateSessionResponse {
  session_id: string
  status: string
  round_limit: number
  created_at: string
}

export interface ChatRequest {
  content: string
}

export interface RoundInfo {
  current: number
  limit: number
  is_last: boolean
}

export interface ChatResponse {
  reply: string
  round_info: RoundInfo
  memory_source: string
  user_message_created_at: string
  assistant_message_created_at: string
}

export interface EndSessionResponse {
  session_id: string
  status: string
  final_round: number
}

export interface SessionDetail {
  session_id: string
  status: string
  role_description: string
  scenario: string
  round_limit: number
  created_at: string
}

export interface DimensionAnalysis {
  name: string
  description: string
  score: number
  comment: string
  suggestions: string[]
}

export interface ReportResponse {
  report_id: number
  session_id: string
  dimensions: DimensionAnalysis[]
  markdown: string
  model_used: string
  created_at: string
}

export interface ReportSummary {
  report_id: number
  created_at: string
  model_used: string
}

export interface AnalyzeResponse {
  report_id: number
  session_id: string
  dimensions: DimensionAnalysis[]
  markdown: string
  model_used: string
  created_at: string
}

// --- API functions ---

export const api = {
  recommendGoals(roleDescription: string) {
    return apiPost<RecommendGoalsResponse>('/roles/recommend-goals', {
      role_description: roleDescription,
    })
  },

  recommendDimensions(req: RecommendDimensionsRequest) {
    return apiPost<RecommendDimensionsResponse>('/roles/recommend-dimensions', req)
  },

  createSession(req: CreateSessionRequest) {
    return apiPost<CreateSessionResponse>('/sessions', req)
  },

  chat(sessionId: string, content: string) {
    return apiPost<ChatResponse>(`/sessions/${sessionId}/chat`, { content })
  },

  endSession(sessionId: string) {
    return apiPost<EndSessionResponse>(`/sessions/${sessionId}/end`)
  },

  getSession(sessionId: string) {
    return apiGet<SessionDetail>(`/sessions/${sessionId}`)
  },

  analyze(sessionId: string) {
    return apiPost<AnalyzeResponse>(`/sessions/${sessionId}/analyze`)
  },

  getReport(sessionId: string) {
    return apiGet<ReportResponse>(`/sessions/${sessionId}/report`)
  },

  getReports(sessionId: string) {
    return apiGet<ReportSummary[]>(`/sessions/${sessionId}/reports`)
  },
}

export interface ChatStreamEvent {
  type: 'token' | 'done' | 'error'
  token?: string
  reply?: string
  round_info?: RoundInfo
  memory_source?: string
  user_message_created_at?: string
  assistant_message_created_at?: string
  error?: string
}

async function* chatStream(
  sessionId: string,
  content: string,
  signal?: AbortSignal,
): AsyncGenerator<ChatStreamEvent> {
  const token = useAuth().getAccessToken()
  const tokenParam = token ? `&token=${encodeURIComponent(token)}` : ''
  const res = await fetch(
    `${API_BASE}/sessions/${sessionId}/chat/stream?content=${encodeURIComponent(content)}${tokenParam}`,
    { signal },
  )

  if (!res.ok) {
    let message = '流式请求失败'
    try {
      const data = await res.json()
      if (data.error) message = data.error
    } catch { /* fallthrough */ }
    yield { type: 'error', error: message }
    return
  }

  const reader = res.body!.getReader()
  const decoder = new TextDecoder()
  let buffer = ''

  try {
    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() ?? ''

      for (const line of lines) {
        if (!line.startsWith('data: ')) continue
        const payload = line.slice(6)
        try {
          const event = JSON.parse(payload)
          if (event.error) {
            yield { type: 'error', error: event.error }
            return
          }
          if (event.done) {
            yield {
              type: 'done',
              reply: event.reply,
              round_info: event.round_info,
              memory_source: event.memory_source,
              user_message_created_at: event.user_message_created_at,
              assistant_message_created_at: event.assistant_message_created_at,
            }
            return
          }
          if (event.token) {
            yield { type: 'token', token: event.token }
          }
        } catch {
          // Skip unparseable events
        }
      }
    }
  } finally {
    reader.releaseLock()
  }
}

export { chatStream, ApiError }
