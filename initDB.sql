create database solana_trade_bot;

-- solana_trade_bot.positions definition

CREATE TABLE `positions` (
     `id` int NOT NULL AUTO_INCREMENT,
     `token` varchar(100) DEFAULT NULL COMMENT '代币',
     `symbol` varchar(100) DEFAULT NULL,
     `amount` bigint DEFAULT NULL COMMENT '数量',
     `cost_price` decimal(32,20) DEFAULT NULL COMMENT '买入均价',
     `current_price` decimal(32,20) DEFAULT NULL COMMENT '现价',
     `pnl` decimal(10,2) DEFAULT '0.00' COMMENT '涨幅',
     `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
     `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
     `deleted_at` timestamp NULL DEFAULT NULL,
     `decimals` int DEFAULT NULL,
     PRIMARY KEY (`id`),
     KEY `token` (`token`)
) ENGINE=InnoDB;

-- solana_trade_bot.swap_records definition

CREATE TABLE `swap_records` (
    `id` int NOT NULL AUTO_INCREMENT,
    `signature` varchar(100) DEFAULT NULL,
    `buy_token` varchar(100) DEFAULT NULL,
    `buy_symbol` varchar(100) DEFAULT NULL,
    `buy_amount` varchar(100) DEFAULT NULL,
    `buy_price` decimal(22,18) DEFAULT NULL,
    `sell_token` varchar(100) DEFAULT NULL,
    `sell_symbol` varchar(100) DEFAULT NULL,
    `sell_amount` varchar(100) DEFAULT NULL,
    `sell_price` decimal(22,18) DEFAULT NULL,
    `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` timestamp NULL DEFAULT NULL,
    `decimals` int DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB;