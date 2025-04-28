CREATE TABLE mtg_import_status (
    id INTEGER PRIMARY KEY,
    last_import TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Add trigger for updated_at
CREATE TRIGGER update_mtg_import_status_updated_at
    BEFORE UPDATE ON mtg_import_status
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column(); 