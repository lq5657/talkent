package com.talkent.app.data.local.entity

import androidx.room.ColumnInfo
import androidx.room.Entity
import androidx.room.ForeignKey
import androidx.room.Index
import androidx.room.PrimaryKey

@Entity(
    tableName = "reports",
    foreignKeys = [
        ForeignKey(
            entity = SessionEntity::class,
            parentColumns = ["session_id"],
            childColumns = ["session_id"],
            onDelete = ForeignKey.CASCADE
        )
    ],
    indices = [Index("session_id")]
)
data class ReportEntity(
    @PrimaryKey
    @ColumnInfo(name = "report_id")
    val reportId: Long,

    @ColumnInfo(name = "session_id")
    val sessionId: String,

    @ColumnInfo(name = "dimensions_json")
    val dimensionsJson: String,

    val markdown: String,

    @ColumnInfo(name = "model_used")
    val modelUsed: String,

    @ColumnInfo(name = "created_at")
    val createdAt: String
)
