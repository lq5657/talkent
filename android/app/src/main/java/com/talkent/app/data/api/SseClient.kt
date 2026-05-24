package com.talkent.app.data.api

import com.squareup.moshi.Moshi
import com.talkent.app.data.model.ChatStreamEvent
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow
import okhttp3.OkHttpClient
import okhttp3.Request
import java.io.BufferedReader
import java.io.InputStreamReader
import java.util.concurrent.TimeUnit

class SseClient(private val moshi: Moshi) {

    private val client = OkHttpClient.Builder()
        .connectTimeout(30, TimeUnit.SECONDS)
        .readTimeout(60, TimeUnit.SECONDS)
        .build()

    fun chatStream(baseUrl: String, sessionId: String, content: String, token: String): Flow<ChatStreamEvent> = flow {
        val url = "$baseUrl/api/sessions/$sessionId/chat/stream?content=$content&token=$token"
        val request = Request.Builder().url(url).get().build()

        val response = client.newCall(request).execute()
        if (!response.isSuccessful) {
            emit(ChatStreamEvent(type = "error", error = "stream request failed"))
            return@flow
        }

        val body = response.body ?: run {
            emit(ChatStreamEvent(type = "error", error = "empty body"))
            return@flow
        }

        val reader = BufferedReader(InputStreamReader(body.byteStream()))
        val adapter = moshi.adapter(ChatStreamEvent::class.java)

        try {
            var line: String?
            while (reader.readLine().also { line = it } != null) {
                val l = line ?: continue
                if (!l.startsWith("data: ")) continue
                val payload = l.removePrefix("data: ")
                try {
                    val event = adapter.fromJson(payload)
                    if (event != null) {
                        emit(event)
                        if (event.type == "done" || event.type == "error") break
                    }
                } catch (e: Exception) {
                    // Skip unparseable events
                }
            }
        } finally {
            reader.close()
            body.close()
        }
    }
}
