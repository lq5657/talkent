package com.talkent.app.util

import android.content.Context

class UrlConfig(context: Context) {

    private val prefs = context.getSharedPreferences("talkent_settings", Context.MODE_PRIVATE)

    fun getBaseUrl(): String {
        return prefs.getString(KEY_BASE_URL, DEFAULT_BASE_URL) ?: DEFAULT_BASE_URL
    }

    fun setBaseUrl(url: String) {
        prefs.edit().putString(KEY_BASE_URL, url.trimEnd('/')).apply()
    }

    companion object {
        private const val KEY_BASE_URL = "base_url"
        const val DEFAULT_BASE_URL = "http://10.0.2.2:8080"
    }
}
