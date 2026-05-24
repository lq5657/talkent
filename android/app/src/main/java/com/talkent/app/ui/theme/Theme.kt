package com.talkent.app.ui.theme

import androidx.compose.material3.*
import androidx.compose.runtime.Composable

private val LightColorScheme = lightColorScheme(
    primary = Blue600,
    onPrimary = Gray50,
    primaryContainer = Blue50,
    secondary = Gray600,
    onSecondary = Gray900,
    surface = Gray50,
    onSurface = Gray900,
    background = Gray50,
    onBackground = Gray900,
    error = Red500,
    surfaceVariant = Gray100,
    outline = Gray200
)

@Composable
fun TalkentTheme(content: @Composable () -> Unit) {
    MaterialTheme(
        colorScheme = LightColorScheme,
        content = content
    )
}
