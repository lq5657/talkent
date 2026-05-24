package com.talkent.app.ui.setup

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.talkent.app.data.model.Dimension
import com.talkent.app.data.model.TrainingGoal
import com.talkent.app.data.repository.SessionRepo
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

data class SetupUiState(
    val roleDescription: String = "",
    val scenario: String = "",
    val roleType: String = "聊天伙伴",
    val roundLimit: Int = 10,
    val goals: List<TrainingGoal> = emptyList(),
    val selectedGoals: Set<Int> = emptySet(),
    val dimensions: List<Dimension> = emptyList(),
    val isLoadingGoals: Boolean = false,
    val isLoadingDimensions: Boolean = false,
    val isCreating: Boolean = false,
    val createdSessionId: String? = null,
    val error: String? = null
)

class SetupViewModel(private val sessionRepo: SessionRepo) : ViewModel() {

    private val _uiState = MutableStateFlow(SetupUiState())
    val uiState: StateFlow<SetupUiState> = _uiState.asStateFlow()

    fun updateRoleDescription(value: String) {
        _uiState.value = _uiState.value.copy(roleDescription = value, error = null)
    }

    fun updateScenario(value: String) {
        _uiState.value = _uiState.value.copy(scenario = value, error = null)
    }

    fun updateRoleType(value: String) {
        _uiState.value = _uiState.value.copy(roleType = value, error = null)
    }

    fun updateRoundLimit(value: Int) {
        _uiState.value = _uiState.value.copy(roundLimit = value, error = null)
    }

    fun toggleGoal(index: Int) {
        val current = _uiState.value.selectedGoals.toMutableSet()
        if (current.contains(index)) current.remove(index) else current.add(index)
        _uiState.value = _uiState.value.copy(selectedGoals = current)
    }

    fun recommendGoals() {
        val desc = _uiState.value.roleDescription
        if (desc.isBlank()) {
            _uiState.value = _uiState.value.copy(error = "请输入角色描述")
            return
        }
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoadingGoals = true, error = null)
            sessionRepo.recommendGoals(desc)
                .onSuccess { response ->
                    _uiState.value = _uiState.value.copy(
                        goals = response.goals,
                        selectedGoals = emptySet(),
                        isLoadingGoals = false
                    )
                }
                .onFailure { e ->
                    _uiState.value = _uiState.value.copy(
                        error = e.message ?: "推荐目标失败",
                        isLoadingGoals = false
                    )
                }
        }
    }

    fun recommendDimensions() {
        val state = _uiState.value
        val selectedGoals = state.selectedGoals.map { state.goals[it] }
        if (selectedGoals.isEmpty()) {
            _uiState.value = _uiState.value.copy(error = "请至少选择一个目标")
            return
        }
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoadingDimensions = true, error = null)
            sessionRepo.recommendDimensions(
                roleType = state.roleType,
                goals = selectedGoals,
                mode = "recommend",
                roleDesc = state.roleDescription
            )
                .onSuccess { response ->
                    _uiState.value = _uiState.value.copy(
                        dimensions = response.dimensions,
                        isLoadingDimensions = false
                    )
                }
                .onFailure { e ->
                    _uiState.value = _uiState.value.copy(
                        error = e.message ?: "推荐维度失败",
                        isLoadingDimensions = false
                    )
                }
        }
    }

    fun createSession() {
        val state = _uiState.value
        if (state.dimensions.isEmpty()) {
            _uiState.value = _uiState.value.copy(error = "请先推荐维度")
            return
        }
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isCreating = true, error = null)
            val selectedGoals = state.selectedGoals.map { state.goals[it] }
            sessionRepo.createSession(
                roleDescription = state.roleDescription,
                scenario = state.scenario,
                roleType = state.roleType,
                goals = selectedGoals,
                dimensions = state.dimensions,
                roundLimit = state.roundLimit
            )
                .onSuccess { response ->
                    _uiState.value = _uiState.value.copy(
                        createdSessionId = response.sessionId,
                        isCreating = false
                    )
                }
                .onFailure { e ->
                    _uiState.value = _uiState.value.copy(
                        error = e.message ?: "创建会话失败",
                        isCreating = false
                    )
                }
        }
    }
}
