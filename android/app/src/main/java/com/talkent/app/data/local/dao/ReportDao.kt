package com.talkent.app.data.local.dao

import androidx.room.*
import com.talkent.app.data.local.entity.ReportEntity

@Dao
interface ReportDao {

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insert(report: ReportEntity)

    @Query("SELECT * FROM reports WHERE session_id = :sessionId ORDER BY created_at DESC LIMIT 1")
    suspend fun getBySessionId(sessionId: String): ReportEntity?

    @Delete
    suspend fun delete(report: ReportEntity)
}
