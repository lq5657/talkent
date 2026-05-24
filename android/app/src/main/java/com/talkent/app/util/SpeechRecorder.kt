package com.talkent.app.util

import android.content.Context
import android.content.Intent
import android.speech.RecognitionListener
import android.speech.RecognizerIntent
import android.speech.SpeechRecognizer
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

data class SpeechState(
    val isListening: Boolean = false,
    val partialResult: String = "",
    val finalResult: String? = null,
    val error: String? = null
)

class SpeechRecorder(private val context: Context) {

    private val recognizer: SpeechRecognizer = SpeechRecognizer.createSpeechRecognizer(context)
    private val _state = MutableStateFlow(SpeechState())
    val state: StateFlow<SpeechState> = _state.asStateFlow()

    private var onFinalResult: ((String) -> Unit)? = null

    init {
        recognizer.setRecognitionListener(object : RecognitionListener {
            override fun onReadyForSpeech(params: android.os.Bundle?) {
                _state.value = _state.value.copy(isListening = true, error = null, partialResult = "")
            }

            override fun onBeginningOfSpeech() {}

            override fun onRmsChanged(rmsdB: Float) {
                // Volume level for UI animation
            }

            override fun onBufferReceived(buffer: ByteArray?) {}

            override fun onEndOfSpeech() {
                _state.value = _state.value.copy(isListening = false)
            }

            override fun onError(error: Int) {
                val msg = when (error) {
                    SpeechRecognizer.ERROR_AUDIO -> "录音错误"
                    SpeechRecognizer.ERROR_CLIENT -> "客户端错误"
                    SpeechRecognizer.ERROR_INSUFFICIENT_PERMISSIONS -> "需要录音权限"
                    SpeechRecognizer.ERROR_NETWORK -> "网络错误"
                    SpeechRecognizer.ERROR_NETWORK_TIMEOUT -> "网络超时"
                    SpeechRecognizer.ERROR_NO_MATCH -> "未识别到语音"
                    SpeechRecognizer.ERROR_RECOGNIZER_BUSY -> "引擎忙碌"
                    SpeechRecognizer.ERROR_SERVER -> "服务端错误"
                    SpeechRecognizer.ERROR_SPEECH_TIMEOUT -> "语音超时"
                    else -> "识别失败"
                }
                _state.value = _state.value.copy(isListening = false, error = msg)
            }

            override fun onResults(results: android.os.Bundle?) {
                val matches = results?.getStringArrayList(SpeechRecognizer.RESULTS_RECOGNITION)
                val text = matches?.firstOrNull() ?: ""
                _state.value = _state.value.copy(
                    isListening = false,
                    finalResult = text,
                    partialResult = ""
                )
                onFinalResult?.invoke(text)
            }

            override fun onPartialResults(partialResults: android.os.Bundle?) {
                val matches = partialResults?.getStringArrayList(SpeechRecognizer.RESULTS_RECOGNITION)
                val text = matches?.firstOrNull() ?: ""
                _state.value = _state.value.copy(partialResult = text)
            }

            override fun onEvent(eventType: Int, params: android.os.Bundle?) {}
        })
    }

    fun startListening(onResult: (String) -> Unit) {
        onFinalResult = onResult
        val intent = Intent(RecognizerIntent.ACTION_RECOGNIZE_SPEECH).apply {
            putExtra(RecognizerIntent.EXTRA_LANGUAGE_MODEL, RecognizerIntent.LANGUAGE_MODEL_FREE_FORM)
            putExtra(RecognizerIntent.EXTRA_LANGUAGE, "zh-CN")
            putExtra(RecognizerIntent.EXTRA_PARTIAL_RESULTS, true)
        }
        recognizer.startListening(intent)
    }

    fun stopListening() {
        recognizer.stopListening()
    }

    fun destroy() {
        recognizer.destroy()
    }
}
