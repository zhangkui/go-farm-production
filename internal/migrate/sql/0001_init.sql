-- =============================================================================
-- go-farm-production  schema migration 0001_init
-- -----------------------------------------------------------------------------
-- All DECIMAL columns use (12,2) for money/area/weight and (8,2) for hours.
-- Status enums are TINYINT with numeric codes (see domain/enums.go).
-- Foreign keys use RESTRICT to prevent orphaning; referential integrity is
-- enforced at the database layer as well as in application code.
-- =============================================================================

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- -----------------------------------------------------------------------------
-- Auth & RBAC
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS users (
    id            BIGINT       NOT NULL AUTO_INCREMENT,
    username      VARCHAR(64)  NOT NULL,
    email         VARCHAR(128) NOT NULL,
    full_name     VARCHAR(128) NOT NULL DEFAULT '',
    password_hash VARCHAR(255) NOT NULL,
    status        TINYINT      NOT NULL DEFAULT 1 COMMENT '1=active 0=inactive',
    created_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_users_username (username),
    UNIQUE KEY uk_users_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS roles (
    id          BIGINT       NOT NULL AUTO_INCREMENT,
    code        VARCHAR(64)  NOT NULL,
    name        VARCHAR(128) NOT NULL,
    description VARCHAR(255) NOT NULL DEFAULT '',
    created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_roles_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS permissions (
    id          BIGINT       NOT NULL AUTO_INCREMENT,
    code        VARCHAR(128) NOT NULL,
    name        VARCHAR(128) NOT NULL,
    resource    VARCHAR(64)  NOT NULL,
    action      VARCHAR(32)  NOT NULL,
    description VARCHAR(255) NOT NULL DEFAULT '',
    created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_permissions_code (code),
    KEY idx_permissions_resource (resource)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS user_roles (
    user_id BIGINT NOT NULL,
    role_id BIGINT NOT NULL,
    PRIMARY KEY (user_id, role_id),
    CONSTRAINT fk_user_roles_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT fk_user_roles_role FOREIGN KEY (role_id) REFERENCES roles (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id       BIGINT NOT NULL,
    permission_id BIGINT NOT NULL,
    PRIMARY KEY (role_id, permission_id),
    CONSTRAINT fk_role_perms_role FOREIGN KEY (role_id) REFERENCES roles (id) ON DELETE CASCADE,
    CONSTRAINT fk_role_perms_perm FOREIGN KEY (permission_id) REFERENCES permissions (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id         BIGINT       NOT NULL AUTO_INCREMENT,
    token      VARCHAR(128) NOT NULL,
    user_id    BIGINT       NOT NULL,
    expires_at DATETIME     NOT NULL,
    revoked    TINYINT      NOT NULL DEFAULT 0,
    created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_refresh_token (token),
    KEY idx_refresh_user (user_id),
    CONSTRAINT fk_refresh_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- Farms & fields
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS farms (
    id          BIGINT       NOT NULL AUTO_INCREMENT,
    name        VARCHAR(128) NOT NULL,
    location    VARCHAR(255) NOT NULL DEFAULT '',
    total_area  DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    description TEXT,
    status      TINYINT      NOT NULL DEFAULT 1,
    created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS fields (
    id              BIGINT       NOT NULL AUTO_INCREMENT,
    farm_id         BIGINT       NOT NULL,
    code            VARCHAR(64)  NOT NULL,
    name            VARCHAR(128) NOT NULL,
    area            DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    soil_type       VARCHAR(64)  NOT NULL DEFAULT '',
    irrigation_zone VARCHAR(64)  NOT NULL DEFAULT '',
    status          TINYINT      NOT NULL DEFAULT 1 COMMENT '1=available 2=planting 3=fallow 0=inactive',
    remark          TEXT,
    created_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_field_code_in_farm (farm_id, code),
    KEY idx_fields_irrigation (irrigation_zone),
    CONSTRAINT fk_fields_farm FOREIGN KEY (farm_id) REFERENCES farms (id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- Crop varieties & seasons
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS crop_varieties (
    id           BIGINT       NOT NULL AUTO_INCREMENT,
    code         VARCHAR(64)  NOT NULL,
    name         VARCHAR(128) NOT NULL,
    category     VARCHAR(64)  NOT NULL DEFAULT '',
    growth_cycle INT          NOT NULL DEFAULT 0,
    description  TEXT,
    status       TINYINT      NOT NULL DEFAULT 1,
    created_at   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_variety_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS seasons (
    id         BIGINT       NOT NULL AUTO_INCREMENT,
    code       VARCHAR(64)  NOT NULL,
    name       VARCHAR(128) NOT NULL,
    start_date DATE         NOT NULL,
    end_date   DATE         NOT NULL,
    status     TINYINT      NOT NULL DEFAULT 1 COMMENT '1=planned 2=active 3=completed 0=cancelled',
    created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_season_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- Planting plans
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS planting_plans (
    id                  BIGINT       NOT NULL AUTO_INCREMENT,
    field_id            BIGINT       NOT NULL,
    crop_variety_id     BIGINT       NOT NULL,
    season_id           BIGINT       NOT NULL,
    planned_area        DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    planned_sow_date    DATE         NOT NULL,
    planned_harvest_date DATE,
    actual_sow_date     DATE,
    actual_harvest_date DATE,
    status              TINYINT      NOT NULL DEFAULT 1 COMMENT '1=planned 2=planted 3=growing 4=harvested 5=completed 0=cancelled',
    remark              TEXT,
    created_at          DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_plans_field (field_id),
    KEY idx_plans_season (season_id),
    KEY idx_plans_status (status),
    CONSTRAINT fk_plans_field FOREIGN KEY (field_id) REFERENCES fields (id) ON DELETE RESTRICT,
    CONSTRAINT fk_plans_variety FOREIGN KEY (crop_variety_id) REFERENCES crop_varieties (id) ON DELETE RESTRICT,
    CONSTRAINT fk_plans_season FOREIGN KEY (season_id) REFERENCES seasons (id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- Farm tasks
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS farm_tasks (
    id              BIGINT       NOT NULL AUTO_INCREMENT,
    planting_plan_id BIGINT      NOT NULL,
    task_type       TINYINT      NOT NULL DEFAULT 5 COMMENT '1=fertilizing 2=irrigation 3=pest_control 4=harvesting 5=other',
    title           VARCHAR(128) NOT NULL,
    description     TEXT,
    planned_date    DATE         NOT NULL,
    completed_date  DATE,
    status          TINYINT      NOT NULL DEFAULT 1 COMMENT '1=scheduled 2=in_progress 3=completed 0=cancelled',
    labour_hours    DECIMAL(8,2) NOT NULL DEFAULT 0.00,
    equipment_cost  DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    assignee_id     BIGINT,
    remark          TEXT,
    created_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_tasks_plan (planting_plan_id),
    KEY idx_tasks_status (status),
    KEY idx_tasks_assignee (assignee_id),
    CONSTRAINT fk_tasks_plan FOREIGN KEY (planting_plan_id) REFERENCES planting_plans (id) ON DELETE RESTRICT,
    CONSTRAINT fk_tasks_assignee FOREIGN KEY (assignee_id) REFERENCES users (id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- Inputs: materials, batches, allocations
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS materials (
    id          BIGINT       NOT NULL AUTO_INCREMENT,
    code        VARCHAR(64)  NOT NULL,
    name        VARCHAR(128) NOT NULL,
    category    TINYINT      NOT NULL DEFAULT 5 COMMENT '1=fertilizer 2=pesticide 3=seed 4=fuel 5=other',
    unit        VARCHAR(16)  NOT NULL DEFAULT 'kg',
    unit_price  DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    description TEXT,
    status      TINYINT      NOT NULL DEFAULT 1,
    created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_material_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS input_batches (
    id             BIGINT       NOT NULL AUTO_INCREMENT,
    material_id    BIGINT       NOT NULL,
    batch_no       VARCHAR(64)  NOT NULL,
    quantity       DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    remaining_qty  DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    purchase_date  DATE,
    expiry_date    DATE,
    purchase_price DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    supplier       VARCHAR(128) NOT NULL DEFAULT '',
    status         TINYINT      NOT NULL DEFAULT 1 COMMENT '1=active 2=depleted 3=expired 0=void',
    created_at     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_batch_no_in_material (material_id, batch_no),
    KEY idx_batches_expiry (expiry_date),
    KEY idx_batches_status (status),
    CONSTRAINT fk_batches_material FOREIGN KEY (material_id) REFERENCES materials (id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS input_allocations (
    id           BIGINT       NOT NULL AUTO_INCREMENT,
    batch_id     BIGINT       NOT NULL,
    material_id  BIGINT       NOT NULL,
    task_id      BIGINT,
    quantity     DECIMAL(12,2) NOT NULL,
    type         TINYINT      NOT NULL COMMENT '1=allocate 0=return 2=waste',
    remark       TEXT,
    operator_id  BIGINT       NOT NULL,
    created_at   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_alloc_batch (batch_id),
    KEY idx_alloc_returnable (batch_id, task_id, type),
    KEY idx_alloc_task (task_id),
    KEY idx_alloc_material (material_id),
    CONSTRAINT fk_alloc_batch FOREIGN KEY (batch_id) REFERENCES input_batches (id) ON DELETE RESTRICT,
    CONSTRAINT fk_alloc_material FOREIGN KEY (material_id) REFERENCES materials (id) ON DELETE RESTRICT,
    CONSTRAINT fk_alloc_task FOREIGN KEY (task_id) REFERENCES farm_tasks (id) ON DELETE SET NULL,
    CONSTRAINT fk_alloc_operator FOREIGN KEY (operator_id) REFERENCES users (id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- Harvests
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS harvests (
    id              BIGINT       NOT NULL AUTO_INCREMENT,
    planting_plan_id BIGINT      NOT NULL,
    harvest_date    DATE         NOT NULL,
    total_weight    DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    grade           VARCHAR(32)  NOT NULL DEFAULT '',
    remark          TEXT,
    approved        TINYINT      NOT NULL DEFAULT 0,
    created_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_harvest_plan (planting_plan_id),
    CONSTRAINT fk_harvest_plan FOREIGN KEY (planting_plan_id) REFERENCES planting_plans (id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS harvest_details (
    id         BIGINT       NOT NULL AUTO_INCREMENT,
    harvest_id BIGINT       NOT NULL,
    field_id   BIGINT       NOT NULL,
    weight     DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    grade      VARCHAR(32)  NOT NULL DEFAULT '',
    remark     TEXT,
    created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_hdetail_harvest (harvest_id),
    KEY idx_hdetail_field (field_id),
    CONSTRAINT fk_hdetail_harvest FOREIGN KEY (harvest_id) REFERENCES harvests (id) ON DELETE CASCADE,
    CONSTRAINT fk_hdetail_field FOREIGN KEY (field_id) REFERENCES fields (id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- Produce inventory & cost analysis
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS produce_inventory (
    id               BIGINT       NOT NULL AUTO_INCREMENT,
    crop_variety_id  BIGINT       NOT NULL,
    harvest_id       BIGINT       NOT NULL,
    quantity         DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    grade            VARCHAR(32)  NOT NULL DEFAULT '',
    unit             VARCHAR(16)  NOT NULL DEFAULT 'kg',
    storage_location VARCHAR(128) NOT NULL DEFAULT '',
    status           TINYINT      NOT NULL DEFAULT 1 COMMENT '1=in_stock 2=sold 3=processed 0=wasted',
    created_at       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_produce_variety (crop_variety_id),
    KEY idx_produce_status (status),
    CONSTRAINT fk_produce_variety FOREIGN KEY (crop_variety_id) REFERENCES crop_varieties (id) ON DELETE RESTRICT,
    CONSTRAINT fk_produce_harvest FOREIGN KEY (harvest_id) REFERENCES harvests (id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS cost_analyses (
    id               BIGINT       NOT NULL AUTO_INCREMENT,
    planting_plan_id BIGINT       NOT NULL,
    analysis_date    DATE         NOT NULL,
    labour_cost      DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    input_cost       DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    equipment_cost   DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    other_cost       DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    total_cost       DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    total_yield      DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    unit_cost        DECIMAL(16,4) NOT NULL DEFAULT 0.00,
    remark           TEXT,
    created_at       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_cost_plan (planting_plan_id),
    CONSTRAINT fk_cost_plan FOREIGN KEY (planting_plan_id) REFERENCES planting_plans (id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- Audit log
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS audit_logs (
    id            BIGINT       NOT NULL AUTO_INCREMENT,
    user_id       BIGINT,
    username      VARCHAR(64)  NOT NULL DEFAULT '',
    action        VARCHAR(32)  NOT NULL,
    resource_type VARCHAR(64)  NOT NULL,
    resource_id   VARCHAR(64)  NOT NULL DEFAULT '',
    details       JSON,
    ip_address    VARCHAR(64)  NOT NULL DEFAULT '',
    user_agent    VARCHAR(255) NOT NULL DEFAULT '',
    created_at    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_audit_user (user_id),
    KEY idx_audit_resource (resource_type, resource_id),
    KEY idx_audit_action (action),
    KEY idx_audit_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- -----------------------------------------------------------------------------
-- Migration bookkeeping
-- -----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS schema_migrations (
    version    BIGINT      NOT NULL,
    applied_at DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (version)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO schema_migrations (version) VALUES (1)
    ON DUPLICATE KEY UPDATE version = version;

SET FOREIGN_KEY_CHECKS = 1;
