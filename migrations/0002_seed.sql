-- =============================================================================
-- go-farm-production  migration 0002_seed
-- Seeds permissions, roles, default admin user, and role-permission grants.
-- Run once after 0001_init. Idempotent on the permission/role lookups via code.
-- =============================================================================

SET NAMES utf8mb4;

-- Permissions ----------------------------------------------------------------
INSERT INTO permissions (code, name, resource, action, description) VALUES
    ('user:manage',           '用户管理',           'user',              'manage', '创建/编辑/删除/重置用户'),
    ('user:view',             '用户查看',           'user',              'view',   '查看用户列表与详情'),
    ('role:manage',           '角色权限管理',       'role',              'manage', '管理角色与权限分配'),
    ('role:view',             '角色查看',           'role',              'view',   '查看角色与权限列表'),
    ('farm:manage',           '农场管理',           'farm',              'manage', '创建/编辑/删除农场'),
    ('farm:view',             '农场查看',           'farm',              'view',   '查看农场列表与详情'),
    ('field:manage',          '地块管理',           'field',             'manage', '创建/编辑/删除地块'),
    ('field:view',            '地块查看',           'field',             'view',   '查看地块列表与详情'),
    ('crop_variety:manage',   '作物品种管理',       'crop_variety',      'manage', '管理作物品种字典'),
    ('crop_variety:view',     '作物品种查看',       'crop_variety',      'view',   '查看作物品种列表'),
    ('season:manage',         '种植季管理',         'season',            'manage', '创建/编辑/删除季次'),
    ('season:view',           '种植季查看',         'season',            'view',   '查看季次列表与详情'),
    ('planting_plan:manage',  '种植计划管理',       'planting_plan',     'manage', '创建/编辑种植计划'),
    ('planting_plan:approve', '种植计划审批',       'planting_plan',     'approve','审批种植计划'),
    ('planting_plan:view',    '种植计划查看',       'planting_plan',     'view',   '查看计划列表与详情'),
    ('farm_task:manage',      '农事任务管理',       'farm_task',         'manage', '创建/分配农事任务'),
    ('farm_task:execute',     '农事任务执行',       'farm_task',         'execute','执行/记录任务'),
    ('farm_task:view',        '农事任务查看',       'farm_task',         'view',   '查看任务列表与详情'),
    ('material:manage',      '投入品物料管理',     'material',          'manage', '创建/编辑物料'),
    ('batch:manage',          '批次管理',           'batch',             'manage', '入库/批次管理'),
    ('allocation:manage',     '投入品领用管理',     'allocation',        'manage', '领用/退回/损耗'),
    ('inventory:view',        '库存查看',           'inventory',         'view',   '查看库存与预警'),
    ('harvest:manage',        '采收管理',           'harvest',           'manage', '创建/编辑采收记录'),
    ('harvest:approve',       '采收审核',           'harvest',           'approve','审核采收记录'),
    ('harvest:view',          '采收查看',           'harvest',           'view',   '查看采收列表与详情'),
    ('produce_inventory:manage','农产品库存管理',   'produce_inventory', 'manage', '管理农产品库存'),
    ('produce_inventory:view',  '农产品库存查看',   'produce_inventory', 'view',   '查看农产品库存'),
    ('cost_analysis:view',    '成本分析查看',       'cost_analysis',     'view',   '查看成本报表'),
    ('cost_analysis:export',  '成本报表导出',       'cost_analysis',     'export', '导出成本报表'),
    ('audit_log:view',        '审计日志查看',       'audit_log',         'view',   '查看所有审计日志')
ON DUPLICATE KEY UPDATE name = VALUES(name);

-- Roles ----------------------------------------------------------------------
INSERT INTO roles (code, name, description) VALUES
    ('admin',                '系统管理员', '系统最高权限，管理用户、角色、权限和系统配置'),
    ('farm_manager',         '业务管理员', '农场经营管理，管理地块、种植计划、审批任务'),
    ('production_supervisor','生产主管',   '执行层面管理，分配农事任务、管理投入品'),
    ('operator',             '操作人员',   '日常农事操作记录、投入品领用'),
    ('auditor',              '审计人员',   '只读访问所有业务数据'),
    ('read_only',            '只读用户',   '仅查看报表和仪表盘')
ON DUPLICATE KEY UPDATE name = VALUES(name);

-- Grants ---------------------------------------------------------------------
-- Helper: grant all permissions except role:user management where appropriate.
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r CROSS JOIN permissions p
WHERE r.code = 'admin'
ON DUPLICATE KEY UPDATE role_id = role_id;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r JOIN permissions p
WHERE r.code = 'farm_manager'
  AND p.code IN (
    'farm:manage','farm:view','field:manage','field:view',
    'crop_variety:manage','crop_variety:view','season:manage','season:view',
    'planting_plan:manage','planting_plan:approve','planting_plan:view',
    'farm_task:manage','farm_task:execute','farm_task:view',
    'material:manage','batch:manage','allocation:manage','inventory:view',
    'harvest:manage','harvest:approve','harvest:view',
    'produce_inventory:manage','produce_inventory:view',
    'cost_analysis:view','cost_analysis:export'
  )
ON DUPLICATE KEY UPDATE role_id = role_id;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r JOIN permissions p
WHERE r.code = 'production_supervisor'
  AND p.code IN (
    'farm:view','field:manage','field:view',
    'crop_variety:manage','crop_variety:view','season:manage','season:view',
    'planting_plan:manage','planting_plan:view',
    'farm_task:manage','farm_task:execute','farm_task:view',
    'material:manage','batch:manage','allocation:manage','inventory:view',
    'harvest:manage','harvest:view','produce_inventory:view','cost_analysis:view'
  )
ON DUPLICATE KEY UPDATE role_id = role_id;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r JOIN permissions p
WHERE r.code = 'operator'
  AND p.code IN (
    'farm:view','field:view','crop_variety:view','season:view','planting_plan:view',
    'farm_task:execute','farm_task:view',
    'batch:manage','allocation:manage','inventory:view',
    'harvest:manage','harvest:view','produce_inventory:view','cost_analysis:view'
  )
ON DUPLICATE KEY UPDATE role_id = role_id;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r JOIN permissions p
WHERE r.code = 'auditor'
  AND p.code IN (
    'farm:view','field:view','crop_variety:view','season:view','planting_plan:view',
    'farm_task:view','material:manage','batch:manage','inventory:view',
    'harvest:view','produce_inventory:view','cost_analysis:view','audit_log:view'
  )
ON DUPLICATE KEY UPDATE role_id = role_id;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r JOIN permissions p
WHERE r.code = 'read_only'
  AND p.code IN (
    'farm:view','field:view','crop_variety:view','season:view','planting_plan:view',
    'farm_task:view','inventory:view','harvest:view',
    'produce_inventory:view','cost_analysis:view'
  )
ON DUPLICATE KEY UPDATE role_id = role_id;

INSERT INTO schema_migrations (version) VALUES (2)
ON DUPLICATE KEY UPDATE version = version;
