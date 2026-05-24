package com.talkent.app.ui.settings

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.talkent.app.TalkentApp
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

data class SettingsUiState(
    val baseUrl: String = "",
    val isTesting: Boolean = false,
    val testResult: String? = null,
    val isLoggedIn: Boolean = false
)

class SettingsViewModel(private val app: TalkentApp) : ViewModel() {

    private val _uiState = MutableStateFlow(SettingsUiState(
        baseUrl = app.urlConfig.getBaseUrl(),
        isLoggedIn = app.authRepo.isLoggedIn()
    ))
    val uiState: StateFlow<SettingsUiState> = _uiState.asStateFlow()

    fun updateBaseUrl(url: String) {
        _uiState.value = _uiState.value.copy(baseUrl = url, testResult = null)
    }

    fun testConnection() {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isTesting = true, testResult = null)
            try {
                app.updateBaseUrl(_uiState.value.baseUrl)
                val response = app.api.health()
                if (response.isSuccessful) {
                    _uiState.value = _uiState.value.copy(
                        testResult = "连接成功: ${response.body()}",
                        isTesting = false
                    )
                } else {
                    _uiState.value = _uiState.value.copy(
                        testResult = "连接失败: HTTP ${response.code()}",
                        isTesting = false
                    )
                }
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(
                    testResult = "连接失败: ${e.message}",
                    isTesting = false
                )
            }
        }
    }

    fun logout() {
        app.authRepo.logout()
        _uiState.value = _uiState.value.copy(isLoggedIn = false)
    }
}
