package com.talkent.app.data.model

import com.squareup.moshi.Moshi
import com.squareup.moshi.kotlin.reflect.KotlinJsonAdapterFactory
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test

class ModelsTest {

    private lateinit var moshi: Moshi

    @Before
    fun setUp() {
        moshi = Moshi.Builder()
            .addLast(KotlinJsonAdapterFactory())
            .build()
    }

    // --- Auth ---

    @Test
    fun `loginRequest serializes username and password`() {
        val json = moshi.adapter(LoginRequest::class.java).toJson(LoginRequest("alice", "secret"))
        assertTrue(json.contains("\"username\":\"alice\""))
        assertTrue(json.contains("\"password\":\"secret\""))
    }

    @Test
    fun `tokenResponse deserializes access and refresh tokens`() {
        val json = """{"access_token":"abc123","refresh_token":"xyz789"}"""
        val result = moshi.adapter(TokenResponse::class.java).fromJson(json)!!
        assertEquals("abc123", result.accessToken)
        assertEquals("xyz789", result.refreshToken)
    }

    @Test
    fun `refreshRequest serializes refresh token`() {
        val json = moshi.adapter(RefreshRequest::class.java).toJson(RefreshRequest("token-1"))
        assertTrue(json.contains("\"refresh_token\":\"token-1\""))
    }

    @Test
    fun `refreshResponse deserializes access token`() {
        val json = """{"access_token":"new-token"}"""
        val result = moshi.adapter(RefreshResponse::class.java).fromJson(json)!!
        assertEquals("new-token", result.accessToken)
    }

    // --- Role / Training ---

    @Test
    fun `trainingGoal deserializes name and description`() {
        val json = """{"name":"Empathy","description":"Show empathy to customers"}"""
        val result = moshi.adapter(TrainingGoal::class.java).fromJson(json)!!
        assertEquals("Empathy", result.name)
        assertEquals("Show empathy to customers", result.description)
    }

    @Test
    fun `recommendGoalsRequest serializes role description`() {
        val json = moshi.adapter(RecommendGoalsRequest::class.java).toJson(
            RecommendGoalsRequest("sales rep")
        )
        assertTrue(json.contains("\"role_description\":\"sales rep\""))

        val parsed = moshi.adapter(RecommendGoalsRequest::class.java).fromJson(json)
        assertEquals("sales rep", parsed?.roleDescription)
    }

    @Test
    fun `recommendGoalsResponse deserializes goals list`() {
        val json = """{"source":"ai","goals":[{"name":"Active Listening","description":"Listen carefully"}]}"""
        val result = moshi.adapter(RecommendGoalsResponse::class.java).fromJson(json)!!
        assertEquals("ai", result.source)
        assertEquals(1, result.goals.size)
        assertEquals("Active Listening", result.goals[0].name)
    }

    @Test
    fun `recommendDimensionsRequest serializes all fields`() {
        val req = RecommendDimensionsRequest(
            roleType = "sales",
            goals = listOf(TrainingGoal("G1", "desc")),
            mode = "auto",
            roleDesc = "a sales representative"
        )
        val json = moshi.adapter(RecommendDimensionsRequest::class.java).toJson(req)
        assertTrue(json.contains(""""role_type":"sales""""))
        assertTrue(json.contains(""""mode":"auto""""))
        assertTrue(json.contains(""""role_desc":"a sales representative""""))
    }

    // --- Session ---

    @Test
    fun `createSessionRequest serializes all fields`() {
        val req = CreateSessionRequest(
            roleDescription = "sales rep",
            scenario = "customer complaint",
            roleType = "sales",
            goals = listOf(TrainingGoal("G1", "d")),
            dimensions = listOf(Dimension("tone", "voice tone")),
            roundLimit = 10
        )
        val json = moshi.adapter(CreateSessionRequest::class.java).toJson(req)
        assertTrue(json.contains(""""round_limit":10"""))
    }

    @Test
    fun `createSessionResponse deserializes session data`() {
        val json = """{"session_id":"s1","status":"active","round_limit":10,"created_at":"2025-01-01T00:00:00Z"}"""
        val result = moshi.adapter(CreateSessionResponse::class.java).fromJson(json)!!
        assertEquals("s1", result.sessionId)
        assertEquals("active", result.status)
        assertEquals(10, result.roundLimit)
        assertEquals("2025-01-01T00:00:00Z", result.createdAt)
    }

    @Test
    fun `chatRequest serializes content`() {
        val json = moshi.adapter(ChatRequest::class.java).toJson(ChatRequest("hello"))
        assertTrue(json.contains("\"content\":\"hello\""))
    }

    @Test
    fun `chatResponse deserializes reply and round info`() {
        val json = """{
            "reply":"Hello!",
            "round_info":{"current":3,"limit":10,"is_last":false},
            "memory_source":"cache",
            "user_message_created_at":"2025-01-01T10:00:00Z",
            "assistant_message_created_at":"2025-01-01T10:00:01Z"
        }"""
        val result = moshi.adapter(ChatResponse::class.java).fromJson(json)!!
        assertEquals("Hello!", result.reply)
        assertEquals(3, result.roundInfo.current)
        assertEquals(10, result.roundInfo.limit)
        assertFalse(result.roundInfo.isLast)
        assertEquals("cache", result.memorySource)
        assertEquals("2025-01-01T10:00:00Z", result.userMessageCreatedAt)
        assertEquals("2025-01-01T10:00:01Z", result.assistantMessageCreatedAt)
    }

    @Test
    fun `roundInfo deserializes is_last marking last round`() {
        val json = """{"current":10,"limit":10,"is_last":true}"""
        val result = moshi.adapter(RoundInfo::class.java).fromJson(json)!!
        assertTrue(result.isLast)
        assertEquals(10, result.current)
    }

    @Test
    fun `sessionDetail deserializes from API response`() {
        val json = """{
            "session_id":"abc",
            "status":"active",
            "role_description":"sales rep",
            "scenario":"customer call",
            "round_limit":5,
            "created_at":"2025-06-01T12:00:00Z"
        }"""
        val result = moshi.adapter(SessionDetail::class.java).fromJson(json)!!
        assertEquals("abc", result.sessionId)
        assertEquals("active", result.status)
        assertEquals("sales rep", result.roleDescription)
        assertEquals("customer call", result.scenario)
        assertEquals(5, result.roundLimit)
    }

    @Test
    fun `endSessionResponse deserializes final round`() {
        val json = """{"session_id":"s1","status":"ended","final_round":8}"""
        val result = moshi.adapter(EndSessionResponse::class.java).fromJson(json)!!
        assertEquals("s1", result.sessionId)
        assertEquals("ended", result.status)
        assertEquals(8, result.finalRound)
    }

    // --- SSE Chat Stream ---

    @Test
    fun `chatStreamEvent deserializes token event`() {
        val json = """{"type":"token","token":"Hello"}"""
        val result = moshi.adapter(ChatStreamEvent::class.java).fromJson(json)!!
        assertEquals("token", result.type)
        assertEquals("Hello", result.token)
    }

    @Test
    fun `chatStreamEvent deserializes done event with reply and round info`() {
        val json = """{
            "type":"done",
            "reply":"Full reply here",
            "round_info":{"current":5,"limit":10,"is_last":false},
            "memory_source":"cache",
            "user_message_created_at":"2025-01-01T10:00:00Z",
            "assistant_message_created_at":"2025-01-01T10:00:02Z"
        }"""
        val result = moshi.adapter(ChatStreamEvent::class.java).fromJson(json)!!
        assertEquals("done", result.type)
        assertEquals("Full reply here", result.reply)
        assertEquals(5, result.roundInfo?.current)
        assertEquals("2025-01-01T10:00:02Z", result.assistantMessageCreatedAt)
    }

    @Test
    fun `chatStreamEvent deserializes error event`() {
        val json = """{"type":"error","error":"stream request failed"}"""
        val result = moshi.adapter(ChatStreamEvent::class.java).fromJson(json)!!
        assertEquals("error", result.type)
        assertEquals("stream request failed", result.error)
    }

    // --- Analysis ---

    @Test
    fun `dimensionAnalysis deserializes score and suggestions`() {
        val json = """{
            "name":"Empathy",
            "description":"Show empathy",
            "score":85,
            "comment":"Good empathy skills",
            "suggestions":["Use more empathetic language","Acknowledge feelings first"]
        }"""
        val result = moshi.adapter(DimensionAnalysis::class.java).fromJson(json)!!
        assertEquals(85, result.score)
        assertEquals(2, result.suggestions.size)
        assertEquals("Use more empathetic language", result.suggestions[0])
    }

    @Test
    fun `analyzeResponse deserializes full analysis`() {
        val json = """{
            "report_id":42,
            "session_id":"s1",
            "dimensions":[],
            "markdown":"# Report",
            "model_used":"gpt-4",
            "created_at":"2025-06-01T12:00:00Z"
        }"""
        val result = moshi.adapter(AnalyzeResponse::class.java).fromJson(json)!!
        assertEquals(42L, result.reportId)
        assertEquals("s1", result.sessionId)
        assertEquals("# Report", result.markdown)
        assertEquals("gpt-4", result.modelUsed)
        assertEquals("2025-06-01T12:00:00Z", result.createdAt)
    }

    // --- Error ---

    @Test
    fun `apiError deserializes error message`() {
        val json = """{"error":"invalid credentials"}"""
        val result = moshi.adapter(ApiError::class.java).fromJson(json)!!
        assertEquals("invalid credentials", result.error)
    }

    // --- Message ---

    @Test
    fun `message is a plain data class not requiring Moshi`() {
        val msg = Message(role = "user", content = "hello", createdAt = "2025-01-01T00:00:00Z")
        assertEquals("user", msg.role)
        assertEquals("hello", msg.content)
        assertEquals("2025-01-01T00:00:00Z", msg.createdAt)
    }

    // --- Edge cases ---

    @Test
    fun `tokenResponse handles unknown json keys gracefully`() {
        val json = """{"access_token":"abc","refresh_token":"xyz","extra_field":"ignored"}"""
        val result = moshi.adapter(TokenResponse::class.java).fromJson(json)
        assertNotNull(result)
        assertEquals("abc", result!!.accessToken)
    }

    @Test
    fun `chatStreamEvent handles null optional fields`() {
        val json = """{"type":"done"}"""
        val result = moshi.adapter(ChatStreamEvent::class.java).fromJson(json)!!
        assertEquals("done", result.type)
        assertNull(result.reply)
        assertNull(result.token)
        assertNull(result.roundInfo)
    }
}
