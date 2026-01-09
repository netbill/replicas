-- +migrate Up
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE organization_status AS ENUM (
    'active',
    'inactive',
    'suspended'
);

CREATE TABLE organizations (
    id        UUID                  PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    status    organization_status   NOT NULL DEFAULT 'active',
    verified  BOOLEAN               NOT NULL DEFAULT FALSE,
    name      VARCHAR(255)          NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE organization_members (
    id              UUID PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    account_id      UUID NOT NULL REFERENCES profiles(account_id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    position        VARCHAR(255),
    label           VARCHAR(128),

    created_at TIMESTAMPTZ NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT (now() at time zone 'utc'),

    UNIQUE(account_id, organization_id)
);

CREATE TYPE organization_invite_status AS ENUM (
    'sent',
    'declined',
    'accepted'
);

CREATE TABLE organization_invites (
    id              UUID PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    account_id      UUID NOT NULL REFERENCES profiles(account_id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    status          organization_invite_status NOT NULL DEFAULT 'sent',

    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT (now() at time zone 'utc')
);

CREATE TABLE organization_roles (
    id              UUID    PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    organization_id UUID    NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    head            BOOLEAN NOT NULL DEFAULT false,
    rank            INT     NOT NULL DEFAULT 0 CHECK (rank >= 0),
    name            TEXT    NOT NULL,
    color           TEXT    NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (organization_id, name)
);

CREATE UNIQUE INDEX roles_one_head_per_organization
    ON organization_roles (organization_id)
    WHERE head = true;

CREATE TABLE organization_member_roles (
    member_id UUID NOT NULL REFERENCES organization_members(id) ON DELETE CASCADE,
    role_id   UUID NOT NULL REFERENCES organization_roles (id) ON DELETE CASCADE,

    PRIMARY KEY (member_id, role_id)
);

-- permissions dictionary
CREATE TABLE organization_role_permissions (
    id   UUID          PRIMARY KEY,
    code VARCHAR(255)  UNIQUE NOT NULL,
);

INSERT INTO organization_role_permissions (id, code, description) VALUES
    (uuid_generate_v4(), 'organization.manage'),
    (uuid_generate_v4(), 'invites.manage'),
    (uuid_generate_v4(), 'members.manage'),
    (uuid_generate_v4(), 'roles.manage');

-- role ↔ permission links
CREATE TABLE organization_role_permission_links (
    role_id       UUID NOT NULL REFERENCES organization_roles (id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES organization_role_permissions (id) ON DELETE CASCADE,

    PRIMARY KEY (role_id, permission_id)
);

-- 1) if role.head=true -> add all permissions to role
-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION ensure_head_role_permissions()
RETURNS trigger AS $$
BEGIN
    IF NEW.head = true THEN
        INSERT INTO organization_role_permission_links (role_id, permission_id)
        SELECT NEW.id, p.id
        FROM organization_role_permissions p
        ON CONFLICT DO NOTHING;
    END IF;

    RETURN NEW;
END
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_roles_ensure_head_perms_ins ON roles;
CREATE TRIGGER trg_roles_ensure_head_perms_ins
AFTER INSERT ON organization_roles
FOR EACH ROW
EXECUTE FUNCTION ensure_head_role_permissions();

DROP TRIGGER IF EXISTS trg_roles_ensure_head_perms_upd ON roles;
CREATE TRIGGER trg_roles_ensure_head_perms_upd
AFTER UPDATE OF head ON organization_roles
FOR EACH ROW
EXECUTE FUNCTION ensure_head_role_permissions();

CREATE OR REPLACE FUNCTION grant_new_permission_to_head_roles()
RETURNS trigger AS $$
BEGIN
    INSERT INTO organization_role_permission_links (role_id, permission_id)
    SELECT r.id, NEW.id
    FROM organization_roles r
    WHERE r.head = true
    ON CONFLICT DO NOTHING;

RETURN NEW;
END
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_role_permissions_grant_to_head_roles ON role_permissions;
CREATE TRIGGER trg_role_permissions_grant_to_head_roles
AFTER INSERT ON organization_role_permissions
FOR EACH ROW
EXECUTE FUNCTION grant_new_permission_to_head_roles();
-- +migrate StatementEnd

-- 3) ban delete permissions from head-roles
-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION prevent_delete_head_role_permissions()
RETURNS trigger AS $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM organization_roles r
        WHERE r.id = OLD.role_id
        AND r.head = true
    ) THEN
        RAISE EXCEPTION 'cannot delete permissions from head role %', OLD.role_id
            USING ERRCODE = '23514';
        END IF;

    RETURN OLD;
END
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_role_permission_links_prevent_delete_head ON role_permission_links;
CREATE TRIGGER trg_role_permission_links_prevent_delete_head
BEFORE DELETE ON organization_role_permission_links
FOR EACH ROW
EXECUTE FUNCTION prevent_delete_head_role_permissions();
-- +migrate StatementEnd

-- 4) ban change of organization_id for roles
-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION prevent_role_organization_change()
RETURNS trigger AS $$
BEGIN
    IF NEW.organization_id <> OLD.organization_id THEN
        RAISE EXCEPTION 'cannot change organization_id for role %', OLD.id
            USING ERRCODE = '23514';
    END IF;

    RETURN NEW;
END
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_roles_prevent_organization_change ON roles;
CREATE TRIGGER trg_roles_prevent_organization_change
BEFORE UPDATE OF organization_id ON organization_roles
FOR EACH ROW
EXECUTE FUNCTION prevent_role_organization_change();
-- +migrate StatementEnd

-- 5) ban delete head-roles
-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION prevent_delete_head_role()
RETURNS trigger AS $$
BEGIN
    IF OLD.head = true THEN
        RAISE EXCEPTION 'cannot delete head role %', OLD.id
            USING ERRCODE = '23514';
    END IF;

    RETURN OLD;
END
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_roles_prevent_delete_head ON roles;
CREATE TRIGGER trg_roles_prevent_delete_head
BEFORE DELETE ON organization_roles
FOR EACH ROW
EXECUTE FUNCTION prevent_delete_head_role();
-- +migrate StatementEnd

-- 6) ban remove head-role from members
-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION prevent_remove_head_role_from_member()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM organization_roles r
        WHERE r.id = OLD.role_id
        AND r.head = true
    ) THEN
        RAISE EXCEPTION 'cannot remove head role % from member %',
        OLD.role_id, OLD.member_id
        USING ERRCODE = '23514';
    END IF;

    RETURN OLD;
END;
$$;
-- +migrate StatementEnd

DROP TRIGGER IF EXISTS trg_member_roles_prevent_delete_head_role ON organization_member_roles;
CREATE TRIGGER trg_member_roles_prevent_delete_head_role
BEFORE DELETE ON organization_member_roles
FOR EACH ROW
EXECUTE FUNCTION prevent_remove_head_role_from_member();

-- +migrate Down
DROP TRIGGER IF EXISTS trg_member_roles_prevent_delete_head_role ON organization_member_roles;
DROP FUNCTION IF EXISTS prevent_remove_head_role_from_member();

DROP TRIGGER IF EXISTS trg_roles_prevent_delete_head ON organization_roles;
DROP FUNCTION IF EXISTS prevent_delete_head_role();

DROP TRIGGER IF EXISTS trg_roles_prevent_organization_change ON organization_roles;
DROP FUNCTION IF EXISTS prevent_role_organization_change();

DROP TRIGGER IF EXISTS trg_role_permission_links_prevent_delete_head ON organization_role_permission_links;
DROP FUNCTION IF EXISTS prevent_delete_head_role_permissions();

DROP TRIGGER IF EXISTS trg_role_permissions_grant_to_head_roles ON organization_role_permissions;
DROP FUNCTION IF EXISTS grant_new_permission_to_head_roles();

DROP TRIGGER IF EXISTS trg_roles_ensure_head_perms_upd ON organization_roles;
DROP TRIGGER IF EXISTS trg_roles_ensure_head_perms_ins ON organization_roles;
DROP FUNCTION IF EXISTS ensure_head_role_permissions();

DROP TABLE IF EXISTS organization_role_permission_links CASCADE;
DROP TABLE IF EXISTS organization_role_permissions CASCADE;
DROP TABLE IF EXISTS organization_member_roles CASCADE;
DROP TABLE IF EXISTS organization_roles CASCADE;
DROP TABLE IF EXISTS organization_members CASCADE;
DROP TABLE IF EXISTS organization_invites CASCADE;
DROP TABLE IF EXISTS organizations CASCADE;
DROP TABLE IF EXISTS profiles CASCADE;

DROP TYPE IF EXISTS organization_status;
DROP TYPE IF EXISTS organization_invite_status;