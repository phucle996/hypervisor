ALTER TABLE hypervisor_node_agents
    DROP COLUMN IF EXISTS cert_not_after,
    DROP COLUMN IF EXISTS cert_serial;

ALTER TABLE hypervisor_nodes
    DROP COLUMN IF EXISTS management_ip;
