CREATE TABLE `server_management`
(
    `id`         bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `service_name`   varchar(128) NOT NULL DEFAULT '' COMMENT 'ServiceName',
    `idl_content`   text NOT NULL COMMENT 'IdlContent',
    `server_version`   varchar(128) NOT NULL DEFAULT '' COMMENT 'ServerVersion',
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'create time',
    `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'update time',
    `deleted_at` timestamp NULL DEFAULT NULL COMMENT 'delete time',
    PRIMARY KEY (`id`),
    KEY          `idx_service_name` (`service_name`) COMMENT 'service_name index'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='server management table';