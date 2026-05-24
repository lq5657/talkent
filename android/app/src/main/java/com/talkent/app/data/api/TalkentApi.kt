package com.talkent.app.data.api

import com.talkent.app.data.model.*
import retrofit2.Response
import retrofit2.http.*

interface TalkentApi {

    // Auth
    @POST("api/auth/login")
    suspend fun login(@Body request: LoginRequest): TokenResponse

    @POST("api/auth/refresh")
    suspend fun refreshToken(@Body request: RefreshRequest): RefreshResponse

    // Roles
    @POST("api/roles/recommend-goals")
    suspend fun recommendGoals(@Body request: RecommendGoalsRequest): RecommendGoalsResponse

    @POST("api/roles/recommend-dimensions")
    suspend fun recommendDimensions(@Body request: RecommendDimensionsRequest): RecommendDimensionsResponse

    // Sessions
    @POST("api/sessions")
    suspend fun createSession(@Body request: CreateSessionRequest): CreateSessionResponse

    @POST("api/sessions/{id}/chat")
    suspend fun chat(@Path("id") sessionId: String, @Body request: ChatRequest): ChatResponse

    @POST("api/sessions/{id}/end")
    suspend fun endSession(@Path("id") sessionId: String): EndSessionResponse

    @GET("api/sessions/{id}")
    suspend fun getSession(@Path("id") sessionId: String): SessionDetail

    // Analysis
    @POST("api/sessions/{id}/analyze")
    suspend fun analyze(@Path("id") sessionId: String): AnalyzeResponse

    @GET("api/sessions/{id}/report")
    suspend fun getReport(@Path("id") sessionId: String): ReportResponse

    @GET("api/sessions/{id}/reports")
    suspend fun getReports(@Path("id") sessionId: String): List<ReportSummary>

    // Health
    @GET("health")
    suspend fun health(): Response<Map<String, String>>
}
