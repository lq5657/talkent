package com.talkent.app.data.repository

import com.talkent.app.data.api.SseClient
import com.talkent.app.data.api.TalkentApi
import com.talkent.app.data.model.*
import com.talkent.app.util.TokenManager
import com.talkent.app.util.UrlConfig
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.withContext

class SessionRepo(
    private val api: TalkentApi,
    private val sseClient: SseClient,
    private val tokenManager: TokenManager,
    private val urlConfig: UrlConfig
) {
    suspend fun recommendGoals(roleDescription: String): Result<RecommendGoalsResponse> = withContext(Dispatchers.IO) {
        runCatching { api.recommendGoals(RecommendGoalsRequest(roleDescription)) }
    }

    suspend fun recommendDimensions(
        roleType: String,
        goals: List<TrainingGoal>,
        mode: String,
        roleDesc: String
    ): Result<RecommendDimensionsResponse> = withContext(Dispatchers.IO) {
        runCatching {
            api.recommendDimensions(
                RecommendDimensionsRequest(
                    roleType = roleType,
                    goals = goals,
                    mode = mode,
                    roleDesc = roleDesc
                )
            )
        }
    }

    suspend fun createSession(
        roleDescription: String,
        scenario: String,
        roleType: String,
        goals: List<TrainingGoal>,
        dimensions: List<Dimension>,
        roundLimit: Int
    ): Result<CreateSessionResponse> = withContext(Dispatchers.IO) {
        runCatching {
            api.createSession(
                CreateSessionRequest(
                    roleDescription = roleDescription,
                    scenario = scenario,
                    roleType = roleType,
                    goals = goals,
                    dimensions = dimensions,
                    roundLimit = roundLimit
                )
            )
        }
    }

    suspend fun chat(sessionId: String, content: String): Result<ChatResponse> = withContext(Dispatchers.IO) {
        runCatching { api.chat(sessionId, ChatRequest(content)) }
    }

    fun chatStream(sessionId: String, content: String): Flow<ChatStreamEvent> {
        val token = tokenManager.getAccessToken() ?: ""
        return sseClient.chatStream(urlConfig.getBaseUrl(), sessionId, content, token)
    }

    suspend fun endSession(sessionId: String): Result<EndSessionResponse> = withContext(Dispatchers.IO) {
        runCatching { api.endSession(sessionId) }
    }

    suspend fun getSession(sessionId: String): Result<SessionDetail> = withContext(Dispatchers.IO) {
        runCatching { api.getSession(sessionId) }
    }

    suspend fun analyze(sessionId: String): Result<AnalyzeResponse> = withContext(Dispatchers.IO) {
        runCatching { api.analyze(sessionId) }
    }

    suspend fun getReport(sessionId: String): Result<ReportResponse> = withContext(Dispatchers.IO) {
        runCatching { api.getReport(sessionId) }
    }
}
