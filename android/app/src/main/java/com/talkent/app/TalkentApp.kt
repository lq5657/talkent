package com.talkent.app

import android.app.Application
import com.squareup.moshi.Moshi
import com.squareup.moshi.kotlin.reflect.KotlinJsonAdapterFactory
import com.talkent.app.data.api.AuthInterceptor
import com.talkent.app.data.api.SseClient
import com.talkent.app.data.api.TalkentApi
import com.talkent.app.data.repository.AuthRepo
import com.talkent.app.data.repository.SessionRepo
import com.talkent.app.util.TokenManager
import com.talkent.app.util.UrlConfig
import okhttp3.OkHttpClient
import retrofit2.Retrofit
import retrofit2.converter.moshi.MoshiConverterFactory
import java.util.concurrent.TimeUnit

class TalkentApp : Application() {

    lateinit var tokenManager: TokenManager
    lateinit var urlConfig: UrlConfig
    lateinit var authRepo: AuthRepo
    lateinit var sessionRepo: SessionRepo
    lateinit var api: TalkentApi
    lateinit var sseClient: SseClient
    lateinit var authInterceptor: AuthInterceptor

    override fun onCreate() {
        super.onCreate()

        tokenManager = TokenManager(this)
        urlConfig = UrlConfig(this)

        val moshi = Moshi.Builder()
            .add(KotlinJsonAdapterFactory())
            .build()

        authInterceptor = AuthInterceptor(tokenManager) { urlConfig.getBaseUrl() }

        val okHttpClient = OkHttpClient.Builder()
            .connectTimeout(30, TimeUnit.SECONDS)
            .readTimeout(60, TimeUnit.SECONDS)
            .addInterceptor(authInterceptor)
            .build()

        val retrofit = Retrofit.Builder()
            .baseUrl(urlConfig.getBaseUrl() + "/")
            .client(okHttpClient)
            .addConverterFactory(MoshiConverterFactory.create(moshi))
            .build()

        api = retrofit.create(TalkentApi::class.java)
        sseClient = SseClient(moshi)
        authRepo = AuthRepo(api, tokenManager)
        sessionRepo = SessionRepo(api, sseClient, tokenManager, urlConfig)
    }

    fun updateBaseUrl(newUrl: String) {
        urlConfig.setBaseUrl(newUrl)
    }
}
