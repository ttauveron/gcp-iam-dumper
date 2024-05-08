DROP TABLE IF EXISTS hierarchy;
CREATE TABLE IF NOT EXISTS hierarchy
(
    id        TEXT PRIMARY KEY,
    name      TEXT NOT NULL,
    type      TEXT NOT NULL CHECK (type IN ('project', 'folder', 'organization')),
    parent_id TEXT,
    FOREIGN KEY (parent_id) REFERENCES hierarchy (id)
);

DROP TABLE IF EXISTS principal;
CREATE TABLE IF NOT EXISTS principal
(
    id   TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    type TEXT NOT NULL CHECK (type IN ('user', 'group', 'serviceAccount'))
);

DROP TABLE IF EXISTS principal_hierarchy;
CREATE TABLE IF NOT EXISTS principal_hierarchy
(
    parent_id TEXT NOT NULL,
    child_id  TEXT NOT NULL,
    PRIMARY KEY (parent_id, child_id),
    FOREIGN KEY (parent_id) REFERENCES principal (id),
    FOREIGN KEY (child_id) REFERENCES principal (id)
);

DROP TABLE IF EXISTS resource_role_principal;
CREATE TABLE IF NOT EXISTS resource_role_principal
(
    resource_id    TEXT NOT NULL,
    principal_name TEXT NOT NULL,
    role_id        TEXT NOT NULL,
    conditional    TEXT,
    asset_type     TEXT NOT NULL,
    hierarchy_id   TEXT NOT NULL,
    PRIMARY KEY (resource_id, principal_name, role_id, conditional, hierarchy_id, asset_type),
    FOREIGN KEY (principal_name) REFERENCES principal (name),
    FOREIGN KEY (hierarchy_id) REFERENCES hierarchy (id),
    FOREIGN KEY (role_id) REFERENCES role (id)
);

DROP TABLE IF EXISTS role;
CREATE TABLE IF NOT EXISTS role
(
    id    TEXT NOT NULL,
    title TEXT NOT NULL,
    PRIMARY KEY (id)
);

DROP TABLE IF EXISTS role_permission;
CREATE TABLE IF NOT EXISTS role_permission
(
    role_id    TEXT NOT NULL,
    permission_id TEXT NOT NULL,
    PRIMARY KEY (role_id, permission_id),
    FOREIGN KEY (role_id) REFERENCES role(id)
);