-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create a function to get current user ID from session
CREATE OR REPLACE FUNCTION get_current_user_id()
RETURNS INTEGER AS $$
BEGIN
    RETURN COALESCE(current_setting('app.current_user_id', TRUE)::INTEGER, -1);
END;
$$ LANGUAGE plpgsql;

-- Create audit columns type
CREATE TYPE audit_fields AS (
    created_at TIMESTAMPTZ,
    modified_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_by INTEGER,
    modified_by INTEGER,
    deleted_by INTEGER
);

-- Create a function to handle audit fields
CREATE OR REPLACE FUNCTION update_audit_fields()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'INSERT') THEN
        -- Set UUID if not provided
        IF NEW.id IS NULL THEN
            NEW.id = uuid_generate_v4();
        END IF;
        NEW.created_at = NOW();
        NEW.created_by = get_current_user_id();
        RETURN NEW;
    ELSIF (TG_OP = 'UPDATE') THEN
        NEW.modified_at = NOW();
        NEW.modified_by = get_current_user_id();
        RETURN NEW;
    ELSIF (TG_OP = 'DELETE') THEN
        IF EXISTS(SELECT 1 FROM information_schema.columns 
                 WHERE table_name=TG_TABLE_NAME::text 
                 AND column_name='deleted_at') THEN
            NEW.deleted_at = NOW();
            NEW.deleted_by = get_current_user_id();
            RETURN NEW;
        ELSE
            RETURN OLD;
        END IF;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Example of how to create a table with UUID
CREATE TABLE example_table (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    -- Your table-specific columns here
    created_at TIMESTAMPTZ NOT NULL,
    modified_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_by INTEGER NOT NULL,
    modified_by INTEGER,
    deleted_by INTEGER
);

-- Create triggers for the table
CREATE TRIGGER example_table_audit
    BEFORE INSERT OR UPDATE ON example_table
    FOR EACH ROW
    EXECUTE FUNCTION update_audit_fields();