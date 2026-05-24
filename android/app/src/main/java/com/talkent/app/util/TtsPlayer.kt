package com.talkent.app.util

import android.content.Context
import android.media.AudioAttributes
import android.media.AudioFocusRequest
import android.media.AudioManager
import android.speech.tts.TextToSpeech
import android.speech.tts.UtteranceProgressListener
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import java.util.Locale

data class TtsState(
    val isSpeaking: Boolean = false,
    val isAvailable: Boolean = true,
    val utteranceId: String = ""
)

class TtsPlayer(private val context: Context) {

    private var tts: TextToSpeech? = null
    private val audioManager = context.getSystemService(Context.AUDIO_SERVICE) as AudioManager

    private val _state = MutableStateFlow(TtsState())
    val state: StateFlow<TtsState> = _state.asStateFlow()

    private var onDoneCallback: (() -> Unit)? = null

    fun initialize(onReady: () -> Unit = {}) {
        tts = TextToSpeech(context) { status ->
            if (status == TextToSpeech.SUCCESS) {
                val result = tts?.setLanguage(Locale.CHINESE)
                if (result == TextToSpeech.LANG_MISSING_DATA || result == TextToSpeech.LANG_NOT_SUPPORTED) {
                    _state.value = _state.value.copy(isAvailable = false)
                } else {
                    tts?.setAudioAttributes(
                        AudioAttributes.Builder()
                            .setUsage(AudioAttributes.USAGE_MEDIA)
                            .setContentType(AudioAttributes.CONTENT_TYPE_SPEECH)
                            .build()
                    )
                    onReady()
                }
            } else {
                _state.value = _state.value.copy(isAvailable = false)
            }
        }

        tts?.setOnUtteranceProgressListener(object : UtteranceProgressListener() {
            override fun onStart(utteranceId: String?) {
                _state.value = _state.value.copy(isSpeaking = true, utteranceId = utteranceId ?: "")
            }

            override fun onDone(utteranceId: String?) {
                _state.value = _state.value.copy(isSpeaking = false)
                onDoneCallback?.invoke()
            }

            @Deprecated("Deprecated in Java")
            override fun onError(utteranceId: String?) {
                _state.value = _state.value.copy(isSpeaking = false)
            }

            override fun onError(utteranceId: String?, errorCode: Int) {
                _state.value = _state.value.copy(isSpeaking = false)
            }
        })
    }

    fun speak(text: String, onDone: () -> Unit = {}) {
        if (!_state.value.isAvailable || tts == null) return
        onDoneCallback = onDone
        // Limit text length for TTS (first 500 chars of the reply)
        val utteranceId = System.currentTimeMillis().toString()
        tts?.speak(text.take(500), TextToSpeech.QUEUE_FLUSH, null, utteranceId)
    }

    fun stop() {
        tts?.stop()
        _state.value = _state.value.copy(isSpeaking = false)
    }

    fun shutdown() {
        tts?.stop()
        tts?.shutdown()
    }
}
