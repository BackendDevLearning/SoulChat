# gorm 对 sql 的用法

## 初始化和链接

```golang
var db gorm.DB*
var err error

//注意这里，dsn为opensql的指令，dsn = "%s:%s@(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local"
connArgs := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",dbConfig.User, dbConfig.PassWord, dbConfig.Host, dbConfig.Port, dbConfig.DbName)

db, err = gorm.Open(mysql.Open(connArgs), &gorm.Config{})

sqlDB, err := db.DB()

sqlDB.SetMaxIdleConns(dbConfig.MaxIdleConn)
sqlDB.SetMaxOpenConns(dbConfig.MaxOpenConn)
sqlDB.SetConnMaxIdleTime(time.Duration(dbConfig.MaxIdleTime * int(time.Second)))
```





```lua
--thread_mgr.lua
local select         = select
local tsort          = table.sort
local tunpack        = table.unpack
local tpack          = table.pack
local tremove        = table.remove
local co_yield       = coroutine.yield
local co_create      = coroutine.create
local co_resume      = coroutine.resume
local co_running     = coroutine.running

local mrandom        = math_ext.random
local tsize          = table_ext.size
local hxpcall        = hive.xpcall
local log_err        = logger.err
local log_warn       = logger.warn

local QueueFIFO      = import("container/queue_fifo.lua")
local SyncLock       = import("kernel/internal/sync_lock.lua")

local MINUTE_MS      = hive.enum("PeriodTime", "MINUTE_MS")
local THREAD_TIMEOUT = hive.enum("KernCode", "THREAD_TIMEOUT")
local SYNC_PERFRAME  = 5

local ThreadMgr      = singleton()
local prop           = property(ThreadMgr)
prop:reader("session_id", 1)
prop:reader("entry_map", {})
prop:reader("syncqueue_map", {})
prop:reader("coroutine_waitings", {})
prop:reader("coroutine_yields", {})
prop:reader("coroutine_pool", nil)

function ThreadMgr:__init()
    self.session_id     = mrandom()
    self.coroutine_pool = setmetatable({}, { __mode = "kv" })
end

function ThreadMgr:idle_size()
    return #self.coroutine_pool
end

function ThreadMgr:wait_size()
    local co_yield_size = tsize(self.coroutine_yields)
    local co_wait_size  = tsize(self.coroutine_waitings)
    return co_yield_size + co_wait_size + 1
end

function ThreadMgr:lock_size()
    local count = 0
    for _, v in pairs(self.syncqueue_map) do
        count = count + v:size()
    end
    return count
end

function ThreadMgr:entry(key, func, ...)
    if self.entry_map[key] then
        return false
    end
    self:fork(function(...)
        self.entry_map[key] = true
        hxpcall(func, "entry:%s", ...)
        self.entry_map[key] = nil
    end)
    return true
end

function ThreadMgr:lock(key, waiting)
    local queue = self.syncqueue_map[key]
    if not queue then
        queue                   = QueueFIFO()
        queue.sync_num          = 0
        self.syncqueue_map[key] = queue
    end
    queue.ttl  = hive.clock_ms + MINUTE_MS
    local head = queue:head()
    if not head then
        local lock = SyncLock(self, key)
        queue:push(lock)
        return lock
    end
    if head.co == co_running() then
        --防止重入
        head:lock()
        return head
    end
    if waiting or waiting == nil then
        --等待则挂起
        local lock = SyncLock(self, key)
        queue:push(lock)
        co_yield()
        return lock
    end
end

function ThreadMgr:unlock(key, force)
    local queue = self.syncqueue_map[key]
    if not queue then
        return
    end
    local head = queue:head()
    if not head then
        return
    end
    if head.co == co_running() or force then
        queue:pop()
        local next = queue:head()
        if next then
            local sync_num = queue.sync_num
            if sync_num < SYNC_PERFRAME then
                queue.sync_num = sync_num + 1
                co_resume(next.co)
                return
            end
            self.coroutine_waitings[next.co] = 0
        end
        queue.sync_num = 0
    end
end

function ThreadMgr:co_create(f)
    local pool = self.coroutine_pool
    local co   = tremove(pool)
    if co == nil then
        co = co_create(function(...)
            hxpcall(f, "[ThreadMgr][co_create] fork error: %s", ...)
            while true do
                f               = nil
                pool[#pool + 1] = co
                f               = co_yield()
                if type(f) == "function" then
                    hxpcall(f, "[ThreadMgr][co_create] fork error: %s", co_yield())
                end
            end
        end)
    else
        co_resume(co, f)
    end
    return co
end

function ThreadMgr:check_delay(context)
    local clock_ms = hive.clock_ms
    if context.delay_ms and context.delay_ms < clock_ms then
        if context.title then
            log_warn("[ThreadMgr][check_delay] session_id({}:{}) cost time:{} more than want:{}",
                    context.session_id, context.title, clock_ms - context.stime, context.delay_ms - context.stime)
        end
    end
end

function ThreadMgr:try_response(session_id, ...)
    local context = self.coroutine_yields[session_id]
    if not context then
        return false
    end
    self.coroutine_yields[session_id] = nil
    self:check_delay(context)
    self:resume(context.co, ...)
    return true
end

function ThreadMgr:response(session_id, ...)
    if not self:try_response(session_id, ...) then
        log_err("[ThreadMgr][response][{}] unknown session_id({}) response!,[{}],from:[{}]", hive.frame, session_id, tpack(...), hive.where_call())
    end
end

function ThreadMgr:resume(co, ...)
    return co_resume(co, ...)
end

function ThreadMgr:yield(session_id, title, ms_to, ms_delay, ...)
    local context = { co = co_running(), title = title, to = hive.clock_ms + ms_to, stime = hive.clock_ms }
    if ms_delay then
        context.delay_ms = hive.clock_ms + ms_delay
    end
    self.coroutine_yields[session_id] = context
    return co_yield(...)
end

function ThreadMgr:get_title(session_id)
    local context = self.coroutine_yields[session_id]
    if context then
        return context.title
    end
    return nil
end

function ThreadMgr:on_second30(clock_ms)
    for key, queue in pairs(self.syncqueue_map) do
        if queue:empty() and clock_ms > queue.ttl then
            self.syncqueue_map[key] = nil
        end
    end
end

function ThreadMgr:on_second(clock_ms)
    --处理锁超时
    for key, queue in pairs(self.syncqueue_map) do
        local head = queue:head()
        if head and head:is_timeout(clock_ms) then
            self:unlock(key, true)
            log_err("[ThreadMgr][on_second] the lock is timeout:{},count:{},cost:{},queue:{}",
                    head.key, head.count, head:cost_time(clock_ms), queue:size())
        end
    end
    --检查协程超时
    local timeout_coroutines = {}
    for session_id, context in pairs(self.coroutine_yields) do
        if context.to <= clock_ms then
            context.session_id                          = session_id
            timeout_coroutines[#timeout_coroutines + 1] = context
        end
    end
    --处理协程超时
    if next(timeout_coroutines) then
        tsort(timeout_coroutines, function(a, b)
            return a.to < b.to
        end)
        for _, context in ipairs(timeout_coroutines) do
            local session_id = context.session_id
            if self:try_response(session_id, false, THREAD_TIMEOUT, session_id) then
                if context.title then
                    log_err("[ThreadMgr][on_second][{}] session_id({}:{}) timeout:{} !", hive.frame, session_id, context.title, clock_ms - context.stime)
                end
            end
        end
    end
end

function ThreadMgr:on_frame(clock_ms)
    --检查协程超时
    local timeout_coroutines = {}
    for co, ms_to in pairs(self.coroutine_waitings) do
        if ms_to <= clock_ms then
            timeout_coroutines[#timeout_coroutines + 1] = co
        end
    end
    --处理协程超时
    if next(timeout_coroutines) then
        for _, co in pairs(timeout_coroutines) do
            self.coroutine_waitings[co] = nil
            co_resume(co)
        end
    end
end

function ThreadMgr:fork(f, ...)
    local n = select("#", ...)
    local co
    if n == 0 then
        co = self:co_create(f)
    else
        local args = { ... }
        co         = self:co_create(function()
            f(tunpack(args, 1, n))
        end)
    end
    self:resume(co, ...)
    return co
end

function ThreadMgr:sleep(ms)
    local co                    = co_running()
    self.coroutine_waitings[co] = hive.clock_ms + ms
    co_yield()
end

function ThreadMgr:build_session_id()
    self.session_id = self.session_id + 1
    if self.session_id >= 0x7fffffff then
        self.session_id = 1
    end
    return self.session_id
end

function ThreadMgr:success_call(period, success_func, delay, try_times)
    if delay and delay > 0 then
        self:sleep(delay)
    end
    self:fork(function()
        try_times = try_times or 10
        while true do
            if success_func() or try_times <= 0 then
                break
            end
            try_times = try_times - 1
            self:sleep(period)
        end
    end)
end

hive.thread_mgr = ThreadMgr()

return ThreadMgr
```





## sql 的学习

### 主键 

1. 不能重复
2. 不为空

选取主键的一个基本原则是：不使用任何业务相关的字段作为主键

一般选取方式：

1. 自增整数类型：数据库会在插入数据时自动为每一条记录分配一个自增整数，这样我们就完全不用担心主键重复，也不用自己预先生成主键；
2. 全局唯一GUID类型：也称UUID，使用一种全局唯一的字符串作为主键，类似`8f55d96b-8acc-4636-8cb8-76bf8abc2f57`。GUID算法通过**网卡MAC**地址、**时间戳**和**随机数**保证任意计算机在任意时间生成的字符串都是不同的，大部分编程语言都内置了GUID算法，可以自己预算出主键。

#### 联合主键

允许通过多个字段唯一标识记录，即两个或更多的字段都设置为主键，这种主键被称为联合主键。

对于联合主键，允许一列有重复，只要不是所有主键列都重复即可

```
在docker中进sql
docker exec -it <name> sql -u root -p 

创建库
create database mydb;

创建表
create table user (
	id int auto_increment peimary key,
	name varchar(50) not null,
	email varchar(100) unique,
	created_at timestamp default current_timestamp
);
```

#### 外键约束

两表关系有

1 对 1

1 对 多

多 对 多



1 对 多，一个班级对多个学生

 

```sql
alter table sudents
add constraint class_id
foreign key (class_ids)
refrences classes (id);
```



删除外键约束

```sql
alter table students
drop foreign key class_ids;
```



注意：删除外键约束并没有删除外键这一列。删除列是通过`DROP COLUMN ...`实现的







## 社交功能

### 需求

关注和取关功能，相互关注才可以添加好友



### 采用方案

先写 sql，再写 redis 



### 可能出现的问题

#### **场景 1：关注操作重复写入**

两个请求几乎同时关注同一个人：

```
A关注B   ----+
             |----> MySQL 插入第一条成功
A关注B   ----+
             |----> MySQL 插入第二条报唯一索引错误
```

- 结果：MySQL 只保留一条记录 ✅
- Redis：由于后面 SADD 是幂等的，最终结果也是一致的 ✅

> 这个场景**问题不大**，因为 SADD 自动去重。

#### **场景 2：Redis 写失败**

```
MySQL 插入关注关系成功
 ↓
Redis 网络闪断，SADD 失败
```

- 结果：MySQL 有记录，Redis 没记录 → **Redis 变成脏数据**
- 后续读取时，会**误判为未关注**，导致用户界面显示错误。

> 这类**数据丢失问题很常见**，必须通过异步补偿或定期修复解决。

#### **场景 3：并发+取关竞态**

A 发起「关注」，B 同时发起「取关」。

1. A 的 `FollowUser` 执行：
   - MySQL 插入成功。
   - Redis **还未 SADD**。
2. B 的 `UnfollowUser` 执行：
   - MySQL 删除成功。
   - Redis `SREM` **执行成功**。
3. A 的 Redis SADD **晚于** B 的 SREM。
   - Redis 变成有这条记录，但 MySQL 已经删除，**最终不一致**。

> **这是最棘手的并发问题**，典型的“写后写”冲突。
>
> 

## **解决思路**

| 方案                          | 难度 | 一致性   | 性能 | 适用场景               |
| ----------------------------- | ---- | -------- | ---- | ---------------------- |
| **1. 双写+定期修复**（当前）  | 简单 | 弱一致   | 高   | 数据实时一致性要求不高 |
| **2. 先删 Redis 再写 MySQL**  | 中   | 较好     | 高   | 适合计数型缓存         |
| **3. Redis 作为唯一写入口**   | 高   | 强一致   | 中   | 高并发核心业务         |
| **4. 消息队列异步更新 Redis** | 中   | 最终一致 | 中   | 高吞吐场景             |
| **5. Lua 或事务锁**           | 中   | 强一致   | 中   | 小规模关键场景         |



### **方案 1：双写+定期修复（最常用）**

你的流程保持不变，增加一个**异步修复任务**：

- Redis 写失败只记录日志，不回滚 MySQL。
- 定时从 MySQL 全量扫描 → 和 Redis 校对 → 修复脏数据。

> 这种方案在大部分社交场景**足够用了**，因为 Redis 主要做缓存，不做最终权威存储。

------

### **方案 2：先删 Redis，再写 MySQL**

如果 Redis 中缓存了计数或关注状态，更新时可以**先删缓存，再写数据库**：

```
DEL user:{uid}:following
DEL user:{uid}:followers
↓
UPDATE MySQL
```

- 下次读时发现 Redis 没值，会从 MySQL 加载，保证读到的是最新值。
- **优点**：避免 Redis 脏数据。
- **缺点**：会有短暂缓存缺失，导致**缓存击穿**。

------

### **方案 3：Redis 作为唯一入口（强一致）**

所有写操作**只写 Redis**，通过异步 binlog 或消息队列同步到 MySQL。

1. 用户关注操作：
   - Redis 原子写入集合。
   - 同时写入 Redis Stream/Kafka 作为事件。
2. 后台消费事件写 MySQL。

> **典型案例：微信朋友圈、微博点赞**
>  优点：强一致、高性能。
>  缺点：架构复杂度高，需要运维消息队列。

------

### **方案 4：消息队列异步更新 Redis**

1. 只保证 MySQL 一定成功。
2. Redis 更新通过 Kafka/RabbitMQ 之类的 MQ 异步执行。
3. 即使 Redis 更新失败，MQ 也会重试，保证最终一致。

> 缺点：Redis 不是实时更新，会有延迟。

------

### **方案 5：Lua 脚本或分布式锁**

使用 Lua 将 **MySQL 检查+Redis 更新** 封装成一个原子操作：

- 但需要将核心数据也迁移到 Redis。
- 如果核心权威数据还在 MySQL，Redis 只能做缓存，无法完全解决一致性。

------

## **四、总结建议**

如果你的项目**初期流量不大**，完全可以：

1. 先采用 **先写 SQL，再写 Redis**。
2. Redis 写失败只打日志，后面用 **定时修复任务** 做一致性修复。

如果你对实时一致性要求极高（比如计费、强社交核心链路）：

- 考虑使用 **Redis 作为写入口** + **消息队列同步 MySQL** 的架构。









基于你目前的代码，要实现 **Redis 和 MySQL 的定时同步**，可以使用 **先写 SQL，再写 Redis** 的架构，然后通过 **定时修复任务** 确保最终一致性。下面我给你设计一个完整的方案。

------

## **1. 设计原则**

1. **MySQL 是最终真值存储**，所有写操作必须先落库。
2. **Redis 只作为缓存或计数器**，写失败不会影响主流程，只记录错误日志。
3. 定时任务负责检测 Redis 和 MySQL 的数据差异，并进行修复。

------

## **2. 关注/取关流程修改**

你的 `FollowUser` 方法中，先写 MySQL，写成功后再异步写 Redis。

```
func (r *ProfileRepo) FollowUser(ctx context.Context, followerID uint32, followeeID uint32) error {
	// 1. 确认目标用户存在
	user, err := r.GetProfileByUserID(ctx, followerID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	// 2. 先写 MySQL
	follow := bizProfile.FollowTB{FollowerID: followerID, FolloweeID: followeeID}
	if err := r.data.DB().Create(&follow).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return errors.New("already followed")
		}
		return err
	}

	// 3. 更新计数（MySQL）
	r.data.DB().Model(&bizProfile.ProfileTB{}).Where("user_id = ?", followerID).
		Update("follow_count", gorm.Expr("follow_count + 1"))
	r.data.DB().Model(&bizProfile.ProfileTB{}).Where("user_id = ?", followeeID).
		Update("fan_count", gorm.Expr("fan_count + 1"))

	// 4. 写 Redis，如果失败只记录日志
	go func() {
		if err := r.updateFollowCache(ctx, followerID, followeeID); err != nil {
			r.log.Errorf("redis update follow cache failed: %v", err)
			// 记录到修复队列
			_ = r.recordFollowRepairTask(ctx, followerID, followeeID, "follow")
		}
	}()

	return nil
}
```

取关逻辑 `UnFollowUser` 同理，先 MySQL，再 Redis。

------

## **3. Redis 修复记录表**

当 Redis 更新失败时，需要记录一个**修复任务**，后续由定时任务进行修复。

```
CREATE TABLE redis_repair_tasks (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    task_type VARCHAR(50) NOT NULL COMMENT 'follow 或 unfollow',
    follower_id BIGINT NOT NULL,
    followee_id BIGINT NOT NULL,
    retry_count INT DEFAULT 0,
    status VARCHAR(20) DEFAULT 'pending' COMMENT 'pending, done, failed',
    error_msg TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

------

## **4. 写入修复任务**

当 Redis 写失败时，记录一条修复任务。

```
func (r *ProfileRepo) recordFollowRepairTask(ctx context.Context, followerID, followeeID uint32, taskType string) error {
	task := map[string]interface{}{
		"task_type":   taskType,
		"follower_id": followerID,
		"followee_id": followeeID,
		"status":      "pending",
	}
	return r.data.DB().Table("redis_repair_tasks").Create(task).Error
}
```

------

## **5. 定时修复任务**

使用 `cron` 或 Go 的 `time.Ticker` 定时扫描 `redis_repair_tasks` 表中 **pending** 状态的任务，尝试重新同步 Redis。

```
func (r *ProfileRepo) StartRepairScheduler(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute) // 每分钟执行一次
	go func() {
		for {
			select {
			case <-ticker.C:
				r.repairRedisData(ctx)
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

func (r *ProfileRepo) repairRedisData(ctx context.Context) {
	var tasks []map[string]interface{}
	// 一次取100条
	if err := r.data.DB().
		Table("redis_repair_tasks").
		Where("status = ?", "pending").
		Limit(100).Find(&tasks).Error; err != nil {
		r.log.Errorf("query repair tasks failed: %v", err)
		return
	}

	for _, task := range tasks {
		followerID := uint32(task["follower_id"].(int64))
		followeeID := uint32(task["followee_id"].(int64))
		taskType := task["task_type"].(string)

		var err error
		if taskType == "follow" {
			err = r.updateFollowCache(ctx, followerID, followeeID)
		} else {
			err = r.updateUnfollowCache(ctx, followerID, followeeID)
		}

		if err != nil {
			r.log.Errorf("repair task failed: %v", err)
			r.data.DB().Table("redis_repair_tasks").
				Where("id = ?", task["id"]).
				Updates(map[string]interface{}{
					"retry_count": gorm.Expr("retry_count + 1"),
					"error_msg":   err.Error(),
				})
		} else {
			r.data.DB().Table("redis_repair_tasks").
				Where("id = ?", task["id"]).
				Update("status", "done")
		}
	}
}
```

------

## **6. 周期性全量一致性校验**

除了修复队列，建议每天定时做一次**全量一致性对账**，确保 MySQL 和 Redis 一致：

### SQL 统计

```
-- 统计 MySQL 中某用户的关注数
SELECT COUNT(*) FROM follow_tb WHERE follower_id = 123 AND status = 'normal';
```

### Redis 统计

```
SCARD user:following:123
```

Go 代码中：

```
func (r *ProfileRepo) verifyUserFollowConsistency(ctx context.Context, userID uint32) {
	var dbCount int64
	r.data.DB().Model(&bizProfile.FollowTB{}).
		Where("follower_id = ? AND status = ?", userID, "normal").
		Count(&dbCount)

	redisKey := UserRedisKey(UserCachePrefix, "following", userID)
	redisCount, err := r.data.Cache().SCard(ctx, redisKey).Result()
	if err != nil {
		r.log.Errorf("redis count failed for user %d: %v", userID, err)
		return
	}

	if int64(redisCount) != dbCount {
		r.log.Warnf("data mismatch for user %d: mysql=%d redis=%d", userID, dbCount, redisCount)
		// 可以把任务放入修复表
	}
}
```

------

## **7. 总体架构图**

```
用户关注操作
   |
   v
写 MySQL --------> 成功返回给用户
   |
   +--异步写 Redis
         |
         +--写失败 -> 记录 redis_repair_tasks

定时任务
   |
   v
扫描 redis_repair_tasks 表
   |
   +--成功 -> 标记 done
   +--失败 -> retry_count++
```

------

## **8. 最终实现效果**

- 主流程只依赖 MySQL，保证核心业务的强一致性。
- Redis 写失败不影响用户体验，通过**修复任务**最终保证一致性。
- 结合定时全量校验，能够发现并修复漏网数据。
