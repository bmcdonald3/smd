/*
 * MIT License
 *
 * (C) Copyright [2022] Hewlett Packard Enterprise Development LP
 *
 * Permission is hereby granted, free of charge, to any person obtaining a
 * copy of this software and associated documentation files (the "Software"),
 * to deal in the Software without restriction, including without limitation
 * the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the
 * Software is furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included
 * in all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 * THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 * OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 * ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 * OTHER DEALINGS IN THE SOFTWARE.
 */
-- Adds back v1 locking database tables and columns

BEGIN;

ALTER TABLE reservations
    ADD COLUMN IF NOT EXISTS v1_lock_id UUID;

CREATE TABLE IF NOT EXISTS component_locks (
    "id"       UUID PRIMARY KEY,
    "created"  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "reason"   VARCHAR(255) NOT NULL, 
    "owner"    VARCHAR(255) NOT NULL,
    "lifetime" INT NOT NULL
);

CREATE TABLE IF NOT EXISTS component_lock_members (
    "component_id"    VARCHAR(63) PRIMARY KEY,
    "lock_id"         UUID NOT NULL,
    FOREIGN KEY ("component_id") REFERENCES components ("id") ON DELETE CASCADE,
    FOREIGN KEY ("lock_id") REFERENCES component_locks ("id") ON DELETE CASCADE
);

-- Decrease the schema version
INSERT INTO system VALUES(0, 19, '{}'::JSON)
    ON CONFLICT(id) DO UPDATE SET schema_version=19;

COMMIT;