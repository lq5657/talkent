package com.talkent.app.ui.chat

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.talkent.app.data.model.ChatStreamEvent
import com.talkent.app.data.repository.SessionRepo
import com.talkent.app.util.SpeechRecorder
import com.talkent.app.util.TtsPlayer
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
    val error: String? = null,
    val isRecording: Boolean = false,
    val voicePartial: String = "",
    val voiceEnabled: Boolean = true,
    val ttsAvailable: Boolean = true
)

class ChatViewModel(
    private val application: Application,
    private val sessionId: String,
    private val sessionRepo: SessionRepo
) : AndroidViewModel(application) {

    private val _uiState = MutableStateFlow(ChatUiState())
    val uiState: StateFlow<ChatUiState> = _uiState.asStateFlow()

    val speechRecorder = SpeechRecorder(application)
    val ttsPlayer = TtsPlayer(application)

    init {
        loadSession()
        ttsPlayer.initialize()
    }

    // --- Voice recording ---

    fun startRecording() {
        ttsPlayer.stop() // Interrupt TTS
        speechRecorder.startListening { text ->
            _uiState.value = _uiState.value.copy(inputText = text, isRecording = false)
            if (text.isNotBlank()) sendMessage()
        }
        _uiState.value = _uiState.value.copy(isRecording = true, voicePartial = "")
    }

    fun stopRecording() {
        speechRecorder.stopListening()
    }

    // --- Voice state sync from SpeechRecorder flow ---

    fun collectSpeechState() {
        viewModelScope.launch {
            speechRecorder.state.collect { speechState ->
                _uiState.value = _uiState.value.copy(
                    isRecording = speechState.isListening,
                    voicePartial = speechState.partialResult,
                    voiceEnabled = speechState.error !is String || speechState.error != "需要录音权限"
                )
            }
        }
    }

    override fun onCleared() {
        super.onCleared()
        speechRecorder.destroy()
        ttsPlayer.shutdown()
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
                        // Auto-play TTS
                        ttsPlayer.speak(finalContent)
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
