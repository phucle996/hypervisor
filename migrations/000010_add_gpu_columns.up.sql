-- Add GPU hardware columns to hypervisor_nodes
ALTER TABLE hypervisor.hypervisor_nodes
    ADD COLUMN IF NOT EXISTS gpu_model TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS gpu_count INT NOT NULL DEFAULT 0;

ALTER TABLE hypervisor.hypervisor_nodes
    DROP CONSTRAINT IF EXISTS chk_hypervisor_nodes_gpu_count;
ALTER TABLE hypervisor.hypervisor_nodes
    ADD CONSTRAINT chk_hypervisor_nodes_gpu_count CHECK (gpu_count >= 0);
