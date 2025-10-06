CREATE TABLE IF NOT EXISTS scans (
    ipv4_addr INET NOT NULL,
    port INT NOT NULL,
    service TEXT NOT NULL,
    resp TEXT NOT NULL,
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (ipv4_addr, port, service)
);
CREATE INDEX IF NOT EXISTS idx_scans_updated_at ON scans (updated_at);
CREATE INDEX IF NOT EXISTS idx_scans_ipv4_addr_port_service ON scans (ipv4_addr, port, service);
CREATE INDEX IF NOT EXISTS idx_scans_updated_at_ipv4_addr_port_service ON scans (updated_at, ipv4_addr, port, service);