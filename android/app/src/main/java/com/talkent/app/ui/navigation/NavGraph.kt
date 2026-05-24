package com.talkent.app.ui.navigation

import androidx.compose.runtime.Composable
import androidx.navigation.NavHostController
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.navArgument
import com.talkent.app.TalkentApp
import com.talkent.app.ui.chat.ChatScreen
import com.talkent.app.ui.login.LoginScreen
import com.talkent.app.ui.report.ReportScreen
import com.talkent.app.ui.settings.SettingsScreen
import com.talkent.app.ui.setup.SetupScreen

object Routes {
    const val LOGIN = "login"
    const val SETUP = "setup"
    const val CHAT = "chat/{sessionId}"
    const val REPORT = "report/{sessionId}"
    const val SETTINGS = "settings"

    fun chat(sessionId: String) = "chat/$sessionId"
    fun report(sessionId: String) = "report/$sessionId"
}

@Composable
fun NavGraph(
    navController: NavHostController,
    app: TalkentApp
) {
    NavHost(
        navController = navController,
        startDestination = if (app.authRepo.isLoggedIn()) Routes.SETUP else Routes.LOGIN
    ) {
        composable(Routes.LOGIN) {
            LoginScreen(
                authRepo = app.authRepo,
                onLoginSuccess = {
                    navController.navigate(Routes.SETUP) {
                        popUpTo(Routes.LOGIN) { inclusive = true }
                    }
                }
            )
        }

        composable(Routes.SETUP) {
            SetupScreen(
                sessionRepo = app.sessionRepo,
                onSessionCreated = { sessionId ->
                    navController.navigate(Routes.chat(sessionId))
                },
                onNavigateToSettings = {
                    navController.navigate(Routes.SETTINGS)
                }
            )
        }

        composable(
            route = Routes.CHAT,
            arguments = listOf(navArgument("sessionId") { type = NavType.StringType })
        ) { backStackEntry ->
            val sessionId = backStackEntry.arguments?.getString("sessionId") ?: return@composable
            ChatScreen(
                application = app,
                sessionId = sessionId,
                sessionRepo = app.sessionRepo,
                onNavigateToReport = { id ->
                    navController.navigate(Routes.report(id)) {
                        popUpTo(Routes.SETUP)
                    }
                },
                onNavigateBack = { navController.popBackStack() }
            )
        }

        composable(
            route = Routes.REPORT,
            arguments = listOf(navArgument("sessionId") { type = NavType.StringType })
        ) { backStackEntry ->
            val sessionId = backStackEntry.arguments?.getString("sessionId") ?: return@composable
            ReportScreen(
                sessionId = sessionId,
                sessionRepo = app.sessionRepo,
                onNavigateBack = {
                    navController.popBackStack(Routes.SETUP, inclusive = false)
                }
            )
        }

        composable(Routes.SETTINGS) {
            SettingsScreen(
                app = app,
                onNavigateBack = { navController.popBackStack() },
                onLoggedOut = {
                    navController.navigate(Routes.LOGIN) {
                        popUpTo(0) { inclusive = true }
                    }
                }
            )
        }
    }
}
