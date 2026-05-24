package com.talkent.app.ui.report

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.talkent.app.data.model.DimensionAnalysis
import com.talkent.app.data.repository.SessionRepo

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ReportScreen(
    sessionId: String,
    sessionRepo: SessionRepo,
    onNavigateBack: () -> Unit
) {
    val viewModel = remember { ReportViewModel(sessionId, sessionRepo) }
    val state by viewModel.uiState.collectAsStateWithLifecycle()

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("分析报告") },
                navigationIcon = {
                    IconButton(onClick = onNavigateBack) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "返回")
                    }
                }
            )
        }
    ) { padding ->
        if (state.isLoading) {
            Box(modifier = Modifier.fillMaxSize().padding(padding), contentAlignment = androidx.compose.ui.Alignment.Center) {
                CircularProgressIndicator()
            }
        } else {
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(padding)
                    .padding(16.dp)
                    .verticalScroll(rememberScrollState())
            ) {
                // Dimension cards
                Text("维度评分", style = MaterialTheme.typography.titleMedium)
                Spacer(Modifier.height(8.dp))

                state.dimensions.forEach { dim ->
                    DimensionCard(dim)
                    Spacer(Modifier.height(8.dp))
                }

                Spacer(Modifier.height(16.dp))

                // Report markdown (plain text)
                Text("分析报告", style = MaterialTheme.typography.titleMedium)
                Spacer(Modifier.height(8.dp))
                Text(
                    text = state.markdown,
                    style = MaterialTheme.typography.bodyMedium
                )

                if (state.modelUsed.isNotEmpty()) {
                    Spacer(Modifier.height(16.dp))
                    Text(
                        "模型: ${state.modelUsed}",
                        style = MaterialTheme.typography.labelSmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            }
        }

        state.error?.let { error ->
            Snackbar(modifier = Modifier.padding(16.dp)) {
                Text(error)
            }
        }
    }
}

@Composable
fun DimensionCard(dim: DimensionAnalysis) {
    Card(modifier = Modifier.fillMaxWidth()) {
        Column(modifier = Modifier.padding(12.dp)) {
            Row {
                Text(dim.name, style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.Bold)
                Spacer(Modifier.weight(1f))
                Text("${dim.score}/10", style = MaterialTheme.typography.titleSmall, color = MaterialTheme.colorScheme.primary)
            }
            Spacer(Modifier.height(4.dp))
            Text(dim.comment, style = MaterialTheme.typography.bodySmall)

            if (dim.suggestions.isNotEmpty()) {
                Spacer(Modifier.height(4.dp))
                Text("改进建议:", style = MaterialTheme.typography.labelSmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
                dim.suggestions.forEach { sug ->
                    Text("• $sug", style = MaterialTheme.typography.bodySmall)
                }
            }

            // Score bar
            Spacer(Modifier.height(4.dp))
            LinearProgressIndicator(
                progress = { dim.score / 10f },
                modifier = Modifier.fillMaxWidth()
            )
        }
    }
}
