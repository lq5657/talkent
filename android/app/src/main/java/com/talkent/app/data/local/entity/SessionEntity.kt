package com.talkent.app.data.local.entity

import androidx.room.ColumnInfo
import androidx.room.Entity
import androidx.room.PrimaryKey

@Entity(tableName = "sessions")
data class SessionEntity(
    @PrimaryKey
    @ColumnInfo(name = "session_id")
    val sessionId: String,

    @ColumnInfo(name = "role_description")
    val roleDescription: String,

    val scenario: String,

    val status: String,

    @ColumnInfo(name = "round_limit")
    val roundLimit: Int,

    @ColumnInfo(name = "created_at")
    val createdAt: String
)
