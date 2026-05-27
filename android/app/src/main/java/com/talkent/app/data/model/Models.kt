package com.talkent.app.data.model

import com.squareup.moshi.Json
import com.squareup.moshi.JsonClass

// --- Auth ---

@JsonClass(generateAdapter = true)
data class LoginRequest(
    val username: String,
    val password: String
)

@JsonClass(generateAdapter = true)
data class TokenResponse(
    @Json(name = "access_token") val accessToken: String,
    @Json(name = "refresh_token") val refreshToken: String
)

@JsonClass(generateAdapter = true)
data class RefreshRequest(
    @Json(name = "refresh_token") val refreshToken: String
)

@JsonClass(generateAdapter = true)
data class RefreshResponse(
    @Json(name = "access_token") val accessToken: String
)

// --- Role ---

@JsonClass(generateAdapter = true)
data class TrainingGoal(
    val name: String,
    val description: String
)

@JsonClass(generateAdapter = true)
data class Dimension(
    val name: String,
    val description: String
)

@JsonClass(generateAdapter = true)
data class RecommendGoalsRequest(
    @Json(name = "role_description") val roleDescription: String
)

@JsonClass(generateAdapter = true)
data class RecommendGoalsResponse(
    val source: String,
    val goals: List<TrainingGoal>
)

@JsonClass(generateAdapter = true)
data class RecommendDimensionsRequest(
    @Json(name = "role_type") val roleType: String,
    val goals: List<TrainingGoal>,
    val mode: String,
    @Json(name = "role_desc") val roleDesc: String
)

@JsonClass(generateAdapter = true)
data class RecommendDimensionsResponse(
    val source: String,
    val dimensions: List<Dimension>
)

// --- Session ---

@JsonClass(generateAdapter = true)
data class CreateSessionRequest(
    @Json(name = "role_description") val roleDescription: String,
    val scenario: String,
    @Json(name = "role_type") val roleType: String,
    val goals: List<TrainingGoal>,
    val dimensions: List<Dimension>,
    @Json(name = "round_limit") val roundLimit: Int
)

@JsonClass(generateAdapter = true)
data class CreateSessionResponse(
    @Json(name = "session_id") val sessionId: String,
    val status: String,
    @Json(name = "round_limit") val roundLimit: Int,
    @Json(name = "created_at") val createdAt: String
)

@JsonClass(generateAdapter = true)
data class ChatRequest(
    val content: String
)

@JsonClass(generateAdapter = true)
data class RoundInfo(
    val current: Int,
    val limit: Int,
    @Json(name = "is_last") val isLast: Boolean
)

@JsonClass(generateAdapter = true)
data class ChatResponse(
    val reply: String,
    @Json(name = "round_info") val roundInfo: RoundInfo,
    @Json(name = "memory_source") val memorySource: String,
    @Json(name = "user_message_created_at") val userMessageCreatedAt: String,
    @Json(name = "assistant_message_created_at") val assistantMessageCreatedAt: String
)

@JsonClass(generateAdapter = true)
data class SessionDetail(
    @Json(name = "session_id") val sessionId: String,
    val status: String,
    @Json(name = "role_description") val roleDescription: String,
    val scenario: String,
    @Json(name = "round_limit") val roundLimit: Int,
    @Json(name = "created_at") val createdAt: String
)

@JsonClass(generateAdapter = true)
data class EndSessionResponse(
    @Json(name = "session_id") val sessionId: String,
    val status: String,
    @Json(name = "final_round") val finalRound: Int
)

// --- Analysis ---

@JsonClass(generateAdapter = true)
data class AnalyzeResponse(
    @Json(name = "report_id") val reportId: Long,
    @Json(name = "session_id") val sessionId: String,
    val dimensions: List<DimensionAnalysis>,
    val markdown: String,
    @Json(name = "model_used") val modelUsed: String,
    @Json(name = "created_at") val createdAt: String
)

@JsonClass(generateAdapter = true)
data class DimensionAnalysis(
    val name: String,
    val description: String,
    val score: Int,
    val comment: String,
    val suggestions: List<String>
)

@JsonClass(generateAdapter = true)
data class ReportResponse(
    @Json(name = "report_id") val reportId: Long,
    @Json(name = "session_id") val sessionId: String,
    val dimensions: List<DimensionAnalysis>,
    val markdown: String,
    @Json(name = "model_used") val modelUsed: String,
    @Json(name = "created_at") val createdAt: String
)

@JsonClass(generateAdapter = true)
data class ReportSummary(
    @Json(name = "report_id") val reportId: Long,
    @Json(name = "created_at") val createdAt: String,
    @Json(name = "model_used") val modelUsed: String
)

// --- SSE Chat Stream ---

data class ChatStreamEvent(
    val type: String, // "token", "done", "error"
    val token: String? = null,
    val reply: String? = null,
    @Json(name = "round_info") val roundInfo: RoundInfo? = null,
    @Json(name = "memory_source") val memorySource: String? = null,
    @Json(name = "user_message_created_at") val userMessageCreatedAt: String? = null,
    @Json(name = "assistant_message_created_at") val assistantMessageCreatedAt: String? = null,
    val error: String? = null
)

// --- Error ---

@JsonClass(generateAdapter = true)
data class ApiError(
    val error: String
)

// --- Message (used across UI and data layers) ---

data class Message(
    val role: String, // "user", "assistant", or "assistant_streaming"
    val content: String,
    val createdAt: String = ""
)
