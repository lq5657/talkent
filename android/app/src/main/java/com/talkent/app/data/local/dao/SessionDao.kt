package com.talkent.app.data.local.dao

import androidx.room.*
import com.talkent.app.data.local.entity.SessionEntity

@Dao
interface SessionDao {

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insert(session: SessionEntity)

    @Query("SELECT * FROM sessions ORDER BY created_at DESC")
    suspend fun getAll(): List<SessionEntity>

    @Query("SELECT * FROM sessions WHERE session_id = :id")
    suspend fun getById(id: String): SessionEntity?

    @Delete
    suspend fun delete(session: SessionEntity)
}
