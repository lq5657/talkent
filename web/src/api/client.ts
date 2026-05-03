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

async function apiRequest<T>(
  method: string,
  path: string,
  body?: unknown,
): Promise<T> {
  const opts: RequestInit = {
    method,
    headers: { 'Content-Type': 'application/json' },
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

export { ApiError }
