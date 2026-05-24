package com.talkent.app.data.api

import com.talkent.app.util.TokenManager
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.runBlocking
import okhttp3.Interceptor
import okhttp3.Response

class AuthInterceptor(
    private val tokenManager: TokenManager,
    private val baseUrlProvider: () -> String
) : Interceptor {

    private val _logoutEvent = MutableStateFlow(false)
    val logoutEvent: StateFlow<Boolean> = _logoutEvent

    override fun intercept(chain: Interceptor.Chain): Response {
        val originalRequest = chain.request()
        val path = originalRequest.url.encodedPath

        // Skip auth for public endpoints
        if (path == "/health" || path == "/api/auth/login" || path == "/api/auth/refresh") {
            return chain.proceed(originalRequest)
        }

        val token = tokenManager.getAccessToken()
        val request = if (token != null) {
            originalRequest.newBuilder()
                .header("Authorization", "Bearer $token")
                .build()
        } else {
            originalRequest
        }

        val response = chain.proceed(request)

        if (response.code == 401) {
            // Try refresh
            val refreshToken = tokenManager.getRefreshToken()
            if (refreshToken != null) {
                val newToken = refreshAccessToken(refreshToken)
                if (newToken != null) {
                    tokenManager.setAccessToken(newToken)
                    // Retry with new token
                    response.close()
                    val retryRequest = originalRequest.newBuilder()
                        .header("Authorization", "Bearer $newToken")
                        .build()
                    return chain.proceed(retryRequest)
                }
            }
            // Refresh failed — trigger logout
            tokenManager.clearTokens()
            _logoutEvent.value = true
        }

        return response
    }

    private fun refreshAccessToken(refreshToken: String): String? {
        return try {
            val client = okhttp3.OkHttpClient()
            val request = okhttp3.Request.Builder()
                .url("${baseUrlProvider()}/api/auth/refresh")
                .header("Content-Type", "application/json")
                .post(
                    okhttp3.RequestBody.create(
                        okhttp3.MediaType.parse("application/json")!!,
                        """{"refresh_token":"$refreshToken"}"""
                    )
                )
                .build()
            val response = client.newCall(request).execute()
            if (response.isSuccessful) {
                val body = response.body?.string() ?: return null
                // Simple JSON extraction without Moshi dependency in interceptor
                val tokenStart = body.indexOf("\"access_token\":\"")
                if (tokenStart == -1) return null
                val tokenBegin = tokenStart + 16
                val tokenEnd = body.indexOf("\"", tokenBegin)
                if (tokenEnd == -1) return null
                body.substring(tokenBegin, tokenEnd)
            } else null
        } catch (e: Exception) {
            null
        }
    }
}
