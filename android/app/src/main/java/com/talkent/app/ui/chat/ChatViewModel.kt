package com.talkent.app.ui.chat

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.talkent.app.data.model.ChatStreamEvent
import com.talkent.app.data.model.EndSessionResponse
import com.talkent.app.data.repository.SessionRepo
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

data class Message(
    val role: String, // "user" or "assistant"
    val content: String,
    val createdAt: String = ""
)

data class ChatUiState(
    val messages: List<Message> = emptyList(),
    val inputText: String = "",
    val isStreaming: Boolean = false,
    val currentRound: Int = 0,
    val roundLimit: Int = 10,
    val isLast: Boolean = false,
    val isEnded: Boolean = false,
    val isLoading: Boolean = false,
    val reportId: Long? = null,
    val error: String? = null
)

class ChatViewModel(
    private val sessionId: String,
    private val sessionRepo: SessionRepo
) : ViewModel() {

    private val _uiState = MutableStateFlow(ChatUiState())
    val uiState: StateFlow<ChatUiState> = _uiState.asStateFlow()

    init {
        loadSession()
    }

    private fun loadSession() {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true)
            sessionRepo.getSession(sessionId)
                .onSuccess { detail ->
                    _uiState.value = _uiState.value.copy(
                        roundLimit = detail.roundLimit,
                        isLoading = false
                    )
                }
                .onFailure { e ->
                    _uiState.value = _uiState.value.copy(
                        error = e.message ?: "加载会话失败",
                        isLoading = false
                    )
                }
        }
    }

    fun updateInput(text: String) {
        _uiState.value = _uiState.value.copy(inputText = text, error = null)
    }

    fun sendMessage() {
        val content = _uiState.value.inputText.trim()
        if (content.isEmpty() || _uiState.value.isStreaming) return

        val userMsg = Message(role = "user", content = content)
        _uiState.value = _uiState.value.copy(
            messages = _uiState.value.messages + userMsg,
            inputText = "",
            isStreaming = true,
            error = null
        )

        // Streaming reply
        val streamingContent = StringBuilder()
        viewModelScope.launch {
            sessionRepo.chatStream(sessionId, content).collect { event ->
                when (event.type) {
                    "token" -> {
                        streamingContent.append(event.token ?: "")
                        val currentMsgs = _uiState.value.messages.toMutableList()
                        // Remove last assistant streaming msg if exists
                        if (currentMsgs.isNotEmpty() && currentMsgs.last().role == "assistant_streaming") {
                            currentMsgs.removeAt(currentMsgs.lastIndex)
                        }
                        currentMsgs.add(Message(role = "assistant_streaming", content = streamingContent.toString()))
                        _uiState.value = _uiState.value.copy(messages = currentMsgs)
                    }
                    "done" -> {
                        val currentMsgs = _uiState.value.messages.toMutableList()
                        if (currentMsgs.isNotEmpty() && currentMsgs.last().role == "assistant_streaming") {
                            currentMsgs.removeAt(currentMsgs.lastIndex)
                        }
                        val finalContent = event.reply ?: streamingContent.toString()
                        currentMsgs.add(Message(
                            role = "assistant",
                            content = finalContent,
                            createdAt = event.assistantMessageCreatedAt ?: ""
                        ))
                        _uiState.value = _uiState.value.copy(
                            messages = currentMsgs,
                            isStreaming = false,
                            currentRound = event.roundInfo?.current ?: _uiState.value.currentRound,
                            isLast = event.roundInfo?.isLast ?: false
                        )
                    }
                    "error" -> {
                        _uiState.value = _uiState.value.copy(
                            isStreaming = false,
                            error = event.error ?: "流式请求失败"
                        )
                    }
                }
            }
        }
    }

    fun endSession() {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true)
            sessionRepo.endSession(sessionId)
                .onSuccess { response ->
                    _uiState.value = _uiState.value.copy(
                        isEnded = true,
                        isLoading = false
                    )
                    // Auto trigger analysis
                    triggerAnalysis()
                }
                .onFailure { e ->
                    _uiState.value = _uiState.value.copy(
                        error = e.message ?: "结束会话失败",
                        isLoading = false
                    )
                }
        }
    }

    private fun triggerAnalysis() {
        viewModelScope.launch {
            sessionRepo.analyze(sessionId)
                .onSuccess { response ->
                    _uiState.value = _uiState.value.copy(reportId = response.reportId)
                }
                .onFailure { e ->
                    _uiState.value = _uiState.value.copy(
                        error = e.message ?: "分析失败"
                    )
                }
        }
    }
}
