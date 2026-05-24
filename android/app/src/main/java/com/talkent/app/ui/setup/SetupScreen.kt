package com.talkent.app.ui.setup

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Settings
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.talkent.app.data.model.Dimension
import com.talkent.app.data.model.TrainingGoal
import com.talkent.app.data.repository.SessionRepo

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SetupScreen(
    sessionRepo: SessionRepo,
    onSessionCreated: (String) -> Unit,
    onNavigateToSettings: () -> Unit
) {
    val viewModel = remember { SetupViewModel(sessionRepo) }
    val state by viewModel.uiState.collectAsStateWithLifecycle()

    LaunchedEffect(state.createdSessionId) {
        state.createdSessionId?.let { onSessionCreated(it) }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Talkent") },
                actions = {
                    IconButton(onClick = onNavigateToSettings) {
                        Icon(Icons.Default.Settings, contentDescription = "设置")
                    }
                }
            )
        }
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(16.dp)
                .verticalScroll(rememberScrollState())
        ) {
            // Role description
            OutlinedTextField(
                value = state.roleDescription,
                onValueChange = viewModel::updateRoleDescription,
                label = { Text("角色描述") },
                modifier = Modifier.fillMaxWidth(),
                minLines = 3,
                placeholder = { Text("描述你想要的角色身份、性格、语言风格...") }
            )

            Spacer(Modifier.height(8.dp))

            OutlinedTextField(
                value = state.scenario,
                onValueChange = viewModel::updateScenario,
                label = { Text("场景 (可选)") },
                modifier = Modifier.fillMaxWidth(),
                placeholder = { Text("对话地点、时间、关系...") }
            )

            Spacer(Modifier.height(12.dp))

            // Role type
            Text("角色类型", style = MaterialTheme.typography.labelLarge)
            Spacer(Modifier.height(4.dp))
            Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                listOf("聊天伙伴", "技术面试官", "数学老师").forEach { type ->
                    FilterChip(
                        selected = state.roleType == type,
                        onClick = { viewModel.updateRoleType(type) },
                        label = { Text(type) }
                    )
                }
            }

            Spacer(Modifier.height(4.dp))

            // Round limit
            OutlinedTextField(
                value = state.roundLimit.toString(),
                onValueChange = { it.toIntOrNull()?.let(viewModel::updateRoundLimit) },
                label = { Text("对话轮数上限") },
                modifier = Modifier.width(100.dp),
                singleLine = true
            )

            Spacer(Modifier.height(16.dp))

            // Recommend goals
            Button(
                onClick = viewModel::recommendGoals,
                enabled = !state.isLoadingGoals,
                modifier = Modifier.fillMaxWidth()
            ) {
                if (state.isLoadingGoals) CircularProgressIndicator(modifier = Modifier.size(20.dp))
                else Text("推荐训练目标")
            }

            if (state.goals.isNotEmpty()) {
                Spacer(Modifier.height(8.dp))
                Text("选择训练目标", style = MaterialTheme.typography.labelLarge)
                state.goals.forEachIndexed { index, goal ->
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        Checkbox(
                            checked = state.selectedGoals.contains(index),
                            onCheckedChange = { viewModel.toggleGoal(index) }
                        )
                        Column {
                            Text(goal.name, style = MaterialTheme.typography.bodyMedium)
                            Text(goal.description, style = MaterialTheme.typography.bodySmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
                        }
                    }
                }
            }

            if (state.goals.isNotEmpty()) {
                Spacer(Modifier.height(12.dp))
                Button(
                    onClick = viewModel::recommendDimensions,
                    enabled = !state.isLoadingDimensions && state.selectedGoals.isNotEmpty(),
                    modifier = Modifier.fillMaxWidth()
                ) {
                    if (state.isLoadingDimensions) CircularProgressIndicator(modifier = Modifier.size(20.dp))
                    else Text("推荐分析维度")
                }
            }

            if (state.dimensions.isNotEmpty()) {
                Spacer(Modifier.height(8.dp))
                Text("分析维度 (${state.dimensions.size})", style = MaterialTheme.typography.labelLarge)
                state.dimensions.forEach { dim ->
                    Text("• ${dim.name}: ${dim.description}", style = MaterialTheme.typography.bodySmall)
                }
            }

            if (state.dimensions.isNotEmpty()) {
                Spacer(Modifier.height(20.dp))
                Button(
                    onClick = viewModel::createSession,
                    enabled = !state.isCreating,
                    modifier = Modifier.fillMaxWidth()
                ) {
                    if (state.isCreating) CircularProgressIndicator(modifier = Modifier.size(20.dp))
                    else Text("开始对话")
                }
            }

            state.error?.let { error ->
                Spacer(Modifier.height(8.dp))
                Text(error, color = MaterialTheme.colorScheme.error, style = MaterialTheme.typography.bodySmall)
            }
        }
    }
}
