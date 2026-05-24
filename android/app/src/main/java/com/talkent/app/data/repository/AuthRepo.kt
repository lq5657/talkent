package com.talkent.app.data.repository

import com.talkent.app.data.api.TalkentApi
import com.talkent.app.data.model.LoginRequest
import com.talkent.app.util.TokenManager
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext

class AuthRepo(
    private val api: TalkentApi,
    private val tokenManager: TokenManager
) {
    suspend fun login(username: String, password: String): Result<Unit> = withContext(Dispatchers.IO) {
        try {
            val response = api.login(LoginRequest(username, password))
            tokenManager.setTokens(response.accessToken, response.refreshToken)
            Result.success(Unit)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    fun isLoggedIn(): Boolean = tokenManager.isAuthenticated()

    fun logout() {
        tokenManager.clearTokens()
    }
}
