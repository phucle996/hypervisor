-- Rollback: remove GPU hardware columns from hypervisor_nodes
ALTER TABLE hypervisor.hypervisor_nodes
    DROP COLUMN IF EXISTS gpu_model,
    DROP COLUMN IF EXISTS gpu_count;
