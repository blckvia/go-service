CREATE TABLE projects (
                          id SERIAL PRIMARY KEY,
                          name VARCHAR(255) NOT NULL,
                          created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO projects (name) VALUES ('Первая запись');


CREATE TABLE goods (
                       id SERIAL PRIMARY KEY,
                       project_id INT NOT NULL,
                       name VARCHAR(255) NOT NULL,
                       description TEXT,
                       priority INT,
                       removed BOOLEAN NOT NULL DEFAULT FALSE,
                       created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                       CONSTRAINT fk_goods_project
                           FOREIGN KEY (project_id)
                               REFERENCES projects (id)
);

CREATE INDEX idx_goods_id ON goods (id);
CREATE INDEX idx_goods_project_id ON goods (project_id);

-- Function to set the priority dynamically
CREATE OR REPLACE FUNCTION set_goods_priority()
RETURNS TRIGGER AS $$
BEGIN
    -- Select the max priority from goods and add 1, if no goods set to 1
    NEW.priority := COALESCE((SELECT MAX(priority) FROM goods) + 1, 1);
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to use the set_goods_priority function before insert
CREATE TRIGGER trigger_set_goods_priority
    BEFORE INSERT ON goods
    FOR EACH ROW
    EXECUTE FUNCTION set_goods_priority();


