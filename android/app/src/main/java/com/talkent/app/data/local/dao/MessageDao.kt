package com.talkent.app.data.local.dao

import androidx.room.*
import com.talkent.app.data.local.entity.MessageEntity

@Dao
interface MessageDao {

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertAll(messages: List<MessageEntity>)

    @Query("SELECT * FROM messages WHERE session_id = :sessionId ORDER BY id ASC")
    suspend fun getBySessionId(sessionId: String): List<MessageEntity>

    @Query("DELETE FROM messages WHERE session_id = :sessionId")
    suspend fun deleteBySessionId(sessionId: String)
}
