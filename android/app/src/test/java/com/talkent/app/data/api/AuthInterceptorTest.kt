package com.talkent.app.data.api

import com.talkent.app.util.TokenManager
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.runBlocking
import okhttp3.Interceptor
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.Protocol
import okhttp3.ResponseBody.Companion.toResponseBody
import okhttp3.Request
import okhttp3.Response
import org.junit.Assert.*
import org.junit.Before
import org.junit.Test
import org.mockito.kotlin.*

class AuthInterceptorTest {

    private lateinit var tokenManager: TokenManager
    private lateinit var interceptor: AuthInterceptor
    private lateinit var chain: Interceptor.Chain

    @Before
    fun setUp() {
        tokenManager = mock()
        interceptor = AuthInterceptor(tokenManager) { "http://10.0.2.2:8080" }
        chain = mock()
    }

    // --- Public endpoint skip ---

    @Test
    fun `skips auth for health endpoint`() {
        val request = Request.Builder().url("http://10.0.2.2:8080/health").build()
        val response = dummyResponse(request, 200)

        whenever(chain.request()).thenReturn(request)
        whenever(chain.proceed(request)).thenReturn(response)

        val result = interceptor.intercept(chain)
        assertEquals(200, result.code)
        verify(tokenManager, never()).getAccessToken()
    }

    @Test
    fun `skips auth for login endpoint`() {
        val request = Request.Builder().url("http://10.0.2.2:8080/api/auth/login").build()
        val response = dummyResponse(request, 200)

        whenever(chain.request()).thenReturn(request)
        whenever(chain.proceed(request)).thenReturn(response)

        interceptor.intercept(chain)
        verify(tokenManager, never()).getAccessToken()
    }

    @Test
    fun `skips auth for refresh endpoint`() {
        val request = Request.Builder().url("http://10.0.2.2:8080/api/auth/refresh").build()
        val response = dummyResponse(request, 200)

        whenever(chain.request()).thenReturn(request)
        whenever(chain.proceed(request)).thenReturn(response)

        interceptor.intercept(chain)
        verify(tokenManager, never()).getAccessToken()
    }

    // --- Token injection ---

    @Test
    fun `adds Bearer token to protected endpoints`() {
        val request = Request.Builder().url("http://10.0.2.2:8080/api/sessions/s1/chat").build()
        val response = dummyResponse(request, 200)

        whenever(tokenManager.getAccessToken()).thenReturn("test-token")
        whenever(chain.request()).thenReturn(request)
        whenever(chain.proceed(any())).thenReturn(response)

        interceptor.intercept(chain)

        val captor = argumentCaptor<Request>()
        verify(chain).proceed(captor.capture())
        assertEquals("Bearer test-token", captor.firstValue.header("Authorization"))
    }

    @Test
    fun `sends request without token when no access token available`() {
        val request = Request.Builder().url("http://10.0.2.2:8080/api/sessions").build()
        val response = dummyResponse(request, 200)

        whenever(tokenManager.getAccessToken()).thenReturn(null)
        whenever(chain.request()).thenReturn(request)
        whenever(chain.proceed(request)).thenReturn(response)

        interceptor.intercept(chain)

        val captor = argumentCaptor<Request>()
        verify(chain).proceed(captor.capture())
        assertNull(captor.firstValue.header("Authorization"))
    }

    // --- 401 token refresh ---

    @Test
    fun `on 401 with refresh token attempts server refresh and returns original 401 on failure`() {
        val request = Request.Builder().url("http://10.0.2.2:8080/api/sessions").build()
        val unauthResponse = dummyResponse(request, 401)

        whenever(tokenManager.getAccessToken()).thenReturn("expired")
        whenever(tokenManager.getRefreshToken()).thenReturn("refresh-token")
        whenever(chain.request()).thenReturn(request)
        whenever(chain.proceed(any())).thenReturn(unauthResponse)

        val result = interceptor.intercept(chain)

        assertEquals(401, result.code)
    }

    @Test
    fun `on 401 without refresh token triggers logout`() {
        val request = Request.Builder().url("http://10.0.2.2:8080/api/sessions").build()
        val unauthResponse = dummyResponse(request, 401)

        whenever(tokenManager.getAccessToken()).thenReturn("expired")
        whenever(tokenManager.getRefreshToken()).thenReturn(null)
        whenever(chain.request()).thenReturn(request)
        whenever(chain.proceed(any())).thenReturn(unauthResponse)

        interceptor.intercept(chain)

        verify(tokenManager).clearTokens()
        assertTrue(runBlocking { interceptor.logoutEvent.first() })
    }

    @Test
    fun `returns original 401 response if refresh token exists but refresh fails`() {
        val request = Request.Builder().url("http://10.0.2.2:8080/api/sessions").build()
        val unauthResponse = dummyResponse(request, 401)

        whenever(tokenManager.getAccessToken()).thenReturn("expired")
        whenever(tokenManager.getRefreshToken()).thenReturn("refresh-token")
        whenever(chain.request()).thenReturn(request)
        whenever(chain.proceed(any())).thenReturn(unauthResponse)

        val result = interceptor.intercept(chain)

        assertEquals(401, result.code)
        verify(tokenManager).clearTokens()
        assertTrue(runBlocking { interceptor.logoutEvent.first() })
    }

    // --- Helper ---

    private fun dummyResponse(request: Request, code: Int): Response {
        return Response.Builder()
            .request(request)
            .protocol(Protocol.HTTP_1_1)
            .code(code)
            .message(if (code == 200) "OK" else "Unauthorized")
            .body("".toResponseBody("application/json".toMediaType()))
            .build()
    }
}
