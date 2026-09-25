-- Enable PostGIS and pgRouting extensions
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS pgrouting;

-- buildings
CREATE TABLE IF NOT EXISTS buildings (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(200) NOT NULL,
    description TEXT,
    affiliation VARCHAR(100)
);

-- nodes
CREATE TABLE IF NOT EXISTS nodes (
    id          SERIAL PRIMARY KEY,
    node_id     VARCHAR(50) NOT NULL UNIQUE,
    geom        GEOMETRY(Point, 4326) NOT NULL,
    node_type   VARCHAR(50) NOT NULL,
    building_id INTEGER,
    floor       INTEGER,
    CONSTRAINT chk_node_id_format CHECK (node_id ~ '^[a-z0-9_]+$'),
    CONSTRAINT chk_node_type CHECK (node_type IN ('entrance', 'road', 'door', 'facility')),
    CONSTRAINT chk_floor_range CHECK (floor IS NULL OR (floor >= -10 AND floor <= 100))
);

CREATE INDEX IF NOT EXISTS idx_nodes_geom ON nodes USING GIST (geom);
CREATE INDEX IF NOT EXISTS idx_nodes_building ON nodes (building_id);
CREATE INDEX IF NOT EXISTS idx_nodes_type ON nodes (node_type);

-- edges
CREATE TABLE IF NOT EXISTS edges (
    id             SERIAL PRIMARY KEY,
    node_id_from   VARCHAR(50) NOT NULL,
    node_id_target VARCHAR(50) NOT NULL,
    distance       FLOAT NOT NULL,
    level          INTEGER NOT NULL,
    has_stairs     BOOLEAN NOT NULL,
    is_accessible  BOOLEAN NOT NULL,
    is_indoor      BOOLEAN NOT NULL,
    CONSTRAINT chk_edge_from_format CHECK (node_id_from ~ '^[a-z0-9_]+$'),
    CONSTRAINT chk_edge_target_format CHECK (node_id_target ~ '^[a-z0-9_]+$'),
    CONSTRAINT chk_edge_no_self_ref CHECK (node_id_from != node_id_target),
    CONSTRAINT chk_distance_positive CHECK (distance >= 0)
);

CREATE INDEX IF NOT EXISTS idx_edges_from ON edges (node_id_from);
CREATE INDEX IF NOT EXISTS idx_edges_target ON edges (node_id_target);
CREATE INDEX IF NOT EXISTS idx_edges_indoor ON edges (is_indoor);

-- rooms
CREATE TABLE IF NOT EXISTS rooms (
    id          SERIAL PRIMARY KEY,
    room_id     VARCHAR(50) NOT NULL UNIQUE,
    building_id INTEGER NOT NULL,
    node_id     VARCHAR(50),
    name        VARCHAR(200),
    description TEXT,
    floor       INTEGER,
    capacity    INTEGER,
    room_type   VARCHAR(50),
    CONSTRAINT chk_room_id_format CHECK (room_id ~ '^[a-z0-9_]+$'),
    CONSTRAINT chk_room_node_id_format CHECK (node_id IS NULL OR node_id ~ '^[a-z0-9_]+$'),
    CONSTRAINT chk_room_capacity CHECK (capacity IS NULL OR capacity >= 0),
    CONSTRAINT chk_room_floor_range CHECK (floor IS NULL OR (floor >= -10 AND floor <= 100))
);

CREATE INDEX IF NOT EXISTS idx_rooms_building ON rooms (building_id);
CREATE INDEX IF NOT EXISTS idx_rooms_node ON rooms (node_id);
CREATE INDEX IF NOT EXISTS idx_rooms_type ON rooms (room_type);

-- photos
CREATE TABLE IF NOT EXISTS photos (
    id            SERIAL PRIMARY KEY,
    building_id   INTEGER NOT NULL,
    url           TEXT NOT NULL,
    caption       TEXT,
    display_order INTEGER NOT NULL,
    CONSTRAINT chk_photo_url CHECK (url ~ '^https?://'),
    CONSTRAINT chk_photo_order CHECK (display_order >= 0)
);

CREATE INDEX IF NOT EXISTS idx_photos_building ON photos (building_id);
CREATE INDEX IF NOT EXISTS idx_photos_order ON photos (building_id, display_order);

-- pgRouting用ビュー
CREATE OR REPLACE VIEW edges_for_routing AS
SELECT
    e.id,
    n1.id AS source,
    n2.id AS target,
    e.distance,
    e.level,
    e.has_stairs,
    e.is_accessible,
    e.is_indoor,
    e.node_id_from,
    e.node_id_target,
    n1.building_id AS building_id_from,
    n2.building_id AS building_id_target
FROM edges e
JOIN nodes n1 ON e.node_id_from = n1.node_id
JOIN nodes n2 ON e.node_id_target = n2.node_id;
