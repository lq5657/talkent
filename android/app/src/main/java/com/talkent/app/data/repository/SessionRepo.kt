package com.talkent.app.data.repository

import com.talkent.app.data.api.SseClient
import com.talkent.app.data.api.TalkentApi
import com.talkent.app.data.local.TalkentDatabase
import com.talkent.app.data.local.entity.MessageEntity
import com.talkent.app.data.local.entity.ReportEntity
import com.talkent.app.data.local.entity.SessionEntity
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
    private val urlConfig: UrlConfig,
    private val database: TalkentDatabase
) {
    suspend fun recommendGoals(roleDescription: String): Result<RecommendGoalsResponse> = withContext(Dispatchers.IO) {
        runCatching { api.recommendGoals(RecommendGoalsRequest(roleDescription)) }
    }

    suspend fun recommendDimensions(
        roleType: String, goals: List<TrainingGoal>, mode: String, roleDesc: String
    ): Result<RecommendDimensionsResponse> = withContext(Dispatchers.IO) {
        runCatching {
            api.recommendDimensions(RecommendDimensionsRequest(roleType, goals, mode, roleDesc))
        }
    }

    suspend fun createSession(
        roleDescription: String, scenario: String, roleType: String,
        goals: List<TrainingGoal>, dimensions: List<Dimension>, roundLimit: Int
    ): Result<CreateSessionResponse> = withContext(Dispatchers.IO) {
        runCatching {
            api.createSession(CreateSessionRequest(roleDescription, scenario, roleType, goals, dimensions, roundLimit))
        }.onSuccess {
            database.sessionDao().insert(SessionEntity(
                sessionId = it.sessionId, roleDescription = roleDescription,
                scenario = scenario, status = "active", roundLimit = roundLimit,
                createdAt = it.createdAt
            ))
        }
    }

    suspend fun chat(sessionId: String, content: String): Result<ChatResponse> = withContext(Dispatchers.IO) {
        runCatching { api.chat(sessionId, ChatRequest(content)) }
            .onSuccess { resp ->
                val msgs = listOf(
                    MessageEntity(sessionId = sessionId, role = "user", content = content, createdAt = resp.userMessageCreatedAt),
                    MessageEntity(sessionId = sessionId, role = "assistant", content = resp.reply, createdAt = resp.assistantMessageCreatedAt)
                )
                database.messageDao().insertAll(msgs)
            }
    }

    fun chatStream(sessionId: String, content: String): Flow<ChatStreamEvent> {
        val token = tokenManager.getAccessToken() ?: ""
        return sseClient.chatStream(urlConfig.getBaseUrl(), sessionId, content, token)
    }

    suspend fun cacheStreamMessages(sessionId: String, userContent: String, userTime: String, assistantContent: String, assistantTime: String) {
        withContext(Dispatchers.IO) {
            database.messageDao().insertAll(listOf(
                MessageEntity(sessionId = sessionId, role = "user", content = userContent, createdAt = userTime),
                MessageEntity(sessionId = sessionId, role = "assistant", content = assistantContent, createdAt = assistantTime)
            ))
        }
    }

    suspend fun endSession(sessionId: String): Result<EndSessionResponse> = withContext(Dispatchers.IO) {
        runCatching { api.endSession(sessionId) }
    }

    // Cache-first: return cached session detail first, then API
    suspend fun getSession(sessionId: String): Result<SessionDetail> = withContext(Dispatchers.IO) {
        val cached = database.sessionDao().getById(sessionId)
        runCatching { api.getSession(sessionId) }.onSuccess { detail ->
            database.sessionDao().insert(SessionEntity(
                sessionId = detail.sessionId, roleDescription = detail.roleDescription,
                scenario = detail.scenario, status = detail.status,
                roundLimit = detail.roundLimit, createdAt = detail.createdAt
            ))
        }.recover { e ->
            if (cached != null) SessionDetail(cached.sessionId, cached.status, cached.roleDescription, cached.scenario, cached.roundLimit, cached.createdAt)
            else throw e
        }
    }

    // Cache-first: cached messages first, then API refresh
    suspend fun getMessages(sessionId: String): Result<List<Message>> = withContext(Dispatchers.IO) {
        val cached = database.messageDao().getBySessionId(sessionId).map {
            Message(role = it.role, content = it.content, createdAt = it.createdAt)
        }
        if (cached.isNotEmpty()) return@withContext Result.success(cached)
        Result.success(emptyList()) // API-only messages come via chat/chatStream
    }

    suspend fun analyze(sessionId: String): Result<AnalyzeResponse> = withContext(Dispatchers.IO) {
        runCatching { api.analyze(sessionId) }
    }

    // Cache-first: cached report first, then API refresh
    suspend fun getReport(sessionId: String): Result<ReportResponse> = withContext(Dispatchers.IO) {
        val cached = database.reportDao().getBySessionId(sessionId)
        runCatching { api.getReport(sessionId) }.onSuccess { report ->
            database.reportDao().insert(ReportEntity(
                reportId = report.reportId, sessionId = report.sessionId,
                dimensionsJson = "", markdown = report.markdown,
                modelUsed = report.modelUsed, createdAt = report.createdAt
            ))
        }.recover { e ->
            if (cached != null) ReportResponse(cached.reportId, cached.sessionId,
                emptyList(), cached.markdown, cached.modelUsed, cached.createdAt)
            else throw e
        }
    }
}
