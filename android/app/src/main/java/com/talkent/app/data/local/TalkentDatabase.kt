package com.talkent.app.data.local

import android.content.Context
import androidx.room.Database
import androidx.room.Room
import androidx.room.RoomDatabase
import com.talkent.app.data.local.dao.MessageDao
import com.talkent.app.data.local.dao.ReportDao
import com.talkent.app.data.local.dao.SessionDao
import com.talkent.app.data.local.entity.MessageEntity
import com.talkent.app.data.local.entity.ReportEntity
import com.talkent.app.data.local.entity.SessionEntity

@Database(
    entities = [SessionEntity::class, MessageEntity::class, ReportEntity::class],
    version = 1,
    exportSchema = false
)
abstract class TalkentDatabase : RoomDatabase() {

    abstract fun sessionDao(): SessionDao
    abstract fun messageDao(): MessageDao
    abstract fun reportDao(): ReportDao

    companion object {
        @Volatile
        private var instance: TalkentDatabase? = null

        fun getInstance(context: Context): TalkentDatabase {
            return instance ?: synchronized(this) {
                instance ?: Room.databaseBuilder(
                    context.applicationContext,
                    TalkentDatabase::class.java,
                    "talkent_cache.db"
                ).build().also { instance = it }
            }
        }
    }
}
