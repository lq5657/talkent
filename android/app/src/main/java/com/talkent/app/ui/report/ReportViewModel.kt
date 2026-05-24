package com.talkent.app.ui.report

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.talkent.app.data.model.DimensionAnalysis
import com.talkent.app.data.repository.SessionRepo
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

data class ReportUiState(
    val dimensions: List<DimensionAnalysis> = emptyList(),
    val markdown: String = "",
    val modelUsed: String = "",
    val isLoading: Boolean = false,
    val error: String? = null
)

class ReportViewModel(
    private val sessionId: String,
    private val sessionRepo: SessionRepo
) : ViewModel() {

    private val _uiState = MutableStateFlow(ReportUiState())
    val uiState: StateFlow<ReportUiState> = _uiState.asStateFlow()

    init {
        loadReport()
    }

    private fun loadReport() {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true)
            sessionRepo.getReport(sessionId)
                .onSuccess { report ->
                    _uiState.value = _uiState.value.copy(
                        dimensions = report.dimensions,
                        markdown = report.markdown,
                        modelUsed = report.modelUsed,
                        isLoading = false
                    )
                }
                .onFailure { e ->
                    _uiState.value = _uiState.value.copy(
                        error = e.message ?: "加载报告失败",
                        isLoading = false
                    )
                }
        }
    }
}
