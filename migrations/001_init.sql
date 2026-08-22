-- AiTreadings v2.0 数据库迁移
-- 创建 trending_snapshots 表

CREATE TABLE IF NOT EXISTS trending_snapshots (
    id SERIAL PRIMARY KEY,
    project_name VARCHAR(255) NOT NULL,
    category VARCHAR(50),
    stars INT NOT NULL DEFAULT 0,
    forks INT NOT NULL DEFAULT 0,
    delta_stars INT NOT NULL DEFAULT 0,
    snapshot_date DATE NOT NULL,
    raw_metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(project_name, snapshot_date)
);

-- 联合索引：按类别和日期查询
CREATE INDEX IF NOT EXISTS idx_snapshots_category_date
    ON trending_snapshots (category, snapshot_date DESC);

-- 按日期查询索引
CREATE INDEX IF NOT EXISTS idx_snapshots_date
    ON trending_snapshots (snapshot_date DESC);

-- 创建 analysis_cache 表
CREATE TABLE IF NOT EXISTS analysis_cache (
    readme_md5 CHAR(32) PRIMARY KEY,
    llm_response JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL DEFAULT NOW() + INTERVAL '30 days'
);

-- 过期缓存清理索引
CREATE INDEX IF NOT EXISTS idx_cache_expires
    ON analysis_cache (expires_at);
